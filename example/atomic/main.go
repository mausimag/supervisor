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

	vint64 := supervisor.NewAtomicUint64(client, "/supervisor/example/atomic/uint64/var01")
	fmt.Println(vint64.TrySet(5))
	fmt.Println(vint64.Get())
	fmt.Println(vint64.Increment())
	fmt.Println(vint64.Get())
	fmt.Println(vint64.Increment())
	fmt.Println(vint64.Get())
	fmt.Println(vint64.Decrement())
	fmt.Println(vint64.Get())
}
