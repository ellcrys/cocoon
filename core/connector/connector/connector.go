package connector

import (
	"fmt"
	"net"
	"os"
	"path"

	"time"

	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/ellcrys/util"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/goware/urlx"
	cutil "github.com/ncodes/cocoon-util"
	"github.com/ncodes/cocoon/core/api/api"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/connector/connector/languages"
	"github.com/ncodes/cocoon/core/connector/monitor"
	"github.com/ncodes/cocoon/core/connector/router"
	"github.com/ncodes/cocoon/core/orderer/orderer"
	"github.com/ncodes/cocoon/core/platform"
	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/modo"
	logging "github.com/op/go-logging"
	"github.com/pkg/errors"
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
	log = config.MakeLogger("connector")
	buildLog = config.MakeLogger("ccode.build")
	runLog = config.MakeLoggerMessageOnly("ccode.run")
	configLog = config.MakeLogger("ccode.config")
	ccodeLog = config.MakeLogger("ccode.main")
	fetchLog = config.MakeLogger("ccode.fetch")
	logHealthChecker = config.MakeLogger("connector.health_checker")
}

// Connector defines a structure for starting and managing a cocoon (ccode)
type Connector struct {
	waitCh            chan bool
	spec              *types.Spec
	connectorRPCAddr  string
	cocoonCodeRPCAddr string
	languages         []languages.Language
	container         *docker.APIContainers
	cocoonRunning     bool
	routerHelper      *router.Helper
	monitor           *monitor.Monitor
	healthCheck       *HealthChecker
	Platform          *platform.Platform
	lang              languages.Language
}

// NewConnector creates a new connector
func NewConnector(platform *platform.Platform, spec *types.Spec, waitCh chan bool) *Connector {
	return &Connector{
		spec:     spec,
		waitCh:   waitCh,
		monitor:  monitor.NewMonitor(spec.ID),
		Platform: platform,
	}
}

// Launch starts a cocoon code
func (cn *Connector) Launch(connectorRPCAddr, cocoonCodeRPCAddr string) {

	endpoint := "unix:///var/run/docker.sock"
	dckClient, err := docker.NewClient(endpoint)
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "failed to create docker client. Is dockerd running locally?"))
		cn.Stop(true)
		return
	}

	cn.monitor.SetDockerClient(dckClient)
	cn.healthCheck = NewHealthChecker(cn.cocoonCodeRPCAddr, cn.onCodeHealthCheckFail)

	// No need downloading, building and starting a cocoon code
	// if DEV_ADDR_COCOON_CODE_RPC has been specified. This means a dev cocoon code
	// is running at that address.
	if devCocoonCodeRPCAddr := os.Getenv("DEV_ADDR_COCOON_CODE_RPC"); len(devCocoonCodeRPCAddr) > 0 {
		cn.cocoonCodeRPCAddr = devCocoonCodeRPCAddr
		log.Infof("[Dev] Will interact with cocoon code at %s", devCocoonCodeRPCAddr)
		cn.healthCheck.Start()
		fmt.Println("failed")
		return
	}

	log.Info("Ready to install cocoon code")
	log.Debugf("Found ccode url=%s and lang=%s", cn.spec.URL, cn.spec.Lang)

	cn.lang = cn.GetLanguage(cn.spec.Lang)
	if cn.lang == nil {
		log.Errorf("cocoon code language (%s) not supported", cn.spec.Lang)
		cn.Stop(true)
		return
	}

	cocoonCodeContainer, err := cn.prepareContainer()
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, ""))
		cn.Stop(true)
		return
	}

	log.Info("Preparing cocoon code environment variables")

	// cocoon environment a cocoon must have
	var env = map[string]string{
		"COCOON_ID":           cn.spec.ID,
		"CONNECTOR_RPC_ADDR":  cn.connectorRPCAddr,
		"COCOON_RPC_ADDR":     cn.cocoonCodeRPCAddr,
		"COCOON_LINK":         cn.spec.Link,
		"COCOON_CODE_VERSION": cn.spec.Version,
	}

	// include the environment variable of the release
	cocoonEnv := cn.spec.Release.Env.ProcessAsOne(false)
	for k, v := range cocoonEnv {
		env[k] = v
	}

	cn.lang.SetRunEnv(env)
	go cn.monitor.Monitor()

	go func() {
		if err = cn.run(cocoonCodeContainer); err != nil {
			log.Errorf("%+v", errors.Wrap(err, ""))
			cn.Stop(true)
			return
		}
	}()
}

