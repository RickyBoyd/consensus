package raft

//VoteRequestChan gives the coms manager the channel to respond to
type VoteRequestChan struct {
	request      VoteRequest
	responseChan chan VoteResponse
}

//AppendEntriesRequestChan gives the coms manager the channel to respond to
type AppendEntriesRequestChan struct {
	request      AppendEntriesRequest
	responseChan chan AppendEntriesResponse
}

//AgentChannelRPC manages RPCs for an Agent
type AgentChannelRPC struct {
	id                    int
	requestVoteRPC        chan VoteRequestChan
	requestVoteResponse   chan VoteResponse
	appendEntriesRPC      chan AppendEntriesRequestChan
	appendEntriesResponse chan AppendEntriesResponse
}

func (agentRPC AgentChannelRPC) requestVote(request VoteRequest) {
	agentRPC.requestVoteRPC <- VoteRequestChan{request, agentRPC.requestVoteResponse}
}

func (agentRPC AgentChannelRPC) appendEntries(request AppendEntriesRequest) {
	agentRPC.appendEntriesRPC <- AppendEntriesRequestChan{request, agentRPC.appendEntriesResponse}
}

func (agentRPC AgentChannelRPC) ID() int {
	return agentRPC.id
}

//AgentChannelEventHandler which uses channels to forward requests and responses to an AgentInterface
type AgentChannelEventHandler struct {
	agent                 AgentInterface
	requestVoteRPC        chan VoteRequestChan
	requestVoteResponse   chan VoteResponse
	appendEntriesRPC      chan AppendEntriesRequestChan
	appendEntriesResponse chan AppendEntriesResponse
}

func (eventHandler *AgentChannelEventHandler) start() {
	eventHandler.agent.start()
	eventHandler.eventLoop()
}

func (eventHandler *AgentChannelEventHandler) eventLoop() {
	for true {
		select {
		case request := <-eventHandler.requestVoteRPC:
			// Deal with a request to vote
			//fmt.Printf("Vote RPC\n")
			response := eventHandler.agent.handleRequestVoteRPC(request.request)
			request.responseChan <- response
		case response := <-eventHandler.requestVoteResponse:
			// handle a cast vote
			//fmt.Printf("Vote Response\n")
			eventHandler.agent.handleRequestVoteResponse(response)
		case request := <-eventHandler.appendEntriesRPC:
			// handle entries
			//fmt.Printf("Logs RPC\n")
			response := eventHandler.agent.handleAppendEntriesRPC(request.request)
			request.responseChan <- response
		case response := <-eventHandler.appendEntriesResponse:
			// response to a appendEntriesRPC
			//fmt.Printf("Logs Response\n")
			eventHandler.agent.handleAppendEntriesResponse(response)
		}
	}
}
