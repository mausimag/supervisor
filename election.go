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
	client           *Client
	roleChangeCb     NodeRoleChangeFunc
	notificationSent bool
	path             string
	nodePath         string
}

// Start starts listening for node role change
func (rs *RoleSelector) Start() error {
	if !rs.client.isConnected {
		return errors.New("Client not connected")
	}

	_, err := rs.client.createParentNodeIfNotExists(rs.path, []byte{})
	if err != nil {
		return err
	}

	abspath, guid, err := rs.client.createProtectedEphemeralSequential(rs.path, []byte{})
	if err != nil {
		return fmt.Errorf("%s - %s", err.Error(), rs.path)
	}
	rs.nodePath = abspath

	for {
		children, _, channel, err := rs.client.childrenWatch(rs.path)
		if err != nil {
			return err
		}

		if len(children) == 1 && !rs.notificationSent {
			rs.client.currentRole = NodeRoleMaster
			rs.notificationSent = true
			go rs.roleChangeCb()
		}

		event := <-channel

		if event.Type == zk.EventNodeChildrenChanged {
			nodeGUIDList, err := rs.client.getSortedNodeGUIDList(rs.path)
			if err != nil {
				return err
			}

			// notify only if current node turns master
			if nodeGUIDList[0] == guid && rs.client.currentRole != NodeRoleMaster && !rs.notificationSent {
				rs.notificationSent = true
				go rs.roleChangeCb()
			}
		}
	}
}

// NewRoleSelector returns new role selector for master election
func NewRoleSelector(c *Client, path string, roleChangeCb NodeRoleChangeFunc) *RoleSelector {
	rs := RoleSelector{
		client:       c,
		path:         path,
		roleChangeCb: roleChangeCb,
	}
	return &rs
}
