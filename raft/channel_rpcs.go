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
	id                    int64
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

func (agentRPC AgentChannelRPC) ID() int64 {
	return agentRPC.id
}

//AgentChannelEventHandler which uses channels to forward requests and responses to an Agent
type AgentChannelEventHandler struct {
	Agent                 Agent
	requestVoteRPC        chan VoteRequestChan
	requestVoteResponse   chan VoteResponse
	appendEntriesRPC      chan AppendEntriesRequestChan
	appendEntriesResponse chan AppendEntriesResponse
}

//Tick function will advance time by a single tick, must be called regularly
func (eventHandler *AgentChannelEventHandler) Tick() {
	eventsLeft := true
	for eventsLeft {
		select {
		case request := <-eventHandler.requestVoteRPC:
			// Deal with a request to vote
			//log.Printf("Vote RPC\n")
			response := eventHandler.Agent.handleRequestVoteRPC(request.request)
			request.responseChan <- response
		case response := <-eventHandler.requestVoteResponse:
			// handle a cast vote
			//log.Printf("Vote Response\n")
			eventHandler.Agent.handleRequestVoteResponse(response)
		case request := <-eventHandler.appendEntriesRPC:
			// handle entries
			//log.Printf("Logs RPC\n")
			response := eventHandler.Agent.handleAppendEntriesRPC(request.request)
			request.responseChan <- response
		case response := <-eventHandler.appendEntriesResponse:
			// response to a appendEntriesRPC
			//log.Printf("Logs Response\n")
			eventHandler.Agent.handleAppendEntriesResponse(response)
		default:
			eventsLeft = false
		}
	}
	eventHandler.Agent.tick()
}
