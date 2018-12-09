package raft

type MockAgent struct {
	id                     int
	voteResponses          []VoteResponse
	appendEntriesResponses []AppendEntriesResponse
	agentRPCs              []AgentChannelRPC
}

func (agent *MockAgent) handleAppendEntriesRPC(request AppendEntriesRequest) AppendEntriesResponse {
	if len(agent.appendEntriesResponses) > 0 {
		response := agent.appendEntriesResponses[0]
		agent.appendEntriesResponses = agent.appendEntriesResponses[1:]
		return response
	}
	return AppendEntriesResponse{}
}

func (agent *MockAgent) handleAppendEntriesResponse(response AppendEntriesResponse) {

}

func (agent *MockAgent) handleRequestVoteRPC(request VoteRequest) VoteResponse {
	if len(agent.voteResponses) > 0 {
		response := agent.voteResponses[0]
		agent.voteResponses = agent.voteResponses[1:]
		return response
	}
	return VoteResponse{}
}

func (agent *MockAgent) handleRequestVoteResponse(response VoteResponse) {

}

func (agent *MockAgent) start() {

}

func (agent *MockAgent) AddCallback(AgentRPC) {

}

func (agent *MockAgent) ID() int {
	return agent.id
}
