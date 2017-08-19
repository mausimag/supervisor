Supervisor
====================

Just playing with zookeeper.

Create client:

	client := supervisor.NewClient(
		supervisor.SetZookeeperNodes("10.0.0.1,10.0.0.2,10.0.0.3"),
	)

	if err := client.Connect(); err != nil {
		fmt.Println(err.Error())
	}
	

Leader Election:

	election := supervisor.NewRoleSelector(client, "/election/test01", func() {
		fmt.Println("CURRENT NODE IS: MASTER")
	})

	if err := election.Start(); err != nil {
		fmt.Println(err)
	}


Distributed Lock:

	lock := supervisor.NewMutex(client, "/group01/key01")

	if err := lock.Acquire(maxWait, waitUnit); err == nil {
		fmt.Println("Acquired Lock:", lockPath)

		if errRelease := lock.Release(); errRelease == nil {
			fmt.Println("Release lock:", lockPath)
		} else {
			fmt.Println("Error Release:", err)
		}
	} else {
		fmt.Println("Error Acquire:", err)
	}

Atomic UInt64:

	vint64 := supervisor.NewAtomicUint64(client, "/vars/var01")
	fmt.Println(vint64.TrySet(10))
	fmt.Println(vint64.Get()) // 10
	fmt.Println(vint64.Increment()) // 11
	fmt.Println(vint64.Get()) // 11
	fmt.Println(vint64.Decrement()) // 10
	fmt.Println(vint64.Get()) // 10