package launcher

import (
	"fmt"
	"os"
	"os/exec"

	"path"

	"time"

	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/ellcrys/crypto"
	"github.com/ellcrys/util"
	cutil "github.com/ncodes/cocoon-util"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/connector/client"
	"github.com/ncodes/cocoon/core/connector/monitor"
	docker "github.com/ncodes/go-dockerclient"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("launcher")
var buildLog = logging.MustGetLogger("ccode.build")
var runLog = logging.MustGetLogger("ccode.run")
var configLog = logging.MustGetLogger("ccode.config")
var ccodeLog = logging.MustGetLogger("ccode")
var dckClient *docker.Client

func init() {
	runLog.SetBackend(config.MessageOnlyBackend)
}

// Launcher defines cocoon code
// deployment service.
type Launcher struct {
	waitCh           chan bool
	languages        []Language
	container        *docker.Container
	client           *client.Client
	containerRunning bool
	monitor          *monitor.Monitor
	req              *Request
}

// NewLauncher creates a new launcher
func NewLauncher(waitCh chan bool) *Launcher {
	return &Launcher{
		waitCh:  waitCh,
		client:  client.NewClient(),
		monitor: monitor.NewMonitor(),
	}
}

// GetClient returns the connector to cocoon code client
func (lc *Launcher) GetClient() *client.Client {
	return lc.client
}

// Launch starts a cocoon code
func (lc *Launcher) Launch(req *Request) {

	lc.req = req

	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	if err != nil {
		log.Errorf("failed to create docker client. Is dockerd running locally?. %s", err)
		lc.Stop(true)
		return
	}

	dckClient = client
	lc.client.SetCocoonID(req.ID)
	log.Info("Connecting to: ", req.CocoonAddr)
	lc.client.SetCocoonCodeAddr("127.0.0.1" + req.CocoonAddr)
	lc.monitor.SetDockerClient(dckClient)

	// No need downloading, building and starting a cocoon code
	// if DEV_COCOON_ADDR has been specified. This means a dev cocoon code
	// is running at that address. Just start the connector's client.
	if devCocoonCodeAddr := os.Getenv("DEV_COCOON_ADDR"); len(devCocoonCodeAddr) > 0 {
		log.Infof("Connecting to dev cocoon code at %s", devCocoonCodeAddr)
		if err = lc.GetClient().Connect(); err != nil {
			log.Error(err)
			lc.Stop(true)
			return
		}
	}

	log.Info("Ready to install cocoon code")
	log.Debugf("Found ccode url=%s and lang=%s", req.URL, req.Lang)

	lang := lc.GetLanguage(req.Lang)
	if lang == nil {
		log.Errorf("cocoon code language (%s) not supported", req.Lang)
		lc.Stop(true)
		return
	}

	newContainer, err := lc.prepareContainer(req, lang)
	if err != nil {
		log.Error(err.Error())
		lc.Stop(true)
		return
	}

	go lc.monitor.Monitor()

	if err = lc.run(newContainer, lang); err != nil {
		log.Error(err.Error())
		lc.Stop(true)
		return
	}
}

// prepareContainer fetches the cocoon code source, creates a container,
// moves the source in to it, builds the source within the container (if required)
// and configures default firewall.
func (lc *Launcher) prepareContainer(req *Request, lang Language) (*docker.Container, error) {

	_, err := lc.fetchSource(req, lang)
	if err != nil {
		return nil, err
	}

	// ensure cocoon code isn't already launched on a container
	c, err := lc.getContainer(req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check whether cocoon code is already active. %s ", err.Error())
	} else if c != nil {
		return nil, fmt.Errorf("cocoon code already exists on a container")
	}

	newContainer, err := lc.createContainer(
		req.ID,
		lang,
		[]string{
			fmt.Sprintf("COCOON_ID=%s", req.ID),
			fmt.Sprintf("COCOON_ADDR=%s", req.CocoonAddr),
			fmt.Sprintf("COCOON_LINK=%s", req.Link),
		})
	if err != nil {
		return nil, fmt.Errorf("failed to create new container to run cocoon code. %s ", err.Error())
	}

	lc.container = newContainer
	lc.monitor.SetContainerID(lc.container.ID)
	lc.HookToMonitor(req)

	if lang.RequiresBuild() {
		var buildParams map[string]interface{}
		if len(req.BuildParams) > 0 {

			req.BuildParams, err = crypto.FromBase64(req.BuildParams)
			if err != nil {
				return nil, fmt.Errorf("failed to decode build parameter. Expects a base 64 encoded string. %s", err)
			}

			if err = util.FromJSON([]byte(req.BuildParams), &buildParams); err != nil {
				return nil, fmt.Errorf("failed to parse build parameter. Expects valid json string. %s", err)
			}
		}

		if err = lang.SetBuildParams(buildParams); err != nil {
			return nil, fmt.Errorf("failed to set and validate build parameter. %s", err)
		}

		err = lc.build(newContainer, lang)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}

	} else {
		log.Info("Cocoon code does not require a build processing. Skipped.")
	}

	if err = lc.configFirewall(newContainer, req); err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return newContainer, nil
}

