package supervisor

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

var defaultClient = &Client{
	clusterName:    "local",
	zookeeperNodes: "127.0.0.1",
}

// Client holds connection information
type Client struct {
	clusterName    string
	zookeeperNodes string

	zkConn      *zk.Conn
	guid        string
	currentRole NodeRole

	currentReceiveMessageCallback NodeReceiveMessageFunc

	isConnected bool
}

// NodeReceiveMessageFunc callback function when node receives message
type NodeReceiveMessageFunc func([]byte)

// NodeOpionsFunc client definition
type NodeOpionsFunc func(*Client) error

// SetZookeeperNodes sets zookeepers ips separated by ','
func SetZookeeperNodes(zookeeperNodes string) NodeOpionsFunc {
	return func(c *Client) error {
		c.zookeeperNodes = zookeeperNodes
		return nil
	}
}

// SetNodeReceiveMessageCallback registers callback function when node receives a message
func SetNodeReceiveMessageCallback(receiveMessageCb NodeReceiveMessageFunc) NodeOpionsFunc {
	return func(c *Client) error {
		c.currentReceiveMessageCallback = receiveMessageCb
		return nil
	}
}

// Connect connects to zookeeper
func (c *Client) Connect() error {
	var err error

	c.zkConn, _, err = zk.Connect(strings.Split(c.zookeeperNodes, ","), time.Second)
	if err != nil {
		return err
	}

	c.isConnected = true

	return nil
}

func (c *Client) checkAndGetNode(path string) ([]byte, *zk.Stat, error) {
	if exists, _, err := c.zkConn.Exists(path); err != nil || !exists {
		return nil, nil, err
	}

	data, stat, err := c.zkConn.Get(path)
	if err != nil {
		return nil, nil, err
	}

	return data, stat, nil
}

func (c *Client) setNodeData(path string, data []byte, version int32) (*zk.Stat, error) {
	return c.zkConn.Set(path, data, version)
}

func (c *Client) createNodeIfNotExists(path string, data []byte) (bool, error) {
	exists, _, err := c.zkConn.Exists(path)
	if err != nil {
		return false, err
	}

	if !exists {
		if _, err := c.zkConn.Create(path, data, 0, zk.WorldACL(zk.PermAll)); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (c *Client) createParentNodeIfNotExists(path string, data []byte) (bool, error) {
	parts := strings.Split(path, "/")
	lparts := len(parts)
	current := ""

	if lparts > 1 {
		for idx := 0; idx < lparts-1; idx++ {
			current += parts[idx]
			c.createNodeIfNotExists(current, []byte{})
			current += "/"
		}
	}

	current += parts[lparts-1]
	return c.createNodeIfNotExists(current, data)
}

func (c *Client) getSortedNodeGUIDList(path string) ([]string, error) {
	nodeListGUID, _, _, err := c.zkConn.ChildrenW(path)
	if err != nil {
		return nil, err
	}
	sort.Sort(ByNodeGUID(nodeListGUID))
	return nodeListGUID, nil
}

func (c *Client) createProtectedEphemeralSequential(path string, data []byte) (string, string, error) {
	npath, err := c.zkConn.CreateProtectedEphemeralSequential(path+"/", data, zk.WorldACL(zk.PermAll))
	if err != nil {
		return "", "", err
	}
	guid := npath[len(path)+1:]
	return npath, guid, nil
}

func (c *Client) childrenWatch(path string) ([]string, *zk.Stat, <-chan zk.Event, error) {
	return c.zkConn.ChildrenW(path)
}

func (c *Client) deleteBaseNode(path string) error {
	parts := strings.Split(strings.TrimLeft(path, "/"), "/")
	lparts := len(parts)

	for idx := 0; idx < lparts; idx++ {
		curr := "/" + strings.Join(parts[:lparts-idx], "/")

		nodeGUIDList, err := c.getSortedNodeGUIDList(curr)
		if err != nil {
			return err
		}

		if len(nodeGUIDList) > 0 {
			return nil
		}

		if err := c.deleteNodeLastVersion(curr); err != nil {
			return fmt.Errorf("Can't remove %s: %s", curr, err.Error())
		}
	}

	return nil
}

func (c *Client) deleteNode(path string, version int32) error {
	return c.zkConn.Delete(path, version)
}

func (c *Client) deleteNodeLastVersion(path string) error {
	exists, stat, err := c.zkConn.Exists(path)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	return c.zkConn.Delete(path, stat.Version)
}

// Disconnect disconect from zk servers
func (c *Client) Disconnect() {
	c.zkConn.Close()
}

// NewClient creates new Supervisor client
func NewClient(options ...NodeOpionsFunc) *Client {
	n := &Client{
		clusterName:    defaultClient.clusterName,
		currentRole:    NodeRoleSlave,
		zookeeperNodes: defaultClient.zookeeperNodes,
	}

	for _, option := range options {
		option(n)
	}

	return n
}
