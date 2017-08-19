package main

import (
	"flag"
	"fmt"

	"github.com/mausimag/supervisor"
)

var (
	zookeeperServers = flag.String("s", "127.0.0.1", "Zookeeper servers separated by ','")
)

func main() {
	client := supervisor.NewClient(
		supervisor.SetZookeeperNodes(*zookeeperServers),
	)

	if err := client.Connect(); err != nil {
		fmt.Println(err.Error())
	}

	election := supervisor.NewRoleSelector(client, "/election/test01", func() {
		fmt.Println("CURRENT NODE IS: MASTER")
	})

	if err := election.Start(); err != nil {
		fmt.Println(err)
	}
}
