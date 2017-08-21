package supervisor

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getClient() (Client, error) {
	client := NewClient(
		SetZookeeperNodes("127.0.0.1"),
	)
	client.Connect()
	return *client, nil
}

func makeClientSlice(q int) []*Client {
	var r []*Client
	for i := 0; i < q; i++ {
		c, _ := getClient()
		r = append(r, &c)
	}
	return r
}

func startListen(idx int, e *RoleSelector, done chan bool) {
	e.Start()
	for {
		select {
		case <-e.IsMaster:
			fmt.Println("CURRENT NODE IS MASTER - ", idx)
			done <- true
		case err := <-e.Error:
			fmt.Println("Error:", err)
			done <- true
		}
	}
}

func TestSimple(t *testing.T) {
	assert := assert.New(t)

	clients := makeClientSlice(2)

	election01 := NewRoleSelector(clients[0], "/election/test01")
	election02 := NewRoleSelector(clients[1], "/election/test01")

	done := make(chan bool, 2)
	go startListen(0, election01, done)
	<-done

	go startListen(1, election02, done)

	assert.Equal(election01.Role, NodeRoleMaster)
	assert.Equal(election02.Role, NodeRoleSlave)

	election01.Stop()
	election02.Stop()
}

func TestSimpleDisconnect(t *testing.T) {
	assert := assert.New(t)

	clients := makeClientSlice(2)

	election01 := NewRoleSelector(clients[0], "/election/test01")
	election02 := NewRoleSelector(clients[1], "/election/test01")

	done := make(chan bool, 2)
	go startListen(0, election01, done)
	<-done

	go startListen(1, election02, done)

	assert.Equal(election01.Role, NodeRoleMaster)
	assert.Equal(election02.Role, NodeRoleSlave)

	election01.Stop()
	time.Sleep(2 * time.Second)

	assert.Equal(election02.Role, NodeRoleMaster)
}
