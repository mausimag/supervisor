package supervisor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeClientSlice(q int) []*Client {
	var r []*Client
	for i := 0; i < q; i++ {
		c := NewClient(
			SetZookeeperNodes("127.0.0.1"),
		)
		c.Connect()
		r = append(r, c)
	}
	return r
}

func closeClients(clients []*Client) {
	for _, client := range clients {
		client.Disconnect()
	}
}

func createElection(clients []*Client) []*RoleSelector {
	lc := len(clients)
	r := make([]*RoleSelector, lc)

	// create master
	election01 := NewRoleSelector(clients[0], "/supervisor/test/election")
	election01.Start()

	// wait for the master
	<-election01.IsMaster
	r[0] = election01
	close(election01.IsMaster)

	// start slaves
	for idx := 1; idx < lc; idx++ {
		e := NewRoleSelector(clients[idx], "/supervisor/test/election")
		e.Start()
		r[idx] = e
	}

	return r
}
func TestElectionSimple(t *testing.T) {
	assert := assert.New(t)
	clients := makeClientSlice(2)
	election := createElection(clients)

	assert.Equal(election[0].Role, NodeRoleMaster)
	assert.Equal(election[1].Role, NodeRoleSlave)

	election[1].Stop()
	election[0].Stop()

	closeClients(clients)
}
func TestElectionDisconnect(t *testing.T) {
	assert := assert.New(t)
	clients := makeClientSlice(2)
	election := createElection(clients)

	assert.Equal(election[0].Role, NodeRoleMaster)
	assert.Equal(election[1].Role, NodeRoleSlave)

	election[0].Stop()

	<-election[1].IsMaster

	assert.Equal(election[1].Role, NodeRoleMaster)

	election[1].Stop()

	closeClients(clients)
}
