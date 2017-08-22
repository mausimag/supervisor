package supervisor

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/samuel/go-zookeeper/zk"
)

const (
	// NodeRoleMaster current node is master
	NodeRoleMaster NodeRole = 1

	// NodeRoleSlave current node is slave
	NodeRoleSlave NodeRole = 2
)

// NodeRole current node state (master or slave)
type NodeRole int32

func (ns NodeRole) String() string {
	ss := "Master"
	if ns == 2 {
		return "Slave"
	}
	return ss
}

// NodeRoleChangeFunc callback function when node changes role (Master or Slave)
type NodeRoleChangeFunc func()

// ByNodeGUID order list of nodes by incremental id
type ByNodeGUID []string

func (ni ByNodeGUID) getID(nodeGUID string) int64 {
	id, _ := strconv.ParseInt(nodeGUID[strings.LastIndex(nodeGUID, "-")+1:], 10, 16)
	return id
}

func (ni ByNodeGUID) Len() int {
	return len(ni)
}

func (ni ByNodeGUID) Swap(i, j int) {
	ni[i], ni[j] = ni[j], ni[i]
}

func (ni ByNodeGUID) Less(i, j int) bool {
	return ni.getID(ni[i]) < ni.getID(ni[j])
}

// RoleSelector holds role selector information
type RoleSelector struct {
	client *Client

	notificationSent bool
	path             string
	guid             string
	nodePath         string
	Role             NodeRole

	IsMaster chan bool
	Error    chan error
	close    chan bool
}

// Start starts listening for node role change
func (rs *RoleSelector) Start() {
	if !rs.client.isConnected {
		rs.Error <- errors.New("Client not connected")
	}

	_, err := rs.client.createParentNodeIfNotExists(rs.path, []byte{})
	if err != nil {
		rs.Error <- err
	}

	abspath, guid, err := rs.client.createProtectedEphemeralSequential(rs.path, []byte{})
	if err != nil {
		rs.Error <- fmt.Errorf("%s - %s", err.Error(), rs.path)
	}

	rs.nodePath = abspath
	rs.guid = guid

	go rs.listen()
}

func (rs *RoleSelector) listen() {
	for {
		children, _, channel, err := rs.client.childrenWatch(rs.path)
		if err != nil {
			rs.Error <- err
		}

		if len(children) == 1 && !rs.notificationSent {
			rs.client.currentRole = NodeRoleMaster
			rs.notificationSent = true
			rs.Role = NodeRoleMaster
			rs.IsMaster <- true
		}

		select {
		case event := <-channel:
			rs.notify(event)
		case <-rs.close:

		}
	}
}
func (rs *RoleSelector) notify(event zk.Event) {
	if event.Type == zk.EventNodeChildrenChanged {
		nodeGUIDList, err := rs.client.getSortedNodeGUIDList(rs.path)
		if err != nil {
			rs.Error <- err
		}

		// notify only if current node turns master
		if len(nodeGUIDList) > 0 {
			if nodeGUIDList[0] == rs.guid && rs.client.currentRole == NodeRoleSlave && rs.notificationSent == false {
				rs.notificationSent = true
				rs.Role = NodeRoleMaster
				rs.IsMaster <- true
			}
		}
	}
}

// Stop stops listening for node role change
func (rs *RoleSelector) Stop() error {
	rs.close <- true

	if err := rs.client.deleteNode(rs.nodePath); err != nil {
		return fmt.Errorf("Could not remove node %s - %s", rs.nodePath, err.Error())
	}

	return nil
}

// NewRoleSelector returns new role selector for master election
func NewRoleSelector(c *Client, path string) *RoleSelector {
	rs := RoleSelector{
		client:   c,
		path:     path,
		Role:     NodeRoleSlave,
		IsMaster: make(chan bool),
		Error:    make(chan error),
		close:    make(chan bool),
	}
	return &rs
}
