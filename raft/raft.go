package raft

func ConstructRaftChanInstance(numAgents int) []*AgentChannelEventHandler {
	agents := make([]AgentInterface, 0)
	for ii := 0; ii < numAgents; ii++ {
		agents = append(agents, NewAgent(ii))
	}
	return newRaftInstance(agents)
}

func constructEventHandlers(agents []AgentInterface) []*AgentChannelEventHandler {
	numAgents := len(agents)
	agentEventHandlers := make([]*AgentChannelEventHandler, 0)
	for id := 0; id < numAgents; id++ {
		voteRequest := make(chan VoteRequestChan, 1000)
		voteResponse := make(chan VoteResponse, 1000)
		appendEntriesRequest := make(chan AppendEntriesRequestChan, 1000)
		appendEntriesResponse := make(chan AppendEntriesResponse, 1000)
		agentEventHandler := AgentChannelEventHandler{
			agent:                 agents[id],
			requestVoteRPC:        voteRequest,
			requestVoteResponse:   voteResponse,
			appendEntriesRPC:      appendEntriesRequest,
			appendEntriesResponse: appendEntriesResponse,
		}
		agentEventHandlers = append(agentEventHandlers, &agentEventHandler)
	}
	return agentEventHandlers
}

func newRaftInstance(agents []AgentInterface) []*AgentChannelEventHandler {
	// Need to create RPC channels and
	numAgents := len(agents)
	agentEventHandlers := constructEventHandlers(agents)
	for i := 0; i < numAgents; i++ {
		for j := 0; j < numAgents; j++ {
			if j == i {
				continue
			} else {
				callback := AgentChannelRPC{
					id:                    j,
					requestVoteRPC:        agentEventHandlers[j].requestVoteRPC,
					requestVoteResponse:   agentEventHandlers[i].requestVoteResponse,
					appendEntriesRPC:      agentEventHandlers[j].appendEntriesRPC,
					appendEntriesResponse: agentEventHandlers[i].appendEntriesResponse,
				}
				agentEventHandlers[i].agent.AddCallback(callback)
			}
		}
	}
	return agentEventHandlers
}
