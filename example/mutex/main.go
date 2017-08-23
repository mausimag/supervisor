package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/mausimag/supervisor"
)

var (
	zookeeperServers = flag.String("s", "127.0.0.1", "Zookeeper servers separated by ','")
)

func waitForCtrlC() {
	var endWaiter sync.WaitGroup
	endWaiter.Add(1)
	var signalChannel chan os.Signal
	signalChannel = make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		<-signalChannel
		endWaiter.Done()
	}()
	endWaiter.Wait()
}

func main() {
	client := supervisor.NewClient(
		supervisor.SetZookeeperNodes(*zookeeperServers),
	)

	if err := client.Connect(); err != nil {
		fmt.Println(err.Error())
	}

	lockPath := "/supervisor/example/mutex/key01"
	lock := supervisor.NewMutex(client, lockPath)

	if err := lock.Acquire(60, time.Second); err == nil {
		fmt.Println("Acquired Lock:", lockPath)

		time.Sleep(20 * time.Second)

		if errRelease := lock.Release(); errRelease == nil {
			fmt.Println("Release lock:", lockPath)
		} else {
			fmt.Println("Error Release:", err)
		}
	} else {
		fmt.Println("Error Acquire:", err)
	}

	waitForCtrlC()
}
