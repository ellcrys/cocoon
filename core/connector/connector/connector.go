package connector

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"

	"time"

	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/ellcrys/crypto"
	"github.com/ellcrys/util"
	"github.com/goware/urlx"
	cutil "github.com/ncodes/cocoon-util"
	"github.com/ncodes/cocoon/core/api/api"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/connector/monitor"
	"github.com/ncodes/cocoon/core/orderer/orderer"
	"github.com/ncodes/cocoon/core/types"
	docker "github.com/ncodes/go-dockerclient"
	logging "github.com/op/go-logging"
	context "golang.org/x/net/context"
)

var log *logging.Logger
var buildLog *logging.Logger
var runLog *logging.Logger
var configLog *logging.Logger
var ccodeLog *logging.Logger
var fetchLog *logging.Logger
var logHealthChecker *logging.Logger
var dckClient *docker.Client

var bridgeName = os.Getenv("BRIDGE_NAME")

func init() {
}

// Connector defines a structure for starting and managing a cocoon (coode)
type Connector struct {
	waitCh            chan bool
	req               *Request
	connectorRPCAddr  string
	cocoonCodeRPCAddr string
	languages         []Language
	container         *docker.APIContainers
	containerRunning  bool
	monitor           *monitor.Monitor
	healthCheck       *HealthChecker
	ordererDiscovery  *orderer.Discovery
	cocoon            *types.Cocoon
}

// NewConnector creates a new connector
func NewConnector(req *Request, waitCh chan bool) *Connector {
	return &Connector{
		req:              req,
		waitCh:           waitCh,
		monitor:          monitor.NewMonitor(req.ID),
		ordererDiscovery: orderer.NewDiscovery(),
	}
}

// Launch starts a cocoon code
func (cn *Connector) Launch(connectorRPCAddr, cocoonCodeRPCAddr string) {

	log = config.MakeLogger("connector", fmt.Sprintf("cocoon.%s", cn.req.ID))
	buildLog = config.MakeLogger("ccode.build", fmt.Sprintf("cocoon.%s", cn.req.ID))
	runLog = config.MakeLoggerMessageOnly("ccode.run", fmt.Sprintf("cocoon.%s", cn.req.ID))
	configLog = config.MakeLogger("ccode.config", fmt.Sprintf("cocoon.%s", cn.req.ID))
	ccodeLog = config.MakeLogger("ccode.main", fmt.Sprintf("cocoon.%s", cn.req.ID))
	fetchLog = config.MakeLogger("ccode.fetch", fmt.Sprintf("cocoon.%s", cn.req.ID))
	logHealthChecker = config.MakeLogger("connector.health_checker", fmt.Sprintf("cocoon.%s", cn.req.ID))

	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	if err != nil {
		log.Errorf("failed to create docker client. Is dockerd running locally?. %s", err)
		cn.Stop(true)
		return
	}

	dckClient = client
	cn.monitor.SetDockerClient(dckClient)
	cn.healthCheck = NewHealthChecker(cn.cocoonCodeRPCAddr, cn.cocoonUnresponsive)
	go cn.ordererDiscovery.Discover()

	// No need downloading, building and starting a cocoon code
	// if DEV_COCOON_RPC_ADDR has been specified. This means a dev cocoon code
	// is running at that address.
	if devCocoonCodeRPCAddr := os.Getenv("DEV_COCOON_RPC_ADDR"); len(devCocoonCodeRPCAddr) > 0 {
		cn.cocoonCodeRPCAddr = devCocoonCodeRPCAddr
		log.Infof("[Dev] Will interact with cocoon code at %s", devCocoonCodeRPCAddr)
		cn.healthCheck.Start()
		return
	}

	log.Info("Ready to install cocoon code")
	log.Debugf("Found ccode url=%s and lang=%s", cn.req.URL, cn.req.Lang)

	lang := cn.GetLanguage(cn.req.Lang)
	if lang == nil {
		log.Errorf("cocoon code language (%s) not supported", cn.req.Lang)
		cn.Stop(true)
		return
	}

	cocoonCodeContainer, err := cn.prepareContainer(lang)
	if err != nil {
		log.Error(err.Error())
		cn.Stop(true)
		return
	}

	lang.SetRunEnv(map[string]string{
		"COCOON_ID":           cn.req.ID,
		"CONNECTOR_RPC_ADDR":  cn.connectorRPCAddr,
		"COCOON_RPC_ADDR":     cn.cocoonCodeRPCAddr, // cocoon code server will bind to the port of this address
		"COCOON_LINK":         cn.req.Link,          // the cocoon code id to link to natively
		"COCOON_CODE_VERSION": cn.req.Version,
	})

	go cn.monitor.Monitor()

	go func() {
		if err = cn.run(cocoonCodeContainer, lang); err != nil {
			log.Error(err.Error())
			cn.Stop(true)
			return
		}
	}()
}