// SetRouterHelper sets the router helpder
func (cn *Connector) SetRouterHelper(rh *router.Helper) {
	cn.routerHelper = rh
}

// GetOrdererDiscoverer returns the orderer discovery instance
func (cn *Connector) GetOrdererDiscoverer() *orderer.Discovery {
	return cn.Platform.GetOrdererDiscoverer()
}

// onCodeHealthCheckFail is called when the cocoon code failed health check
func (cn *Connector) onCodeHealthCheckFail() {
	log.Info("Cocoon code has failed health check. Stopping cocoon code.")
	cn.Stop(true)
}

// SetAddrs sets the address of the connector and cocoon code RPC servers
func (cn *Connector) SetAddrs(connectorRPCAddr, cocoonCodeRPCAddr string) {
	cn.connectorRPCAddr = connectorRPCAddr
	cn.cocoonCodeRPCAddr = cocoonCodeRPCAddr
}

// GetSpec returns the current cocoon launch spec
func (cn *Connector) GetSpec() *types.Spec {
	return cn.spec
}

// GetCocoonCodeRPCAddr returns the RPC address of the cocoon code
func (cn *Connector) GetCocoonCodeRPCAddr() string {
	return cn.cocoonCodeRPCAddr
}

// prepareContainer fetches the cocoon code source,
// moves the source in to the running cocoon code container,
// builds the source within the container (if required)
// and configures default firewall.
func (cn *Connector) prepareContainer() (*docker.APIContainers, error) {

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
	cn.cocoonRunning = true
	cn.container = container
	cn.monitor.SetContainerID(cn.container.ID)
	cn.HookToMonitor()

	err = cn.fetchSource()
	if err != nil {
		return nil, err
	}

	if cn.lang.RequiresBuild() {
		var buildParams map[string]interface{}

		if len(cn.spec.BuildParams) == 0 {
			log.Info("No build parameters provided. Running build with language default config.")
		} else {
			log.Info("Parsing and validating build parameters")
			if err = util.FromJSON([]byte(cn.spec.BuildParams), &buildParams); err != nil {
				return nil, fmt.Errorf("failed to parse build parameter. Expects valid json string. %s", err)
			}
			if err = cn.lang.SetBuildParams(buildParams); err != nil {
				return nil, fmt.Errorf("failed to set and validate build parameter. %s", err)
			}
		}

		err = cn.build(container)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
	}

	if err = cn.configureFirewall(container); err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return container, nil
}

// HookToMonitor is where all listeners to the monitor
// are attached.
func (cn *Connector) HookToMonitor() {
	go func() {
		for evt := range cn.monitor.GetEmitter().On("monitor.report") {
			if cn.RestartIfDiskAllocExceeded(evt.Args[0].(monitor.Report).DiskUsage) {
				break
			}
		}
	}()
}

// setStatus Set the cocoon status
func (cn *Connector) setStatus(status string) error {

	ctx, cc := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cc()

	cocoon, err := cn.Platform.GetCocoon(ctx, cn.spec.ID)
	if err != nil {
		return err
	}

	cocoon.Status = status
	err = cn.Platform.PutCocoon(ctx, cocoon)
	if err != nil {
		return err
	}

	return nil
}

