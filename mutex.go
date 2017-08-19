package supervisor

import (
	"errors"
	"fmt"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

// Mutex holds mutex information
type Mutex struct {
	client   *Client
	key      string
	path     string
	lockPath string
	guid     string
	locked   bool
}

// Acquire blocks until it's available
func (m *Mutex) Acquire(waitTime int64, unit time.Duration) error {
	if !m.client.isConnected {
		return errors.New("Client not connected")
	}

	_, err := m.client.createParentNodeIfNotExists(m.path, []byte{})
	if err != nil {
		return err
	}

	abspath, guid, err := m.client.createProtectedEphemeralSequential(m.path, []byte{})
	if err != nil {
		return fmt.Errorf("%s - %s", err.Error(), m.path)
	}

	m.lockPath = abspath
	m.guid = guid

	for !m.locked {
		children, _, channel, err := m.client.childrenWatch(m.path)
		if err != nil {
			return fmt.Errorf("%s - %s", err.Error(), m.path)
		}

		if len(children) == 1 {
			m.locked = true
			break
		}

		timeout := make(chan bool, 0)
		go func() {
			time.Sleep(time.Duration(waitTime) * unit)
			timeout <- true
		}()

		select {
		case <-timeout:
			return err
		case event := <-channel:
			if event.Type == zk.EventNodeChildrenChanged {
				nodeGUIDList, err := m.client.getSortedNodeGUIDList(m.path)
				if err != nil {
					return err
				}

				if nodeGUIDList[0] == m.guid {
					m.locked = true
					break
				}
			}
		}
	}

	return nil
}

// Release performs one release of the mutex
func (m *Mutex) Release() error {
	if !m.locked {
		return errors.New("Key [" + m.key + "] not locked")
	}

	if err := m.cleanup(); err != nil {
		return err
	}

	return nil
}

func (m *Mutex) cleanup() error {
	if err := m.client.deleteNode(m.lockPath); err != nil {
		return fmt.Errorf("Could not remove node %s - %s", m.path, err.Error())
	}

	m.locked = false

	nodeGUIDList, err := m.client.getSortedNodeGUIDList(m.path)
	if err != nil {
		return err
	}

	if len(nodeGUIDList) == 0 {
		if err := m.client.deleteBaseNode(m.path); err != nil {
			return err
		}
	}

	m.guid = ""
	m.lockPath = ""
	return nil
}

// NewMutex returns new mutex for distributed lock
func NewMutex(c *Client, path string) *Mutex {
	m := Mutex{
		client: c,
		path:   path,
		locked: false,
	}
	return &m
}