// GetOrdererDiscoverer returns the orderer discovery instance
func (cn *Connector) GetOrdererDiscoverer() *orderer.Discovery {
	return cn.ordererDiscovery
}

// cocoonUnresponsive is called when the cocoon code failed health check
func (cn *Connector) cocoonUnresponsive() {
	log.Info("Cocoon code has failed health check. Stopping cocoon code.")
	cn.Stop(true)
}

// SetAddrs sets the address of the connector and cocoon code RPC servers
func (cn *Connector) SetAddrs(connectorRPCAddr, cocoonCodeRPCAddr string) {
	cn.connectorRPCAddr = connectorRPCAddr
	cn.cocoonCodeRPCAddr = cocoonCodeRPCAddr
}

// GetRequest returns the current cocoon launch request
func (cn *Connector) GetRequest() *Request {
	return cn.req
}

// GetCocoonCodeRPCAddr returns the RPC address of the cocoon code
func (cn *Connector) GetCocoonCodeRPCAddr() string {
	return cn.cocoonCodeRPCAddr
}

// prepareContainer fetches the cocoon code source,
// moves the source in to the running cocoon code container,
// builds the source within the container (if required)
// and configures default firewall.
func (cn *Connector) prepareContainer(lang Language) (*docker.APIContainers, error) {

	var containerName = util.Env("COCOON_CONTAINER_NAME", "")
	if len(containerName) == 0 {
		return nil, fmt.Errorf("container name is unknown")
	}

	// ensure cocoon code isn't already launched on a container
	container, err := cn.getContainer(containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to check whether cocoon code is already active. %s ", err.Error())
	} else if container == nil {
		return nil, fmt.Errorf("cocoon code container has not been started")
	}

	// set flag to true to indicate that the container is running
	cn.containerRunning = true
	cn.container = container
	cn.monitor.SetContainerID(cn.container.ID)
	cn.HookToMonitor(cn.req)

	err = cn.fetchSource(lang)
	if err != nil {
		return nil, err
	}

	if lang.RequiresBuild() {
		var buildParams map[string]interface{}

		if len(cn.req.BuildParams) == 0 {
			log.Info("No build parameters provided. Making binary.")
		} else {
			log.Info("Parsing build parameters")
			cn.req.BuildParams, err = crypto.FromBase64(cn.req.BuildParams)
			if err != nil {
				return nil, fmt.Errorf("failed to decode build parameter. Expects a base 64 encoded string. %s", err)
			}
			if err = util.FromJSON([]byte(cn.req.BuildParams), &buildParams); err != nil {
				return nil, fmt.Errorf("failed to parse build parameter. Expects valid json string. %s", err)
			}
			if err = lang.SetBuildParams(buildParams); err != nil {
				return nil, fmt.Errorf("failed to set and validate build parameter. %s", err)
			}
		}

		err = cn.build(container, lang)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
	}

	if err = cn.configFirewall(container, cn.req); err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return container, nil
}

// HookToMonitor is where all listeners to the monitor
// are attached.
func (cn *Connector) HookToMonitor(req *Request) {
	go func() {
		for evt := range cn.monitor.GetEmitter().On("monitor.report") {
			if cn.RestartIfDiskAllocExceeded(req, evt.Args[0].(monitor.Report).DiskUsage) {
				break
			}
		}
	}()
}

// setStatus Set the cocoon status
func (cn *Connector) setStatus(status string) error {

	ctx, cc := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cc()

	cocoon, err := cn.GetCocoon(ctx, cn.req.ID)
	if err != nil {
		return err
	}

	cocoon.Status = status
	err = cn.PutCocoon(ctx, cocoon)
	if err != nil {
		return err
	}

	return nil
}

// RestartIfDiskAllocExceeded restarts the cocoon code is disk usages
// has exceeded its set limit.
func (cn *Connector) RestartIfDiskAllocExceeded(req *Request, curDiskSize int64) bool {
	if curDiskSize > req.DiskLimit {
		log.Errorf("cocoon code has used more than its allocated disk space (%s of %s)",
			humanize.Bytes(uint64(curDiskSize)),
			humanize.Bytes(uint64(req.DiskLimit)))
		if err := cn.restart(); err != nil {
			log.Error(err.Error())
			return false
		}
		return true
	}
	return false
}

