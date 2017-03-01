package launcher

import (
	"fmt"
	"os"
	"os/exec"

	"path"

	"time"

	"strings"

	"github.com/ellcrys/crypto"
	"github.com/ellcrys/util"
	cutil "github.com/ncodes/cocoon-util"
	"github.com/ncodes/cocoon/core/connector/client"
	"github.com/ncodes/cocoon/core/connector/config"
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
	failed    chan bool
	languages []Language
	container *docker.Container
	client    *client.Client
}

// NewLauncher creates a new launcher
func NewLauncher(failed chan bool) *Launcher {
	return &Launcher{
		failed: failed,
		client: client.NewClient(8000),
	}
}

// setFailed sends true or false to the
// launch failed channel to indicate success or failure
func (lc *Launcher) setFailed(v bool) {
	lc.failed <- v
}

// Launch starts a cocoon code
func (lc *Launcher) Launch(req *Request) {

	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	if err != nil {
		log.Errorf("failed to create docker client. Is dockerd running locally?. %s", err)
		lc.setFailed(true)
		return
	}

	dckClient = client

	log.Info("Ready to install cocoon code")
	log.Debugf("Found ccode url=%s and lang=%s", req.URL, req.Lang)

	lang := lc.GetLanguage(req.Lang)
	if lang == nil {
		log.Errorf("cocoon code language (%s) not supported", req.Lang)
		lc.setFailed(true)
		return
	}

	_, err = lc.fetchSource(req, lang)
	if err != nil {
		log.Error(err.Error())
		lc.setFailed(true)
		return
	}

	// ensure cocoon code isn't already launched on a container
	c, err := lc.getContainer(req.ID)
	if err != nil {
		log.Errorf("failed to check whether cocoon code is already active. %s ", err.Error())
		lc.setFailed(true)
		return
	} else if c != nil {
		log.Error("cocoon code already exists on a container")
		lc.setFailed(true)
		return
	}

	newContainer, err := lc.createContainer(
		req.ID,
		lang.GetImage(),
		lang.GetDownloadDestination(req.URL),
		lang.GetMountDestination(req.URL),
		[]string{
			fmt.Sprintf("COCOON_ID=%s", req.ID),
		})
	if err != nil {
		log.Errorf("failed to create new container to run cocoon code. %s ", err.Error())
		lc.setFailed(true)
		return
	}

	lc.container = newContainer

	if lang.RequiresBuild() {
		var buildParams map[string]interface{}
		if len(req.BuildParams) > 0 {

			req.BuildParams, err = crypto.FromBase64(req.BuildParams)
			if err != nil {
				log.Errorf("failed to decode build parameter. Expects a base 64 encoded string. %s", err)
				lc.setFailed(true)
				return
			}

			if err = util.FromJSON([]byte(req.BuildParams), &buildParams); err != nil {
				log.Errorf("failed to parse build parameter. Expects valid json string. %s", err)
				lc.setFailed(true)
				return
			}
		}

		if err = lang.SetBuildParams(buildParams); err != nil {
			log.Errorf("failed to set and validate build parameter. %s", err)
			lc.setFailed(true)
			return
		}

		err = lc.build(newContainer, lang)
		if err != nil {
			log.Errorf(err.Error())
			lc.setFailed(true)
			lc.stopContainer(newContainer.ID)
			return
		}
	} else {
		log.Info("Cocoon code does not require a build processing. Skipped.")
	}

	if err = lc.configFirewall(newContainer, req); err != nil {
		log.Error(err.Error())
		lc.setFailed(true)
		return
	}

	if err = lc.run(newContainer, lang); err != nil {
		log.Error(err.Error())
		lc.setFailed(true)
		return
	}
}

// Stop stops the launcher and the container it is running
func (lc *Launcher) Stop() error {

	containerStatus, err := dckClient.InspectContainer(lc.container.ID)
	if err != nil {
		return fmt.Errorf("failed to inspect container before running. %s", err)
	}

	if !containerStatus.State.Running {
		if err = lc.stopContainer(lc.container.ID); err != nil {
			return fmt.Errorf("failed to stop container")
		}
	}

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
	downloadDst = lang.GetDownloadDestination(req.URL)

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
// mounts the source directory and set the mount
// directory as the work directory
func (lc *Launcher) createContainer(name, image, sourceDir, mountDir string, env []string) (*docker.Container, error) {
	container, err := dckClient.CreateContainer(docker.CreateContainerOptions{
		Name: name,
		Config: &docker.Config{
			Image:      image,
			Labels:     map[string]string{"name": name, "type": "cocoon_code"},
			WorkingDir: mountDir,
			Tty:        true,
			ExposedPorts: map[docker.Port]struct{}{
				"8000/tcp": struct{}{},
			},
			Cmd: []string{"bash"},
			Env: env,
		},
		HostConfig: &docker.HostConfig{
			PortBindings: map[docker.Port][]docker.PortBinding{
				"8000/tcp": []docker.PortBinding{
					docker.PortBinding{HostIP: "127.0.0.1", HostPort: "8000"},
				},
			},
			Mounts: []docker.Mount{
				docker.Mount{
					Type:     "bind",
					Source:   sourceDir,
					Target:   mountDir,
					ReadOnly: false,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

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

	for {
		execIns, err := dckClient.InspectExec(exec.ID)
		if err != nil {
			return fmt.Errorf("failed to inspect run exec [%s] op. %s", name, err)
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
func (lc *Launcher) getDefaultFirewall() string {
	return strings.TrimSpace(`iptables -F && 
			iptables -P INPUT ACCEPT && 
			iptables -P FORWARD DROP &&
			iptables -P OUTPUT DROP &&
			iptables -A INPUT -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT &&
			iptables -A INPUT -p tcp --sport 8000 -j ACCEPT
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
	cmd := []string{"bash", "-c", lc.getDefaultFirewall()}
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
