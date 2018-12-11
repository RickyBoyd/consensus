package raft

import (
	"fmt"
	"math/rand"
	"time"
)

type agentState uint

const (
	leader    agentState = 0
	candidate agentState = 1
	follower  agentState = 2
)

const heartBeatFrequency time.Duration = 50

func (state agentState) String() string {
	names := [...]string{
		"leader",
		"candidate",
		"follower"}

	return names[state]
}

//LogEntry type
type LogEntry struct {
	Term    int
	Command int
}

//AgentInterface defines the methods an agent needs to implement to interact with the outside wordl
type AgentInterface interface {
	handleAppendEntriesRPC(request AppendEntriesRequest) AppendEntriesResponse
	handleAppendEntriesResponse(response AppendEntriesResponse)
	handleRequestVoteRPC(request VoteRequest) VoteResponse
	handleRequestVoteResponse(response VoteResponse)
	start()
	AddCallback(AgentRPC)
	ID() int
}

//Agent type
type Agent struct {
	id        int
	agentRPCs []AgentRPC
	state     agentState
	// Channels to receive events
	timeout *time.Timer
	//Persistent state on all servers
	currentTerm int
	votedFor    int
	log         []LogEntry
	// Volatile state on all servers
	commitIndex int
	lastApplied int
	// Volatile state on leaders
	nextIndex  map[int]int
	matchIndex map[int]int
	heartbeat  *time.Timer
	// State to maintain election status
	numVotes int
}

//NewAgent creates a new agent
func NewAgent(id int) *Agent {
	agent := Agent{
		id:          id,
		agentRPCs:   make([]AgentRPC, 0),
		state:       follower,
		timeout:     time.NewTimer(0),
		currentTerm: 0,
		votedFor:    -1,
		log:         []LogEntry{LogEntry{0, 0}},
		commitIndex: 0,
		lastApplied: 0,
		nextIndex:   make(map[int]int),
		matchIndex:  make(map[int]int),
		heartbeat:   time.NewTimer(0),
		numVotes:    0,
	}
	return &agent
}

func (agent *Agent) ID() int {
	return agent.id
}

func (agent *Agent) start() {
	duration := generateTimeoutDuration()
	agent.timeout = time.AfterFunc(duration, agent.handleTimeout)
	fmt.Printf("TImer: %d, ID: %d\n", duration, agent.id)
}

func (agent *Agent) handleTimeout() {
	if agent.state == candidate {
		agent.beginElection()
	} else if agent.state == follower &&
		!agent.grantedVote() {
		agent.beginElection()
	} else {
		fmt.Printf("Timed out when leader? ID:%d\n", agent.id)
	}
	agent.resetTimeout()
}

func (agent *Agent) grantedVote() bool {
	return agent.votedFor != -1
}

func (agent *Agent) beginElection() {
	fmt.Printf("BEGIN ELECTION %d\n", agent.ID())
	agent.state = candidate
	agent.currentTerm++
	agent.votedFor = agent.id
	agent.numVotes = 1
	agent.requestVotes()
}

func (agent *Agent) requestVotes() {
	voteRequest := VoteRequest{
		term:         agent.currentTerm,
		candidateID:  agent.id,
		lastLogIndex: agent.getLastLogIndex(),
		lastLogTerm:  agent.getLastLogTerm(),
	}
	for _, otherAgent := range agent.agentRPCs {
		otherAgent.requestVote(voteRequest)
	}
}

func (agent *Agent) handleRequestVoteResponse(response VoteResponse) {
	fmt.Printf("handleRequestVoteResponse: %d\n", agent.ID())
	if response.votedFor {
		agent.numVotes++
		if agent.numVotes > agent.numAgents()/2 {
			//THEN BECOME LEADER
			agent.becomeLeader()
		}
	}
	//TODO finish
}

func (agent *Agent) becomeLeader() {
	if agent.state == candidate {
		agent.state = leader
		//TODO FINISH THIS
		agent.initialiseNextIndex()
		agent.sendHeartBeat()
		agent.stopElectionTimeout()
		fmt.Printf("I am leader: %d\n", agent.ID())
	}
}

func (agent *Agent) initialiseNextIndex() {
	for id := range agent.nextIndex {
		agent.nextIndex[id] = len(agent.log)
	}
}