// restart restarts the cocoon code. The running cocoon code is stopped
// and relaunched.
func (cn *Connector) restart() error {

	if dckClient == nil || cn.container == nil {
		return nil
	}

	if cn.monitor != nil {
		cn.monitor.Reset()
	}

	log.Info("Restarting cocoon code")

	cn.containerRunning = false

	err := dckClient.RemoveContainer(docker.RemoveContainerOptions{
		ID:            cn.container.ID,
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil {
		return fmt.Errorf("failed to remove container. %s", err)
	}

	cocoonCodeContainer, err := cn.prepareContainer(cn.GetLanguage(cn.req.Lang))
	if err != nil {
		return fmt.Errorf("restart failed: %s", err)
	}

	go cn.monitor.Monitor()

	go func() {
		if err = cn.run(cocoonCodeContainer, cn.GetLanguage(cn.req.Lang)); err != nil {
			log.Errorf("restart failed: %s", err)
			cn.Stop(true)
		}
	}()

	return nil
}

// AddLanguage adds a new langauge to the launcher.
// Will return error if language is already added
func (cn *Connector) AddLanguage(lang Language) error {
	if cn.GetLanguage(lang.GetName()) != nil {
		return fmt.Errorf("language already exist")
	}
	cn.languages = append(cn.languages, lang)
	return nil
}

// GetLanguage will return a langauges or nil if not found
func (cn *Connector) GetLanguage(name string) Language {
	for _, l := range cn.languages {
		if l.GetName() == name {
			return l
		}
	}
	return nil
}

// GetLanguages returns all languages added to the launcher
func (cn *Connector) GetLanguages() []Language {
	return cn.languages
}

// fetchSource fetches the cocoon code source from
// a remote address
func (cn *Connector) fetchSource(lang Language) error {

	if !cutil.IsGithubRepoURL(cn.req.URL) {
		return fmt.Errorf("only public source code hosted on github is supported") // TODO: support zip files
	}

	return cn.fetchFromGit(lang)
}

// fetchFromGit fetches cocoon code from git repo
// and returns the download directory path.
func (cn *Connector) fetchFromGit(lang Language) error {

	var repoTarURL, downloadDst string
	var err error

	// set version to latest if not provided
	versionStr := cn.req.Version
	if versionStr == "" {
		versionStr = "latest"
	}

	// If version is a sha1 hash, it is a commit id, fetch the repo using this id
	// otherwise it is considered a release tag. If version is not set, we fetch the latest release
	if cutil.IsGithubCommitID(cn.req.Version) {
		url, err := urlx.Parse(cn.req.URL)
		if err != nil {
			return fmt.Errorf("Failed to parse git url: %s", err)
		}
		repoTarURL = fmt.Sprintf("https://api.github.com/repos%s/tarball/%s", url.Path, cn.req.Version)
		log.Debugf("Downloading repo with commit id = %s", versionStr)
	} else {
		repoTarURL, err = cutil.GetGithubRepoRelease(cn.req.URL, cn.req.Version)
		if err != nil {
			return fmt.Errorf("Failed to fetch release from github repo. %s", err)
		}
		if len(cn.req.Version) == 0 {
			log.Debug("Downloading latest repo")
		} else {
			log.Debugf("Downloading repo with release tag = %s", cn.req.Version)
		}
	}

	// determine download directory
	downloadDst = lang.GetDownloadDestination()
	log.Infof("Downloading cocoon repository with version=%s", versionStr)

	// construct fetch script
	filePath := path.Join(downloadDst, fmt.Sprintf("%s.tar.gz", cn.req.ID))
	fetchScript := `
	    iptables -F &&
		rm -rf ` + downloadDst + ` &&			
		mkdir -p ` + downloadDst + ` &&
		printf "Downloading source from remote url to download destination\n" &&
		wget ` + repoTarURL + ` -O ` + filePath + ` &> /dev/null &&
		printf "Unpacking downloaded source \n" &&
		tar -xvf ` + filePath + ` -C ` + downloadDst + ` --strip-components 1 &> /dev/null &&
		rm -rf ` + filePath + ` &&
		printf "Creating source root directory\n" &&
		mkdir -p ` + lang.GetSourceRootDir() + ` &&
		printf "Moving source to new source root directory\n" &&
		mv ` + downloadDst + `/* ` + lang.GetSourceRootDir() + `
	`

	cmd := []string{"bash", "-c", strings.TrimSpace(fetchScript)}
	return cn.execInContainer(cn.container, "FETCH", cmd, true, fetchLog, func(state string, exitCode interface{}) error {
		switch state {
		case "end":
			if exitCode.(int) == 0 {
				fetchLog.Info("Fetch succeeded!")
				fetchLog.Infof("Successfully unpacked cocoon code")
			} else {
				return fmt.Errorf("Fetch has failed with exit code=%d", exitCode.(int))
			}
		}
		return nil
	})
}

// getContainer returns a container with a
// matching name or nil if not found.
func (cn *Connector) getContainer(name string) (*docker.APIContainers, error) {
	apiContainers, err := dckClient.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return nil, err
	}

	for _, c := range apiContainers {
		if util.InStringSlice(c.Names, "/"+name) {
			return &c, nil
		}
	}

	return nil, nil
}

// copySourceToContainer copies the downloaded cocoon code source into the cocoon code container
func (cn *Connector) copySourceToContainer(lang Language, container *docker.APIContainers) error {

	// no matter what happens, remove download directory
	defer func() {
		os.RemoveAll(lang.GetDownloadDestination())
		log.Info("Removed download directory")
	}()

	// create source root directory
	cmd := []string{"bash", "-c", "mkdir -p " + lang.GetSourceRootDir()}
	cn.execInContainer(container, "CREATE_SOURCE_DIR", cmd, true, buildLog, func(state string, exitCode interface{}) error {
		if state == "end" && exitCode.(int) != 0 {
			return fmt.Errorf("failed to create source directory")
		}
		return nil
	})

	// copy source directory to the container's source directory
	log.Debugf("Copying source from %s to container:%s", lang.GetDownloadDestination(), lang.GetCopyDestination())
	args := []string{"cp", lang.GetDownloadDestination(), fmt.Sprintf("%s:%s", container.ID, lang.GetCopyDestination())}
	if err := exec.Command("docker", args...).Run(); err != nil {
		return fmt.Errorf("failed to copy cocoon code source to cocoon. %s", err)
	}

	log.Info("Copied cocoon code source to cocoon")

	return nil
}

// stopContainer stop container. Kill it if it doesn't
// end after 5 seconds.
func (cn *Connector) stopContainer(id string) error {
	if err := dckClient.StopContainer(id, uint((5 * time.Second).Seconds())); err != nil {
		return err
	}
	cn.containerRunning = false
	return nil
}

// Executes is a general purpose function
// to execute a command in a running container. If container is not running, it starts it.
// It accepts the container, a unique name for the execution
// and a callback function that is passed a lifecycle status and a value.
// If privileged is set to true, command will attain root powers.
// Supported statuses are before (before command is executed), after (after command is executed)
// and end (when command exits).
func (cn *Connector) execInContainer(container *docker.APIContainers, name string, command []string, privileged bool, logger *logging.Logger, cb func(string, interface{}) error) error {

	containerStatus, err := dckClient.InspectContainer(container.ID)
	if err != nil {
		return fmt.Errorf("failed to inspect container before executing command [%s]. %s", name, err)
	}

	if !containerStatus.State.Running {
		err := dckClient.StartContainer(container.ID, nil)
		if err != nil {
			return fmt.Errorf("failed start container for exec [%s]. %s", name, err.Error())
		}
		cn.containerRunning = true
	}

	exec, err := dckClient.CreateExec(docker.CreateExecOptions{
		Container:    container.ID,
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          command,
		Privileged:   privileged,
	})

	if err != nil {
		return fmt.Errorf("failed to create exec [%s] object. %s", name, err)
	}

	if err = cb("before", nil); err != nil {
		return err
	}

	outStream := NewLogStreamer()
	outStream.SetLogger(logger)

	go func() {
		err = dckClient.StartExec(exec.ID, docker.StartExecOptions{
			OutputStream: outStream.GetWriter(),
			ErrorStream:  outStream.GetWriter(),
		})
		if err != nil {
			log.Infof("failed to start exec [%s] command. %s", name, err)
		}
	}()

	go func() {
		err := outStream.Start()
		if err != nil {
			log.Errorf("failed to start exec [%s] output stream logger. %s", name, err)
		}
	}()

	execExitCode := 0
	time.Sleep(1 * time.Second)

	if err = cb("after", nil); err != nil {
		outStream.Stop()
		return err
	}

	for cn.containerRunning {
		execIns, err := dckClient.InspectExec(exec.ID)
		if err != nil {
			outStream.Stop()
			return err
		}

		if execIns.Running {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		execExitCode = execIns.ExitCode
		break
	}

	outStream.Stop()

	if err = cb("end", execExitCode); err != nil {
		return err
	}

	if execExitCode != 0 {
		return fmt.Errorf("Exec [%s] exited with code=%d", name, execExitCode)
	}

	return nil
}

// build starts up the container and builds the cocoon code
// according to the build script provided by the language.
func (cn *Connector) build(container *docker.APIContainers, lang Language) error {
	cmd := []string{"bash", "-c", lang.GetBuildScript()}
	return cn.execInContainer(container, "BUILD", cmd, false, buildLog, func(state string, exitCode interface{}) error {
		switch state {
		case "before":
			log.Info("Building cocoon code")
			cn.setStatus(api.CocoonStatusBuilding)
		case "end":
			if exitCode.(int) == 0 {
				log.Info("Build succeeded!")
			} else {
				return fmt.Errorf("Build has failed with exit code=%d", exitCode.(int))
			}
		}
		return nil
	})
}

// Run the cocoon code. First it gets the IP address of the container and sets
// the language environment.
func (cn *Connector) run(container *docker.APIContainers, lang Language) error {
	return cn.execInContainer(container, "RUN", lang.GetRunScript(), false, runLog, func(state string, exitCode interface{}) error {
		switch state {
		case "before":
			log.Info("Starting cocoon code")
		case "after":
			go cn.healthCheck.Start()
			cn.setStatus(api.CocoonStatusRunning)
			return nil
		case "end":
			cn.setStatus(api.CocoonStatusStopped)
			if exitCode.(int) == 0 {
				log.Info("Cocoon code successfully stop")
				return nil
			}
		}
		return nil
	})
}

// getDefaultFirewall returns the default firewall rules
// for a cocoon container.
func (cn *Connector) getDefaultFirewall(cocoonFirewall types.Firewall) string {
	_, cocoonCodeRPCPort, _ := net.SplitHostPort(cn.cocoonCodeRPCAddr)
	connectorRPCIP, connectorRPCPort, _ := net.SplitHostPort(cn.connectorRPCAddr)

	var cocoonFirewallRules []string
	for _, rule := range cocoonFirewall {
		cocoonFirewallRules = append(cocoonFirewallRules, fmt.Sprintf("iptables -A OUTPUT -p %s -d %s --dport %s -j ACCEPT", rule.Protocol, rule.Destination, rule.DestinationPort))
	}

	return strings.TrimSpace(`iptables -F && 
			iptables -P INPUT DROP; 
			iptables -P FORWARD DROP; 
			iptables -P OUTPUT DROP;
			iptables -A OUTPUT -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT;
			iptables -A OUTPUT -p tcp -d ` + connectorRPCIP + ` --dport ` + connectorRPCPort + ` -j ACCEPT;
			iptables -A OUTPUT -p udp --dport 53 -j ACCEPT;
			` + strings.Join(cocoonFirewallRules, ";") + `
			iptables -A INPUT -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT;
			iptables -A INPUT -p tcp -s ` + connectorRPCIP + ` --dport ` + cocoonCodeRPCPort + ` -j ACCEPT 
		`)
}

// configFirewall configures the container firewall.
func (cn *Connector) configFirewall(container *docker.APIContainers, req *Request) error {

	// get cocoon object
	ctx, cc := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cc()
	cocoon, err := cn.GetCocoon(ctx, cn.req.ID)
	if err != nil {
		return err
	}

	cmd := []string{"bash", "-c", cn.getDefaultFirewall(cocoon.Firewall)}
	return cn.execInContainer(container, "CONFIG-FIREWALL", cmd, true, configLog, func(state string, exitCode interface{}) error {
		switch state {
		case "before":
			log.Info("Configuring firewall for cocoon")
		case "end":
			if exitCode.(int) == 0 {
				log.Info("Firewall configured for cocoon")
			}
		}
		return nil
	})
}

// Stop stops all sub routines and releases resources.
func (cn *Connector) Stop(failed bool) error {

	defer func() {
		cn.setStatus(api.CocoonStatusStopped)
		cn.containerRunning = false
		cn.waitCh <- failed
	}()

	if cn.monitor != nil {
		cn.monitor.Stop()
	}

	if cn.healthCheck != nil {
		cn.healthCheck.Stop()
	}

	return nil
}
