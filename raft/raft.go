package raft

import "fmt"

func NewRaftInstance(numAgents int) {
	// Need to create RPC channels and
	agents := make([]EventHandler, 0)
	for id := 0; id < numAgents; id++ {
		voteRequest := make(chan VoteRequestChan, 1000)
		voteResponse := make(chan VoteResponse, 1000)
		appendEntriesRequest := make(chan AppendEntriesRequestChan, 1000)
		appendEntriesResponse := make(chan AppendEntriesResponse, 1000)
		agent := NewAgent(id)
		agentEventHandler := EventHandler{
			agent:                 agent,
			requestVoteRPC:        voteRequest,
			requestVoteResponse:   voteResponse,
			appendEntriesRPC:      appendEntriesRequest,
			appendEntriesResponse: appendEntriesResponse,
		}
		agents = append(agents, agentEventHandler)
	}
	for i := 0; i < numAgents; i++ {
		for j := 0; j < numAgents; j++ {
			if j == i {
				continue
			} else {
				callback := ChanRPCImp{
					requestVoteRPC:        agents[j].requestVoteRPC,
					requestVoteResponse:   agents[i].requestVoteResponse,
					appendEntriesRPC:      agents[j].appendEntriesRPC,
					appendEntriesResponse: agents[i].appendEntriesResponse,
				}
				agents[i].agent.AddCallback(callback)
				fmt.Printf("id: %d\n", agents[i].agent.ID)
			}
		}
	}
	for i := 0; i < numAgents; i++ {
		go agents[i].start()
	}
}
