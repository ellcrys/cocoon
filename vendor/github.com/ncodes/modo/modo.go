// Package modo provides the ability to run a series of commands in a docker container
// while also allowing flexible behaviours like attaching a simple function to collect logs,
// stopping the series of commands if one fails, continuing and enabling privileged mode per command
// collecting the full output of each commands and exit code.
package modo

import (
	"fmt"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

// OutputFunc represents a callback function to receive command output
type OutputFunc func(d []byte, stdout bool)

// State represents a lifecycle state of a command
type State string

var (

	// Begin indicates the start of the execution of all commands
	Begin State = "begin"

	// Before indicates the point before a command is run
	Before State = "before"

	// Executing indicates a command that is running
	Executing State = "executing"

	// After indicates a command that has completed
	After State = "after"

	// End indicates the end of execution of all commands
	End State = "end"
)

// DockerSock points to docker socket file
var DockerSock = "unix:///var/run/docker.sock"

// Do defines a command to execute along with resources and directives specific to it.
// Setting `AbortSeriesOnFail` to true will force the execution of a series of Do task to be aborted
// if the command fails. Set `Privileged` to run this command in privileged mode.
// Attach an OutputCallback to receive stdout and stderr streams
// If OutputCallback is not set, the MoDo.OutputCB will be used.
// If KeepOutput is true, output will be store in the Output field.
// Done and ExitCode will be set to true once command completes.
type Do struct {
	Cmd               []string
	AbortSeriesOnFail bool
	Privileged        bool
	OutputCB          OutputFunc
	StateCB           func(state State, task *Do)
	Output            []byte
	KeepOutput        bool
	ExitCode          int
	Done              bool
}

// MoDo defines a structure for collection
// commands to execute in series, manage and collect
// outputs and more
type MoDo struct {
	containerID string
	series      bool
	tasks       []*Do
	client      *docker.Client
	outputCB    OutputFunc
	stateCB     func(state State, task *Do)
	privileged  bool
}

// NewMoDo creates and returns a new MoDo instance.
// Set series to true to run the commands in series. If series
// is false, all commands are executed in parallel.
// Set privileged to true if all commands should be run in privileged mode.
// Attach an output callback function to receive stdout/stderr streams
func NewMoDo(containerID string, series bool, privileged bool, outputCB OutputFunc) *MoDo {
	return &MoDo{
		containerID: containerID,
		series:      series,
		outputCB:    outputCB,
		privileged:  privileged,
	}
}

// SetStateCB sets the life cycle state callback for the MoDo instance.
// This callback will be called at every state of all commands.
func (m *MoDo) SetStateCB(cb func(State, *Do)) {
	m.stateCB = cb
}

// UseClient allows an already existing client to be used
func (m *MoDo) UseClient(client *docker.Client) {
	m.client = client
}

// Add adds a new command to execute.
func (m *MoDo) Add(do ...*Do) {
	m.tasks = append(m.tasks, do...)
}

// GetTasks the tasks
func (m *MoDo) GetTasks() []*Do {
	return m.tasks
}

func (m *MoDo) exec(task *Do) error {
	exec, err := m.client.CreateExec(docker.CreateExecOptions{
		Container:    m.containerID,
		Cmd:          task.Cmd,
		AttachStderr: !(m.outputCB == nil && task.OutputCB == nil),
		AttachStdout: !(m.outputCB == nil && task.OutputCB == nil),
		Privileged:   !(!m.privileged && !task.Privileged),
	})
	if err != nil {
		return err
	}

	outputFunc := task.OutputCB
	if outputFunc == nil {
		outputFunc = m.outputCB
	}

	privileged := task.Privileged
	if !privileged {
		privileged = m.privileged
	}

	outOutputter := NewOutputter(func(d []byte) {
		if outputFunc != nil {
			outputFunc(d, true)
		}
		if task.KeepOutput {
			task.Output = append(task.Output, d...)
		}
	})

	errOutputter := NewOutputter(func(d []byte) {
		if outputFunc != nil {
			outputFunc(d, false)
		}
		if task.KeepOutput {
			task.Output = append(task.Output, d...)
		}
	})

	go outOutputter.Start()
	go errOutputter.Start()

	if task.StateCB != nil {
		task.StateCB(Before, task)
	}
	if m.stateCB != nil {
		m.stateCB(Before, task)
	}

	_, err = m.client.StartExecNonBlocking(exec.ID, docker.StartExecOptions{
		OutputStream: outOutputter.GetWriter(),
		ErrorStream:  errOutputter.GetWriter(),
	})
	if err != nil {
		return err
	}

	// give some time for the execution to start before watching
	time.Sleep(50 * time.Millisecond)

	var executed = false
	var running = true
	for running {
		execStatus, err := m.client.InspectExec(exec.ID)
		if err != nil {
			return err
		}
		running = execStatus.Running
		if running && !executed {
			if task.StateCB != nil {
				task.StateCB(Executing, task)
			}
			if m.stateCB != nil {
				m.stateCB(Executing, task)
			}
			executed = true
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err != nil {
		outOutputter.Stop()
		errOutputter.Stop()
		return err
	}

	// give the outputter some time read all the logs
	time.Sleep(50 * time.Millisecond)

	// stop the outputters
	errOutputter.Stop()
	outOutputter.Stop()

	execIns, err := m.client.InspectExec(exec.ID)
	if err != nil {
		return fmt.Errorf("failed to inspect exec: %s", err)
	}

	task.ExitCode = execIns.ExitCode

	if task.StateCB != nil {
		task.StateCB(After, task)
	}
	if m.stateCB != nil {
		m.stateCB(After, task)
	}

	return err
}

// Do execs all the commands in series or parallel.
// It returns a list of all task related error ([]error) and
// a general error.
func (m *MoDo) Do() ([]error, error) {

	var err error
	var errs []error

	if m.client == nil {
		m.client, err = docker.NewClient(DockerSock)
		if err != nil {
			return nil, err
		}
	}

	// ensure container is running
	_, err = m.client.InspectContainer(m.containerID)
	if err != nil {
		return nil, err
	}

	if m.stateCB != nil {
		m.stateCB(Begin, nil)
	}

	for i, task := range m.tasks {
		if m.series {
			err = m.exec(task)
			task.Done = true
			if err != nil {
				return nil, err
			}

			if task.ExitCode != 0 {
				if task.AbortSeriesOnFail {
					errs = append(errs, fmt.Errorf("task: %d exited with exit code: %d", i, task.ExitCode))
					break
				}
				errs = append(errs, fmt.Errorf("task: %d exited with exit code: %d", i, task.ExitCode))
			}
		}
	}

	if m.stateCB != nil {
		m.stateCB(End, nil)
	}

	return errs, nil
}