// RestartIfDiskAllocExceeded restarts the cocoon code is disk usages
// has exceeded its set limit.
func (cn *Connector) RestartIfDiskAllocExceeded(curDiskSize int64) bool {
	if curDiskSize > cn.spec.DiskLimit {
		log.Errorf("cocoon code has used more than its allocated disk space (%s of %s)",
			humanize.Bytes(uint64(curDiskSize)),
			humanize.Bytes(uint64(cn.spec.DiskLimit)))
		if err := cn.restart(); err != nil {
			log.Errorf("%+v", err)
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

	cn.cocoonRunning = false

	err := dckClient.RemoveContainer(docker.RemoveContainerOptions{
		ID:            cn.container.ID,
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil {
		return fmt.Errorf("failed to remove container. %s", err)
	}

	cocoonCodeContainer, err := cn.prepareContainer()
	if err != nil {
		return fmt.Errorf("restart failed: %s", err)
	}

	go cn.monitor.Monitor()

	go func() {
		if err = cn.run(cocoonCodeContainer); err != nil {
			log.Errorf("%+v", errors.Wrap(err, "restart failed"))
			cn.Stop(true)
		}
	}()

	return nil
}

// AddLanguage adds a new langauge to the launcher.
// Will return error if language is already added
func (cn *Connector) AddLanguage(lang languages.Language) error {
	if cn.GetLanguage(lang.GetName()) != nil {
		return fmt.Errorf("language already exist")
	}
	cn.languages = append(cn.languages, lang)
	return nil
}

// GetLanguage will return a langauges or nil if not found
func (cn *Connector) GetLanguage(name string) languages.Language {
	for _, l := range cn.languages {
		if l.GetName() == name {
			return l
		}
	}
	return nil
}

// GetLanguages returns all languages added to the launcher
func (cn *Connector) GetLanguages() []languages.Language {
	return cn.languages
}

// fetchSource fetches the cocoon code source from
// a remote address
func (cn *Connector) fetchSource() error {

	if !cutil.IsGithubRepoURL(cn.spec.URL) {
		return fmt.Errorf("only public source code hosted on github is supported") // TODO: support zip files
	}

	return cn.fetchFromGit()
}

// fetchFromGit execs fetch script that fetches the cocoon code source
// into the cocoon code container and readies it form the build stage.
func (cn *Connector) fetchFromGit() error {

	var repoTarURL, downloadDst string
	var err error

	// set version to latest if not provided
	versionStr := cn.spec.Version
	if versionStr == "" {
		versionStr = "latest"
		log.Debug("Cocoon code version set to = latest")
	}

	// If version is a sha1 hash, it is a commit id, fetch the repo using this id
	// otherwise it is considered a release tag. If version is not set, we fetch the latest release
	if cutil.IsGithubCommitID(cn.spec.Version) {
		url, err := urlx.Parse(cn.spec.URL)
		if err != nil {
			return fmt.Errorf("Failed to parse git url: %s", err)
		}
		repoTarURL = fmt.Sprintf("https://api.github.com/repos%s/tarball/%s", url.Path, cn.spec.Version)
		log.Debugf("Downloading repo with commit id = %s", versionStr)
	} else {
		repoTarURL, err = cutil.GetGithubRepoRelease(cn.spec.URL, cn.spec.Version)
		if err != nil {
			return fmt.Errorf("Failed to fetch release from github repo. %s", err)
		}
		if len(cn.spec.Version) == 0 {
			log.Debug("Downloading latest repo")
		} else {
			log.Debugf("Downloading repo with release tag = %s", cn.spec.Version)
		}
	}

	// determine download directory
	downloadDst = cn.lang.GetDownloadDestination()
	log.Infof("Downloading cocoon repository with version=%s", versionStr)

	// construct fetch script
	filePath := path.Join(downloadDst, fmt.Sprintf("%s.tar.gz", cn.spec.ID))
	fetchScript := `
	    iptables -F &&
		rm -rf ` + downloadDst + ` &&	
		rm -rf ` + cn.lang.GetSourceRootDir() + ` &&		
		rm -rf ` + filePath + ` &&
		mkdir -p ` + downloadDst + ` &&
		mkdir -p ` + cn.lang.GetSourceRootDir() + ` &&
		printf "> Downloading source from remote url to download destination\n" &&
		wget ` + repoTarURL + ` -O ` + filePath + ` &> /dev/null &&
		printf "> Unpacking downloaded source \n" &&
		tar -xvf ` + filePath + ` -C ` + downloadDst + ` --strip-components 1 &> /dev/null &&
		printf "> Creating source root directory\n" &&
		printf "> Moving source to new source root directory\n" &&
		mv ` + downloadDst + `/* ` + cn.lang.GetSourceRootDir() + `
	`

	cmd := []string{"bash", "-c", strings.TrimSpace(fetchScript)}
	return cn.execInContainer(cn.container, "FETCH", cmd, true, fetchLog, func(state string, exitCode interface{}) error {
		switch state {
		case "end":
			if exitCode.(int) == 0 {
				fetchLog.Info("Fetch succeeded!")
				fetchLog.Infof("Successfully unpacked cocoon code")
			} else {
				return fmt.Errorf("Repository fetch has failed with exit code=%d", exitCode.(int))
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

// stopContainer stop container. Kill it if it doesn't
// end after 5 seconds.
func (cn *Connector) stopContainer(id string) error {
	if err := dckClient.StopContainer(id, uint((5 * time.Second).Seconds())); err != nil {
		return err
	}
	cn.cocoonRunning = false
	return nil
}

// ExecInContainer executes one of more commands in a command in series. Returns a any error for each command in their
// respective index location.
func (cn *Connector) ExecInContainer(id string, commands []*modo.Do, privileged bool, outputCB modo.OutputFunc, stateCB func(modo.State, *modo.Do)) ([]error, error) {
	doer := modo.NewMoDo(id, true, privileged, outputCB)
	doer.UseClient(dckClient)
	doer.Add(commands...)
	doer.SetStateCB(stateCB)

	done := make(chan struct{}, 1)
	var err error
	var errs []error
	go func() {
		errs, err = doer.Do()
		close(done)
	}()
	time.Sleep(3 * time.Hour)
	return errs, err
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
		cn.cocoonRunning = true
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
			log.Errorf("%+v", errors.Wrap(fmt.Errorf("failed to start exec [%s] command. %s", name, err), ""))
		}
	}()

	go func() {
		err := outStream.Start()
		if err != nil {
			log.Errorf("%+v", errors.Wrap(fmt.Errorf("failed to start exec [%s] output stream logger. %s", name, err), ""))
		}
	}()

	execExitCode := 0
	time.Sleep(1 * time.Second)

	if err = cb("after", nil); err != nil {
		outStream.Stop()
		return err
	}

	for cn.cocoonRunning {
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
func (cn *Connector) build(container *docker.APIContainers) error {
	cmd := []string{"bash", "-c", cn.lang.GetBuildScript()}
	return cn.execInContainer(container, "BUILD", cmd, false, buildLog, func(state string, exitCode interface{}) error {
		if cn.cocoonRunning {
			switch state {
			case "before":
				cn.setStatus(api.CocoonStatusBuilding)
				log.Info("Building cocoon code")
			case "end":
				if exitCode.(int) == 0 {
					log.Info("Build succeeded!")
				} else {
					return fmt.Errorf("Build has failed with exit code=%d", exitCode.(int))
				}
			}
		}
		return nil
	})
}

// configureRouter sets up frontend and backend router configurations
func (cn *Connector) configureRouter() error {

	log.Info("Configuring router rules")

	ctx, cc := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cc()
	cocoon, release, err := cn.Platform.GetCocoonAndRelease(ctx, cn.spec.ID, cn.spec.ReleaseID, false)
	if err != nil {
		return err
	}

	// if cocoon's release is not linked, then add a frontend and backend routing rule
	if len(release.Link) == 0 {
		log.Info("Cocoon is not linked to another cocoon, adding new frontend and backend")
		if err := cn.routerHelper.AddFrontend(cocoon.ID); err != nil {
			return err
		}
		if err := cn.routerHelper.AddBackend(cocoon.ID, cocoon.ID); err != nil {
			return err
		}
		log.Info("Successfully configured router")
		return nil
	}

	// cocoon's release is linked and therefore must become a backend server to the linked cocoon
	log.Infof("Cocoon is linked to another cocoon (%s), adding http server to linked cocoon backend", release.Link)
	if err = cn.routerHelper.AddBackend(release.Link, cocoon.ID); err != nil {
		return err
	}
	log.Info("Successfully configured router")
	return nil
}

// Run the cocoon code. First it gets the IP address of the container and sets
// the language environment.
func (cn *Connector) run(container *docker.APIContainers) error {
	errs, err := cn.ExecInContainer(container.ID, []*modo.Do{cn.lang.GetRunCommand()}, false, func(d []byte, stdout bool) {
		runLog.Info(string(d))
	}, func(state modo.State, task *modo.Do) {
		switch state {
		case modo.Before:
			log.Info("Starting cocoon code")
		case modo.Executing:
			time.AfterFunc(2*time.Second, func() {
				cn.healthCheck.Start()
			})
			cn.setStatus(api.CocoonStatusRunning)
			if err := cn.configureRouter(); err != nil {
				log.Errorf("Failed to set frontend router configuration: %s", err)
				cn.Stop(true)
			}
		}
	})
	if err != nil {
		return errors.Wrap(err, "failed to run cocoon code")
	} else if len(errs) > 0 {
		return errors.Wrap(errs[0], "failed to run command has failed")
	}
	return nil
}

// getFirewallScript returns the firewall script to apply to the container.
func (cn *Connector) getFirewallScript(cocoonFirewall types.Firewall) string {
	_, cocoonCodeRPCPort, _ := net.SplitHostPort(cn.cocoonCodeRPCAddr)
	connectorRPCIP, connectorRPCPort, _ := net.SplitHostPort(cn.connectorRPCAddr)

	var cocoonFirewallRules []string
	for _, rule := range cocoonFirewall {
		ipTableVals := []string{}
		if len(rule.Destination) > 0 {
			ipTableVals = append(ipTableVals, fmt.Sprintf("-p %s", rule.Protocol))
			ipTableVals = append(ipTableVals, fmt.Sprintf("-d %s", rule.Destination))
			if len(rule.DestinationPort) > 0 {
				ipTableVals = append(ipTableVals, fmt.Sprintf("--dport %s", rule.DestinationPort))
			}
			cocoonFirewallRules = append(cocoonFirewallRules, fmt.Sprintf("iptables -A OUTPUT %s -j ACCEPT", strings.Join(ipTableVals, " ")))
		}
	}

	a := strings.TrimSpace(`iptables -F && 
			iptables -P INPUT DROP; 
			iptables -P FORWARD DROP; 
			iptables -P OUTPUT DROP;
			iptables -A OUTPUT -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT;
			iptables -A OUTPUT -p tcp -d ` + connectorRPCIP + ` --dport ` + connectorRPCPort + ` -j ACCEPT;
			iptables -A OUTPUT -p udp --dport 53 -j ACCEPT;
			` + strings.Join(cocoonFirewallRules, ";") + `
			iptables -A INPUT -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT;
			iptables -A INPUT -p tcp --dport ` + cocoonCodeRPCPort + ` -j ACCEPT 
		`)

	return a
}

// configureFirewall configures the container firewall.
func (cn *Connector) configureFirewall(container *docker.APIContainers) error {

	ctx, cc := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cc()
	_, release, err := cn.Platform.GetCocoonAndRelease(ctx, cn.spec.ID, cn.spec.ReleaseID, false)
	if err != nil {
		return err
	}

	cmd := []string{"bash", "-c", cn.getFirewallScript(release.Firewall)}
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
		cn.cocoonRunning = false
		cn.setStatus(api.CocoonStatusStopped)
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
