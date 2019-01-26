package raft

func ConstructRaftChanInstance(numAgents int) []*AgentChannelEventHandler {
	agents := make([]Agent, 0)
	for ii := 0; ii < numAgents; ii++ {
		agents = append(agents, *NewAgent(int64(ii)))
	}
	return newRaftInstance(agents)
}

func constructEventHandlers(agents []Agent) []*AgentChannelEventHandler {
	numAgents := len(agents)
	agentEventHandlers := make([]*AgentChannelEventHandler, 0)
	for id := 0; id < numAgents; id++ {
		voteRequest := make(chan VoteRequestChan, 1000)
		voteResponse := make(chan VoteResponse, 1000)
		appendEntriesRequest := make(chan AppendEntriesRequestChan, 1000)
		appendEntriesResponse := make(chan AppendEntriesResponse, 1000)
		agentEventHandler := AgentChannelEventHandler{
			Agent:                 agents[id],
			requestVoteRPC:        voteRequest,
			requestVoteResponse:   voteResponse,
			appendEntriesRPC:      appendEntriesRequest,
			appendEntriesResponse: appendEntriesResponse,
		}
		agentEventHandlers = append(agentEventHandlers, &agentEventHandler)
	}
	return agentEventHandlers
}

func newRaftInstance(agents []Agent) []*AgentChannelEventHandler {
	// Need to create RPC channels and
	numAgents := len(agents)
	agentEventHandlers := constructEventHandlers(agents)
	for i := 0; i < numAgents; i++ {
		for j := 0; j < numAgents; j++ {
			if j == i {
				continue
			} else {
				callback := AgentChannelRPC{
					id:                    int64(j),
					requestVoteRPC:        agentEventHandlers[j].requestVoteRPC,
					requestVoteResponse:   agentEventHandlers[i].requestVoteResponse,
					appendEntriesRPC:      agentEventHandlers[j].appendEntriesRPC,
					appendEntriesResponse: agentEventHandlers[i].appendEntriesResponse,
				}
				agentEventHandlers[i].Agent.AddCallback(callback)
			}
		}
	}
	return agentEventHandlers
}
