package monitor

import (
	"errors"
	"fmt"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/ncodes/cocoon/core/config"
	"github.com/olebedev/emitter"
	logging "github.com/op/go-logging"
)

var log *logging.Logger

// ErrNoContainerFound represents a error about not finding containers
var ErrNoContainerFound = errors.New("no container found")

// HandleFunc is the expected function signature
type HandleFunc func(map[string]interface{})

// Report represents the result of the monitors checks
type Report struct {
	DiskUsage int64
	NetTx     uint64
	NetRx     uint64
}

// Monitor defines a launcher monitor module checking resource
// useage of a cocoon code. This module provides a pubsub feature that allows
// other external modules to subscribe to events from it and to also emit events to
// the module.
type Monitor struct {
	sync.Mutex
	emitter     *emitter.Emitter
	containerID string
	stop        bool
	dckClient   *docker.Client
}

// NewMonitor creates a new monitor instance.
func NewMonitor(cocoonID string) *Monitor {
	log = config.MakeLogger("connector.monitor", fmt.Sprintf("cocoon.%s", cocoonID))
	e := emitter.New(10)
	return &Monitor{
		emitter: e,
	}
}

// SetContainerID sets the id of the container to monitor
func (m *Monitor) SetContainerID(cID string) {
	m.containerID = cID
}

// SetDockerClient sets the docker client
func (m *Monitor) SetDockerClient(dckClient *docker.Client) {
	m.dckClient = dckClient
}

// GetEmitter returns the monitor's emitter
func (m *Monitor) GetEmitter() *emitter.Emitter {
	return m.emitter
}

// Stop the monitor
func (m *Monitor) Stop() {
	m.Lock()
	defer m.Unlock()
	m.stop = true
	m.emitter.Off("*")
}

// Reset resets the monitor
func (m *Monitor) Reset() {
	m.Stop()

	m.Lock()
	defer m.Unlock()
	m.emitter = emitter.New(10)
	m.stop = false
}

// getContainerRootSize fetches the total
// size of all the files in the container.
func (m *Monitor) getContainerRootSize() (int64, error) {
	containers, err := m.dckClient.ListContainers(docker.ListContainersOptions{
		All:   true,
		Size:  true,
		Limit: 1,
		Filters: map[string][]string{
			"id": []string{m.containerID},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list containers. %s", err)
	}

	if len(containers) == 0 {
		return 0, ErrNoContainerFound
	}

	return containers[0].SizeRw, nil
}

// getContainerNetworkIO returns the current Net IO usage of the container
func (m *Monitor) getContainerNetworkIO() (uint64, uint64, error) {
	var stats = make(chan *docker.Stats)
	var done = make(chan bool)
	go func() {
		m.dckClient.Stats(docker.StatsOptions{
			ID:      m.containerID,
			Stats:   stats,
			Timeout: 5 * time.Second,
			Done:    done,
		})
	}()

	for s := range stats {
		close(done)
		rxBytes := uint64(0)
		txBytes := uint64(0)
		for _, n := range s.Networks {
			rxBytes += n.RxBytes
			txBytes += n.TxBytes
		}
		return rxBytes, txBytes, nil
	}

	return 0, 0, nil
}

// Monitor starts the monitor
func (m *Monitor) Monitor() {
	for !m.stop {

		size, err := m.getContainerRootSize()
		if err != nil && err != ErrNoContainerFound {
			log.Error(err.Error())
		}

		rxBytes, txBytes, err := m.getContainerNetworkIO()
		if err != nil {
			log.Error(err.Error())
		}

		// log.Debugf("Rx Bytes: %d / Tx Bytes: %d", rxBytes, txBytes)

		report := Report{
			DiskUsage: size,
			NetRx:     rxBytes,
			NetTx:     txBytes,
		}

		<-m.emitter.Emit("monitor.report", report)
		time.Sleep(1 * time.Second)
	}
}
