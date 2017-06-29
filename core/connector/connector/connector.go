package connector

import (
	"fmt"
	"net"
	"os"
	"path"

	"time"

	"strings"

	"os/exec"

	humanize "github.com/dustin/go-humanize"
	cutil "github.com/ellcrys/cocoon-util"
	"github.com/ellcrys/cocoon/core/api/api"
	"github.com/ellcrys/cocoon/core/api/archiver"
	"github.com/ellcrys/cocoon/core/config"
	"github.com/ellcrys/cocoon/core/connector/connector/languages"
	"github.com/ellcrys/cocoon/core/connector/monitor"
	"github.com/ellcrys/cocoon/core/connector/router"
	"github.com/ellcrys/cocoon/core/orderer/orderer"
	"github.com/ellcrys/cocoon/core/platform"
	"github.com/ellcrys/cocoon/core/types"
	"github.com/ellcrys/util"
	docker "github.com/fsouza/go-dockerclient"
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

// DockerEndpoint is the endpoint to docker API
var DockerEndpoint = "unix:///var/run/docker.sock"

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
	resourceUsage     *monitor.Report
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

	var err error

	dckClient, err = docker.NewClient(DockerEndpoint)
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

	// when DEV_SKIP_PREPARE is set, no need to prepare the container, download source
	// code from archive etc. Simply get the container and more on. This is intend for use
	// in development environment to save network bandwidth and time.
	if os.Getenv("DEV_SKIP_PREPARE") != "true" {
		cn.container, err = cn.prepareContainer()
		if err != nil {
			log.Errorf("%+v", errors.Wrap(err, ""))
			cn.Stop(true)
			return
		}
	} else { // development mode
		// get the existing, prepared container
		containerName := util.Env("COCOON_CONTAINER_NAME", "")
		cn.container, err = cn.getContainer(containerName)
		if err != nil {
			log.Errorf("failed to check whether cocoon code is already active. %s ", err.Error())
			cn.Stop(true)
			return
		} else if cn.container == nil {
			log.Errorf("cocoon code container has not been started")
			cn.Stop(true)
			return
		}

		// clean container
		if err := cn.devCleanFirewallAndCCode(); err != nil {
			log.Errorf(err.Error())
			cn.Stop(true)
			return
		}
	}

	log.Info("Preparing cocoon code environment variables")

	// cocoon environment a cocoon must have
	var env = map[string]string{
		"COCOON_ID":           cn.spec.ID,
		"CONNECTOR_RPC_ADDR":  cn.connectorRPCAddr,
		"COCOON_RPC_ADDR":     cn.cocoonCodeRPCAddr,
		"COCOON_LINK":         cn.spec.Link,
		"COCOON_CODE_VERSION": cn.spec.Version,
		"SOURCE_DIR":          cn.lang.GetSourceDir(),
	}

	// include the environment variable of the release
	cocoonEnv := cn.spec.Release.Env.ProcessAsOne(false)
	for k, v := range cocoonEnv {
		env[k] = v
	}

	cn.lang.SetRunEnv(env)
	go cn.monitor.Monitor()

	go func() {
		if err = cn.run(cn.container); err != nil {
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
		return nil, fmt.Errorf("container name is not set")
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

	// clean container
	if err = cn.cleanContainer(); err != nil {
		return nil, errors.Wrap(err, "failed to complete clean stage")
	}

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

// HookToMonitor is where all listeners to the monitor are attached.
func (cn *Connector) HookToMonitor() {
	cn.monitor.GetEmitter().On("monitor.report", func(report monitor.Report) {

		if cn.resourceUsage == nil {
			cn.resourceUsage = &report
		}

		cn.updateResourceUsage(report)

		// save network usage
		// ctx, cc := context.WithTimeout(context.Background(), 1*time.Second)
		// defer cc()
		// totalInbound, totalOutbound, err := cn.persistNetUsage(ctx)
		// if err != nil {
		// 	log.Errorf("%+v", err)
		// }

		// // shutdown cocoon code if hard limit is exceeded (TODO: we would instead prevent any further outbound or inbound traffic)
		// if (totalInbound + totalOutbound) >= 5000000000 {
		// 	log.Errorf("Total bandwidth used has reached the max limit of 5GB")
		// 	cn.shutdown()
		// 	return
		// }

		// log.Debugf("Rx Bytes: %d / Tx Bytes: %d", report.NetRx, report.NetTx)
		cn.RestartIfDiskAllocExceeded(report.DiskUsage)
	})
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

// fetchSource fetches the cocoon code release source code
func (cn *Connector) fetchSource() error {

	if !cutil.IsGithubRepoURL(cn.spec.URL) {
		return fmt.Errorf("only public source code hosted on github is supported") // TODO: support zip files
	}

	return cn.fetchGitSourceFromArchive()
}

// devCleanFirewallAndCCode resets the firewall and kills running cocoon code.
// Intended for use in development mode where full container clean operation is not
// desirable.
func (cn *Connector) devCleanFirewallAndCCode() error {

	cmds := []*modo.Do{
		&modo.Do{Cmd: []string{"bash", "-c", "iptables -F; iptables -P INPUT ACCEPT; iptables -P OUTPUT ACCEPT; iptables -P FORWARD ACCEPT"}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", `killall -3 ccode 2>/dev/null || true 2>/dev/null`}, AbortSeriesOnFail: false},
	}

	var errCount = 0
	errs, err := cn.ExecInContainer(cn.container.ID, cmds, true, func(d []byte, stdout bool) {
		fetchLog.Info(string(d))
	}, func(state modo.State, task *modo.Do) {
		switch state {
		case modo.Begin:
			log.Infof("Cleaning cocoon code container [dev mode]")
		case modo.After:
			if task.ExitCode != 0 {
				errCount++
			}
		case modo.End:
			if errCount == 0 {
				fetchLog.Info("Cleaning complete. Cocoon code container is squeaky clean! [dev mode]")
			}
		}
	})
	if err != nil {
		return errors.Wrap(err, "failed to clean cocoon code container [dev mode]")
	} else if len(errs) > 0 {
		return errors.Wrap(errs[0], "[clean]")
	}

	return nil
}

// cleanContainer removes build directories, previously
// running cocoon code by force and resets iptables rules.
func (cn *Connector) cleanContainer() error {

	// build directories and path
	downloadDst := cn.lang.GetDownloadDir()
	filePath := path.Join(downloadDst, fmt.Sprintf("%s.tar.gz", cn.spec.ID))

	cmds := []*modo.Do{
		&modo.Do{Cmd: []string{"bash", "-c", "iptables -F; iptables -P INPUT ACCEPT; iptables -P OUTPUT ACCEPT; iptables -P FORWARD ACCEPT"}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", `rm -rf ` + downloadDst + ``}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", `rm -rf ` + cn.lang.GetSourceDir() + ``}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", `rm -rf ` + filePath + ``}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", `rm -rf ` + path.Join(os.Getenv("SHARED_DIR"), "/*") + ``}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", `killall -3 ccode 2>/dev/null || true 2>/dev/null`}, AbortSeriesOnFail: false},
	}

	var errCount = 0
	errs, err := cn.ExecInContainer(cn.container.ID, cmds, true, func(d []byte, stdout bool) {
		fetchLog.Info(string(d))
	}, func(state modo.State, task *modo.Do) {
		switch state {
		case modo.Begin:
			log.Infof("Cleaning cocoon code container")
		case modo.After:
			if task.ExitCode != 0 {
				errCount++
			}
		case modo.End:
			if errCount == 0 {
				fetchLog.Info("Cleaning complete. Cocoon code container is squeaky clean!")
			}
		}
	})
	if err != nil {
		return errors.Wrap(err, "failed to clean cocoon code container")
	} else if len(errs) > 0 {
		return errors.Wrap(errs[0], "[clean]")
	}

	return nil
}

// deleteSharedDirContents deletes the shared directory.
// The contents of shared directory will not be available to the cocoon code.
func (cn *Connector) deleteSharedDirContents() (err error) {
	if sharedDir := os.Getenv("SHARED_DIR"); sharedDir != "" {
		err = exec.Command("bash", "-c", fmt.Sprintf("rm -rf %s", path.Join(sharedDir, "/*"))).Run()
	}
	return
}

// fetchGitSourceFromArchive fetches the current cocoon release source from
// the git source archive.
func (cn *Connector) fetchGitSourceFromArchive() error {

	var repoTarURL = fmt.Sprintf(
		"https://storage.googleapis.com/%s/%s",
		util.Env("REPO_ARCHIVE_BKT", "gitsources"),
		archiver.MakeArchiveName(cn.spec.Cocoon.ID, cn.spec.Release.Version),
	)

	// determine download directory
	downloadDst := cn.lang.GetDownloadDir()
	filePath := path.Join(downloadDst, fmt.Sprintf("%s.tar.gz", cn.spec.ID))

	cmds := []*modo.Do{
		&modo.Do{Cmd: []string{"bash", "-c", fmt.Sprintf(`mkdir -p %s`, downloadDst) + ``}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", fmt.Sprintf(`mkdir -p %s`, cn.lang.GetSourceDir()) + ``}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", fmt.Sprintf(`wget %s -O %s &> /dev/null`, repoTarURL, filePath)}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", fmt.Sprintf(`tar -xvf %s -C %s --strip-components 1 &> /dev/null`, filePath, downloadDst)}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", fmt.Sprintf(`mv %s/* %s`, downloadDst, cn.lang.GetSourceDir())}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", fmt.Sprintf(`mv %s/static %s 2>/dev/null`, cn.lang.GetSourceDir(), os.Getenv("SHARED_DIR"))}, AbortSeriesOnFail: true},
	}

	var errCount = 0
	errs, err := cn.ExecInContainer(cn.container.ID, cmds, true, func(d []byte, stdout bool) {
		fetchLog.Info(string(d))
	}, func(state modo.State, task *modo.Do) {
		switch state {
		case modo.Begin:
			log.Infof("Fetching cocoon code source from archive. [version=%s]", cn.spec.Release.Version)
		case modo.After:
			if task.ExitCode != 0 {
				errCount++
			}
		case modo.End:
			if errCount == 0 {
				fetchLog.Info("Fetch succeeded!")
			}
		}
	})
	if err != nil {
		return errors.Wrap(err, "failed to fetch cocoon code source")
	} else if len(errs) > 0 {
		return errors.Wrap(errs[0], "[fetch]")
	}

	return nil
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
	return doer.Do()
}

// build starts up the container and builds the cocoon code
// according to the build script provided by the language.
func (cn *Connector) build(container *docker.APIContainers) error {
	errs, err := cn.ExecInContainer(container.ID, []*modo.Do{cn.lang.GetBuildCommand()}, false, func(d []byte, stdout bool) {
		buildLog.Info(string(d))
	}, func(state modo.State, task *modo.Do) {
		switch state {
		case modo.Before:
			cn.setStatus(api.CocoonStatusBuilding)
			log.Info("Building cocoon code")
		case modo.After:
			if task.ExitCode == 0 {
				log.Info("Build succeeded!")
			}
		}
	})
	if err != nil {
		return errors.Wrap(err, "failed to build cocoon code")
	} else if len(errs) > 0 {
		return errors.Wrap(fmt.Errorf("[build] could not be completed"), "")
	}
	return nil
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
		return errors.Wrap(errs[0], "[run]")
	}
	return nil
}

// getFirewallCommands returns the firewall script to apply to the container.
func (cn *Connector) getFirewallCommands(cocoonFirewall types.Firewall) []*modo.Do {

	_, ccodeRPCPort, _ := net.SplitHostPort(cn.cocoonCodeRPCAddr)
	conRPCIP, conRPCPort, _ := net.SplitHostPort(cn.connectorRPCAddr)

	cmds := []*modo.Do{
		&modo.Do{Cmd: []string{"bash", "-c", "iptables -F"}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", "iptables -P INPUT DROP"}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", "iptables -P FORWARD DROP"}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", "iptables -P OUTPUT DROP"}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", "iptables -A OUTPUT -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT"}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", `iptables -A OUTPUT -p tcp -d ` + conRPCIP + ` --dport ` + conRPCPort + ` -j ACCEPT`}, AbortSeriesOnFail: true},
		&modo.Do{Cmd: []string{"bash", "-c", "iptables -A OUTPUT -p udp --dport 53 -j ACCEPT"}, AbortSeriesOnFail: true},
	}

	// construct cocoon specific iptables rules
	for _, rule := range cocoonFirewall {
		flags := []string{}
		if len(rule.Destination) > 0 {
			flags = append(flags, fmt.Sprintf("-p %s", rule.Protocol))
			flags = append(flags, fmt.Sprintf("-d %s", rule.Destination))
			if len(rule.DestinationPort) > 0 {
				flags = append(flags, fmt.Sprintf("--dport %s", rule.DestinationPort))
			}
			cmds = append(cmds, &modo.Do{
				Cmd:               []string{"bash", "-c", fmt.Sprintf("iptables -A OUTPUT %s -j ACCEPT", strings.Join(flags, " "))},
				AbortSeriesOnFail: true,
			})
		}
	}

	cmds = append(cmds, &modo.Do{Cmd: []string{"bash", "-c", "iptables -A INPUT -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT"}, AbortSeriesOnFail: true})
	cmds = append(cmds, &modo.Do{Cmd: []string{"bash", "-c", `iptables -A INPUT -p tcp --dport ` + ccodeRPCPort + ` -j ACCEPT`}, AbortSeriesOnFail: true})

	return cmds
}

// configureFirewall configures the container firewall.
func (cn *Connector) configureFirewall(container *docker.APIContainers) error {
	var errCount = 0
	cmds := cn.getFirewallCommands(cn.spec.Release.Firewall)
	errs, err := cn.ExecInContainer(container.ID, cmds, true, func(d []byte, stdout bool) {
		log.Info(string(d))
	}, func(state modo.State, task *modo.Do) {
		switch state {
		case modo.Begin:
			log.Info("Configuring cocoon code firewall for cocoon")
		case modo.After:
			if task.ExitCode != 0 {
				errCount++
			}
		case modo.End:
			if errCount == 0 {
				log.Info("Firewall successfully configured")
			}
		}
	})
	if err != nil {
		return errors.Wrap(err, "failed to configure cocoon code firewall")
	} else if len(errs) > 0 {
		return errors.Wrap(errs[0], "[confire-firewall]")
	}
	return nil
}

// Stop stops all sub routines and releases resources.
func (cn *Connector) Stop(failed bool) error {

	defer func() {
		cn.cocoonRunning = false
		log.Debug("Updating cocoon status to `stopped`")
		cn.setStatus(api.CocoonStatusStopped)
		cn.waitCh <- failed
	}()

	if cn.monitor != nil {
		log.Debug("Stopping monitor")
		cn.monitor.Stop()
	}

	if cn.healthCheck != nil {
		log.Debug("Stopping health checker")
		cn.healthCheck.Stop()
	}

	log.Debug("Deleting shared directory contents")
	cn.deleteSharedDirContents()

	return nil
}

// shutdown stops the cocoon service on the scheduler.
// It is different from Stop() because this will not result in
// a restart by the scheduler.
func (cn *Connector) shutdown() error {

	cn.Stop(true)

	// ask platform to stop cocoon on the scheduler to prevent a restart
	if err := cn.Platform.GetScheduler().Stop(cn.spec.ID); err != nil {
		return errors.Wrap(err, "failed to shutdown cocoon on the scheduler")
	}

	return nil
}