// HookToMonitor is where all listeners to the monitor
// are attached.
func (lc *Launcher) HookToMonitor(req *Request) {
	go func() {
		for evt := range lc.monitor.GetEmitter().On("monitor.report") {
			if lc.RestartIfDiskAllocExceeded(req, evt.Args[0].(monitor.Report).DiskUsage) {
				break
			}
		}
	}()
}

// RestartIfDiskAllocExceeded restarts the cocoon code is disk usages
// has exceeded its set limit.
func (lc *Launcher) RestartIfDiskAllocExceeded(req *Request, curDiskSize int64) bool {
	if curDiskSize > req.DiskLimit {
		log.Errorf("cocoon code has used more than its allocated disk space (%s of %s)",
			humanize.Bytes(uint64(curDiskSize)),
			humanize.Bytes(uint64(req.DiskLimit)))
		if err := lc.restart(); err != nil {
			log.Error(err.Error())
			return false
		}
		return true
	}
	return false
}

// Stop closes the client, stops the container if it is still running
// and deletes the container. This will effectively bring the launcher
// to a halt. Set failed parameter to true to set a positve exit code or
// false for 0 exit code.
func (lc *Launcher) Stop(failed bool) error {

	defer func() {
		lc.waitCh <- failed
	}()

	if dckClient == nil || lc.container == nil {
		return nil
	}

	if lc.client != nil {
		lc.client.Close()
	}

	if lc.monitor != nil {
		lc.monitor.Stop()
	}

	lc.containerRunning = false

	err := dckClient.RemoveContainer(docker.RemoveContainerOptions{
		ID:            lc.container.ID,
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil {
		return fmt.Errorf("failed to remove container. %s", err)
	}

	return nil
}

// restart restarts the cocoon code. The running cocoon code is stopped
// and relaunched.
func (lc *Launcher) restart() error {

	if dckClient == nil || lc.container == nil {
		return nil
	}

	if lc.client != nil {
		lc.client.Close()
	}

	if lc.monitor != nil {
		lc.monitor.Reset()
	}

	log.Info("Restarting cocoon code")

	lc.containerRunning = false

	err := dckClient.RemoveContainer(docker.RemoveContainerOptions{
		ID:            lc.container.ID,
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil {
		return fmt.Errorf("failed to remove container. %s", err)
	}

	newContainer, err := lc.prepareContainer(lc.req, lc.GetLanguage(lc.req.Lang))
	if err != nil {
		return fmt.Errorf("restart: %s", err)
	}

	go lc.monitor.Monitor()

	go func() {
		if err = lc.run(newContainer, lc.GetLanguage(lc.req.Lang)); err != nil {
			log.Info(fmt.Errorf("restart: %s", err))
		}
	}()

	return nil
}

// AddLanguage adds a new langauge to the launcher.
// Will return error if language is already added
func (lc *Launcher) AddLanguage(lang Language) error {
	if lc.GetLanguage(lang.GetName()) != nil {
		return fmt.Errorf("language already exist")
	}
	lc.languages = append(lc.languages, lang)
	return nil
}

// GetLanguage will return a langauges or nil if not found
func (lc *Launcher) GetLanguage(name string) Language {
	for _, l := range lc.languages {
		if l.GetName() == name {
			return l
		}
	}
	return nil
}

// GetLanguages returns all languages added to the launcher
func (lc *Launcher) GetLanguages() []Language {
	return lc.languages
}

// fetchSource fetches the cocoon code source from
// a remote address
func (lc *Launcher) fetchSource(req *Request, lang Language) (string, error) {

	if !cutil.IsGithubRepoURL(req.URL) {
		return "", fmt.Errorf("only public source code hosted on github is supported") // TODO: support zip files
	}

	return lc.fetchFromGit(req, lang)
}

// findLaunch looks for a previous stored launch/Redeployment by id
// TODO: needs implementation
func (lc *Launcher) findLaunch(id string) interface{} {
	return nil
}

// fetchFromGit fetchs cocoon code from git repo.
// and returns the download directory.
func (lc *Launcher) fetchFromGit(req *Request, lang Language) (string, error) {

	var repoTarURL, downloadDst string
	var err error

	// checks if job was previously deployed. find a job by the job name.
	if lc.findLaunch(req.ID) != nil {
		return "", fmt.Errorf("cocoon code was previously launched") // TODO: fetch last launch tag and use it
	}

	repoTarURL, err = cutil.GetGithubRepoRelease(req.URL, req.Tag)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch release from github repo. %s", err)
	}

	// set tag to latest if not provided
	tagStr := req.Tag
	if tagStr == "" {
		tagStr = "latest"
	}

	// determine download directory
	downloadDst = lang.GetDownloadDestination()

	// delete download directory if it exists
	if _, err := os.Stat(downloadDst); err == nil {
		log.Info("Download destination is not empty. Deleting content")
		if err = os.RemoveAll(downloadDst); err != nil {
			return "", fmt.Errorf("failed to delete contents of download directory")
		}
		log.Info("Download directory has been deleted")
	}

	// create the download directory
	if err = os.MkdirAll(downloadDst, os.ModePerm); err != nil {
		return "", fmt.Errorf("Failed to create download directory. %s", err)
	}

	log.Infof("Downloading cocoon repository with tag=%s, dst=%s", tagStr, downloadDst)
	filePath := path.Join(downloadDst, fmt.Sprintf("%s.tar.gz", req.ID))
	err = cutil.DownloadFile(repoTarURL, filePath, func(buf []byte) {})
	if err != nil {
		return "", err
	}

	log.Info("Successfully downloaded cocoon code")
	log.Debugf("Unpacking cocoon code to %s", filePath)

	// unpack tarball
	cmd := "tar"
	args := []string{"-xf", filePath, "-C", downloadDst, "--strip-components", "1"}
	if err = exec.Command(cmd, args...).Run(); err != nil {
		return "", fmt.Errorf("Failed to unpack cocoon code repo tarball. %s", err)
	}

	log.Infof("Successfully unpacked cocoon code to %s", downloadDst)

	os.Remove(filePath)
	log.Info("Deleted the cocoon code tarball")

	return downloadDst, nil
}

// getContainer returns a container with a
// matching name or nil if not found.
func (lc *Launcher) getContainer(name string) (*docker.APIContainers, error) {
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

// createContainer creates a brand new container,
// and copies the cocoon source code to it.
func (lc *Launcher) createContainer(name string, lang Language, env []string) (*docker.Container, error) {
	container, err := dckClient.CreateContainer(docker.CreateContainerOptions{
		Name: name,
		Config: &docker.Config{
			Image:      lang.GetImage(),
			Labels:     map[string]string{"name": name, "type": "cocoon_code"},
			WorkingDir: lang.GetSourceRootDir(),
			Tty:        true,
			ExposedPorts: map[docker.Port]struct{}{
				docker.Port(fmt.Sprintf("%s/tcp", strings.Split(lc.req.CocoonAddr, ":")[1])): struct{}{},
			},
			Cmd: []string{"bash"},
			Env: env,
		},
		HostConfig: &docker.HostConfig{
			PortBindings: map[docker.Port][]docker.PortBinding{
				docker.Port(fmt.Sprintf("%s/tcp", strings.Split(lc.req.CocoonAddr, ":")[1])): []docker.PortBinding{
					docker.PortBinding{HostIP: "127.0.0.1", HostPort: strings.Split(lc.req.CocoonAddr, ":")[1]},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// copy source directory to the container's source directory
	cmd := "docker"
	args := []string{"cp", lang.GetDownloadDestination(), fmt.Sprintf("%s:%s", container.ID, lang.GetCopyDestination())}
	if err = exec.Command(cmd, args...).Run(); err != nil {
		return nil, fmt.Errorf("failed to copy cocoon code source to cocoon. %s", err)
	}

	log.Info("Copied cocoon code source to cocoon")

	return container, nil
}

// stopContainer stop container. Kill it if it doesn't
// end after 5 seconds.
func (lc *Launcher) stopContainer(id string) error {
	return dckClient.StopContainer(id, uint((5 * time.Second).Seconds()))
}

// Executes is a general purpose function
// to execute a command in a running container. If container is not running, it starts it.
// It accepts the container, a unique name for the execution
// and a callback function that is passed a lifecycle status and a value.
// If priviledged is set to true, command will attain root powers.
// Supported statuses are before (before command is executed), after (after command is executed)
// and end (when command exits).
func (lc *Launcher) execInContainer(container *docker.Container, name string, command []string, priviledged bool, logger *logging.Logger, cb func(string, interface{}) error) error {

	containerStatus, err := dckClient.InspectContainer(container.ID)
	if err != nil {
		return fmt.Errorf("failed to inspect container before executing command [%s]. %s", name, err)
	}

	if !containerStatus.State.Running {
		err := dckClient.StartContainer(container.ID, nil)
		if err != nil {
			return fmt.Errorf("failed start container for exec [%s]. %s", name, err.Error())
		}
		lc.containerRunning = true
	}

	exec, err := dckClient.CreateExec(docker.CreateExecOptions{
		Container:    container.ID,
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          command,
		Privileged:   priviledged,
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

	for lc.containerRunning {
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
// according to the build script provided by the languaged.
func (lc *Launcher) build(container *docker.Container, lang Language) error {
	cmd := []string{"bash", "-c", lang.GetBuildScript()}
	return lc.execInContainer(container, "BUILD", cmd, false, buildLog, func(state string, val interface{}) error {
		switch state {
		case "before":
			log.Info("Building cocoon code...")
		case "end":
			if val.(int) == 0 {
				log.Info("Build succeeded!")
			} else {
				return fmt.Errorf("Build has failed with exit code=%d", val.(int))
			}
		}
		return nil
	})
}

// Run the cocoon code. Start the container if it is not
// currently running (this will be true if cocoon build process was not ran).
// Also connects the client to the cocoon code
func (lc *Launcher) run(container *docker.Container, lang Language) error {
	return lc.execInContainer(container, "RUN", lang.GetRunScript(), false, runLog, func(state string, val interface{}) error {
		switch state {
		case "before":
			log.Info("Starting cocoon code")
		case "after":
			if err := lc.client.Connect(); err != nil {
				return err
			}
		case "end":
			if val.(int) == 0 {
				log.Info("Cocoon code successfully stop")
				return nil
			}
		}
		return nil
	})
}

// getDefaultFirewall returns the default firewall rules
// for a cocoon container.
func (lc *Launcher) getDefaultFirewall(ccodePort string) string {
	return strings.TrimSpace(`iptables -F && 
			iptables -P INPUT ACCEPT && 
			iptables -P FORWARD DROP &&
			iptables -P OUTPUT DROP &&
			iptables -A INPUT -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT &&
			iptables -A INPUT -p tcp --sport ` + ccodePort + ` -j ACCEPT
			iptables -A OUTPUT -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT &&
			dnsIPs="$(cat /etc/resolv.conf | grep 'nameserver' | cut -c12-)" &&
			for ip in $dnsIPs;
			do 
				iptables -A OUTPUT -m state --state NEW,ESTABLISHED -d ${ip} -p udp --dport 53 -j ACCEPT;
				iptables -A OUTPUT -m state --state ESTABLISHED -p udp -s ${ip} --sport 53 -j ACCEPT;
				iptables -A OUTPUT -m state --state NEW,ESTABLISHED -d ${ip} -p tcp --dport 53 -j ACCEPT;
				iptables -A OUTPUT -m state --state ESTABLISHED -p tcp -s ${ip} --sport 53 -j ACCEPT;
			done`)
}

// configFirewall configures the container firewall.
func (lc *Launcher) configFirewall(container *docker.Container, req *Request) error {
	cmd := []string{"bash", "-c", lc.getDefaultFirewall(strings.Split(lc.req.CocoonAddr, ":")[1])}
	return lc.execInContainer(container, "CONFIG-FIREWALL", cmd, true, configLog, func(state string, val interface{}) error {
		switch state {
		case "before":
			log.Info("Configuring firewall for cocoon")
		case "end":
			if val.(int) == 0 {
				log.Info("Firewall configured for cocoon")
			}
		}
		return nil
	})
}
