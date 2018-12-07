package raft

import "fmt"

func ConstructRaftChanInstance(numAgents int) {
	agents := make([]AgentInterface, 0)
	for ii := 0; ii < numAgents; ii++ {
		agents = append(agents, NewAgent(ii))
	}
	newRaftInstance(agents)
}

func newRaftInstance(agents []AgentInterface) {
	// Need to create RPC channels and
	numAgents := len(agents)
	agentEventHandlers := make([]EventHandler, 0)
	for id := 0; id < numAgents; id++ {
		voteRequest := make(chan VoteRequestChan, 1000)
		voteResponse := make(chan VoteResponse, 1000)
		appendEntriesRequest := make(chan AppendEntriesRequestChan, 1000)
		appendEntriesResponse := make(chan AppendEntriesResponse, 1000)
		agentEventHandler := EventHandler{
			agent:                 agents[id],
			requestVoteRPC:        voteRequest,
			requestVoteResponse:   voteResponse,
			appendEntriesRPC:      appendEntriesRequest,
			appendEntriesResponse: appendEntriesResponse,
		}
		agentEventHandlers = append(agentEventHandlers, agentEventHandler)
	}
	for i := 0; i < numAgents; i++ {
		for j := 0; j < numAgents; j++ {
			if j == i {
				continue
			} else {
				callback := ChanRPCImp{
					requestVoteRPC:        agentEventHandlers[j].requestVoteRPC,
					requestVoteResponse:   agentEventHandlers[i].requestVoteResponse,
					appendEntriesRPC:      agentEventHandlers[j].appendEntriesRPC,
					appendEntriesResponse: agentEventHandlers[i].appendEntriesResponse,
				}
				agentEventHandlers[i].agent.AddCallback(callback)
				fmt.Printf("id: %d\n", agentEventHandlers[i].agent.ID())
			}
		}
	}
	for i := 0; i < numAgents; i++ {
		go agentEventHandlers[i].start()
	}
}
