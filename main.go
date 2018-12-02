package main

import (
	"consensus/raft"
	"fmt"
)

func main() {
	fmt.Printf("Hello\n")

	raft.NewRaftInstance(3)

	//exit := make(chan string)
	// Spawn all you worker goroutines, and send a message to exit when you're done.
	for {
		//select {
		//case <-exit:
		//	{
		//		os.Exit(0)
		//	}
		//}
	}
}
