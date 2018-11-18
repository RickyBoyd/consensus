package main

import "fmt"
import "consensus/paxos"

func main() {
	fmt.Printf("Hello\n")
	agent := paxos.Agent{ID: 1}
	fmt.Printf("%d\n", agent.ID)
}
