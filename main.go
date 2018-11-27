package main

import (
	"consensus/paxos"
	"consensus/raft"
	"fmt"
	"time"
)

func main() {
	fmt.Printf("Hello\n")
	agent := paxos.Agent{ID: 1}
	fmt.Printf("%d\n", agent.ID)

	raftAgent := raft.NewAgent()
	fmt.Printf("agent num %d\n", raftAgent.ID)
	time.Sleep(200)

	ch := make(chan byte, 1)
	<-ch
}