func (agent *Agent) sendHeartBeat() {
	//fmt.Printf("leader heartbeat: %d\n", agent.ID())
	agent.timeout = time.AfterFunc(heartBeatFrequency, agent.sendHeartBeat)
	for _, otherAgent := range agent.agentRPCs {
		//TODO finish
		nextIndex := agent.nextIndex[otherAgent.ID()]
		entries := agent.log[nextIndex:]
		prevLogIndex := 0
		if nextIndex > 0 {
			prevLogIndex = nextIndex - 1
		}
		prevLogTerm := agent.log[prevLogIndex].Term
		otherAgent.appendEntries(AppendEntriesRequest{agent.currentTerm, agent.id, prevLogIndex, prevLogTerm, entries, agent.commitIndex})
	}
}

func (agent *Agent) sendAppendEntries() {

}

func (agent *Agent) handleRequestVoteRPC(request VoteRequest) VoteResponse {
	fmt.Printf("vote request: %d from %d\n", agent.ID(), request.candidateID)
	if request.term < agent.currentTerm {
		return VoteResponse{agent.currentTerm, false, agent.id}
	}
	if agent.votedFor == -1 &&
		request.lastLogTerm >= agent.getLastLogTerm() &&
		request.lastLogIndex >= agent.getLastLogIndex() {

		fmt.Printf("Voting for %v from %d\n", request, agent.id)
		agent.votedFor = request.candidateID
		agent.updateTerm(request.term)
		return VoteResponse{agent.currentTerm, true, agent.id}
	}
	return VoteResponse{agent.currentTerm, false, agent.id}
}

func (agent *Agent) startTimeout() {
	termTime := generateTimeoutDuration()
	agent.timeout = time.AfterFunc(termTime, agent.beginElection)
}

func (agent *Agent) resetTimeout() {
	termTime := generateTimeoutDuration()
	if !agent.timeout.Stop() {
		<-agent.timeout.C
	}
	agent.timeout.Reset(termTime)
}

func (agent *Agent) stopElectionTimeout() {
	if !agent.timeout.Stop() {
		<-agent.timeout.C
	}
}

func generateTimeoutDuration() time.Duration {
	return time.Duration(rand.Int63n(150) + 150)
}

func (agent *Agent) handleAppendEntriesRPC(request AppendEntriesRequest) AppendEntriesResponse {
	fmt.Printf("handleAppendEntries: %d\n", agent.ID())
	agent.resetTimeout()
	// TODO finish implementation of handling AppendLogsRPC
	agent.state = follower
	agent.votedFor = -1
	if request.term < agent.currentTerm {
		return AppendEntriesResponse{agent.currentTerm, false, agent.id}
	}
	if (len(agent.log) - 1) < request.prevLogIndex {
		return AppendEntriesResponse{agent.currentTerm, false, agent.id}
	}
	agent.addEntriesToLog(request.prevLogIndex, request.entries)
	fmt.Printf("Here in append rpc: myterm %d, req term: %d\n", agent.currentTerm, request.term)
	agent.updateTerm(request.term)
	//TODO fill in
	return AppendEntriesResponse{agent.currentTerm, true, agent.id}
}

func (agent *Agent) addEntriesToLog(prevLogIndex int, entries []LogEntry) {
	index := prevLogIndex + 1
	for _, entry := range entries {
		agent.addToLog(index, entry)
		index++
	}
	agent.log = agent.log[:index]
}

func (agent *Agent) addToLog(index int, entry LogEntry) {
	if index >= len(agent.log) {
		agent.log = append(agent.log, entry)
	} else {
		agent.log[index] = entry
	}
}

func (agent *Agent) handleAppendEntriesResponse(response AppendEntriesResponse) {
	if response.success {

	} else {
		agent.nextIndex[response.id]--
	}
}

func (agent *Agent) updateTerm(newTerm int) {
	if newTerm > agent.currentTerm {
		agent.state = follower
		agent.currentTerm = newTerm
		agent.resetTimeout()
	}
}

func (agent *Agent) getLastLogIndex() int {
	return len(agent.log)
}

func (agent *Agent) getLastLogTerm() int {
	logLen := len(agent.log)
	if logLen == 0 {
		return 0
	}
	return agent.log[logLen-1].Term
}

func (agent *Agent) numAgents() int {
	return len(agent.agentRPCs) + 1
}

//AddCallback : use this to add to the callbacks slice for an agent to communicate
func (agent *Agent) AddCallback(rpc AgentRPC) {
	agent.agentRPCs = append(agent.agentRPCs, rpc)
}
