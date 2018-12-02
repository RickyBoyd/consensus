package raft

import (
	"fmt"
)

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

//ChanRPCImp manages RPCs for an Agent
type ChanRPCImp struct {
	//	ID                    int
	requestVoteRPC        chan VoteRequestChan
	requestVoteResponse   chan VoteResponse
	appendEntriesRPC      chan AppendEntriesRequestChan
	appendEntriesResponse chan AppendEntriesResponse
}

func (agentRPC ChanRPCImp) requestVote(request VoteRequest) {
	fmt.Printf("Vote 4 me\n")
	agentRPC.requestVoteRPC <- VoteRequestChan{request, agentRPC.requestVoteResponse}
}

func (agentRPC ChanRPCImp) appendEntries(request AppendEntriesRequest) {
	agentRPC.appendEntriesRPC <- AppendEntriesRequestChan{request, agentRPC.appendEntriesResponse}
}

type EventHandler struct {
	agent                 *Agent
	requestVoteRPC        chan VoteRequestChan
	requestVoteResponse   chan VoteResponse
	appendEntriesRPC      chan AppendEntriesRequestChan
	appendEntriesResponse chan AppendEntriesResponse
}

func (eventHandler *EventHandler) start() {
	eventHandler.agent.start()
	eventHandler.eventLoop()
}

func (eventHandler *EventHandler) eventLoop() {
	fmt.Printf("Beginning Event Loop ID: %d\n", eventHandler.agent.ID)
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
			response := eventHandler.agent.handleAppendEntries(request.request)
			request.responseChan <- response
		case response := <-eventHandler.appendEntriesResponse:
			// response to a appendEntriesRPC
			//fmt.Printf("Logs Response\n")
			eventHandler.agent.handleAppendEntriesResponse(response)
		}
	}
}
