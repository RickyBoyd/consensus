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
	AddCallback(AgentCallback)
	ID() int
}

//Agent type
type Agent struct {
	id             int
	agentCallbacks []AgentCallback
	state          agentState
	// Channels to receive events
	timeout *time.Timer
	//	requestVoteRPC        chan VoteRequest
	//	requestVoteResponse   chan VoteResponse
	//	appendEntriesRPC      chan AppendEntriesRequest
	//	appendEntriesResponse chan AppendEntriesResponse
	//Persistent state on all servers
	currentTerm int
	votedFor    int
	log         []LogEntry
	// Volatile state on all servers
	commitIndex int
	lastApplied int
	// Volatile state on leaders
	nextIndex  []int
	matchIndex []int
	heartbeat  *time.Timer
	// State to maintain election status
	numVotes int
}

//NewAgent creates a new agent
func NewAgent(id int) Agent {
	agent := Agent{
		id:             id,
		agentCallbacks: make([]AgentCallback, 0),
		state:          follower,
		currentTerm:    0,
		votedFor:       -1,
		log:            make([]LogEntry, 0),
		commitIndex:    0,
		lastApplied:    0,
		nextIndex:      make([]int, 0),
		matchIndex:     make([]int, 0),
		numVotes:       0,
	}
	return agent
}

func (agent Agent) ID() int {
	return agent.id
}

func (agent Agent) start() {
	duration := generateTimeoutDuration()
	agent.timeout = time.AfterFunc(duration, agent.handleTimeout)
	fmt.Printf("TImer: %d, ID: %d\n", duration, agent.id)
}

func (agent Agent) handleTimeout() {
	if agent.state == candidate {
		agent.beginElection()
	} else if agent.state == follower {
		agent.beginElection()
	} else {
		fmt.Printf("Timed out when leader")
	}
}

func (agent Agent) handleRequestVoteResponse(response VoteResponse) {
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

func (agent Agent) becomeLeader() {
	if agent.state == candidate {
		agent.state = leader
		//TODO FINISH THIS
		agent.sendHeartBeat()
		fmt.Printf("I am leader: %d\n", agent.ID())
	}
}

func (agent Agent) sendHeartBeat() {
	fmt.Printf("leader heartbeat: %d\n", agent.ID())
	agent.timeout = time.AfterFunc(heartBeatFrequency, agent.sendHeartBeat)
	for _, otherAgent := range agent.agentCallbacks {
		otherAgent.appendEntries(AppendEntriesRequest{agent.currentTerm, agent.id, 0, 0, []LogEntry{}, 0})
	}
}

func (agent Agent) handleRequestVoteRPC(request VoteRequest) VoteResponse {
	fmt.Printf("vote request: %d\n", agent.ID())
	if request.term < agent.currentTerm {
		return VoteResponse{agent.currentTerm, false}
	}
	if agent.votedFor == -1 &&
		request.lastLogTerm >= agent.getLastLogTerm() &&
		request.lastLogIndex >= agent.getLastLogIndex() {
		return VoteResponse{agent.currentTerm, true}
	}
	return VoteResponse{agent.currentTerm, false}
}

func (agent Agent) startTimeout() {
	termTime := generateTimeoutDuration()
	agent.timeout = time.AfterFunc(termTime, agent.beginElection)
}

func (agent Agent) resetTimeout() {
	termTime := generateTimeoutDuration()
	if !agent.timeout.Stop() {
		<-agent.timeout.C
	}
	agent.timeout.Reset(termTime)
}

func generateTimeoutDuration() time.Duration {
	return time.Duration(rand.Int63n(150) + 150)
}

func (agent Agent) beginElection() {
	fmt.Printf("BEGIN ELECTION %d\n", agent.ID())
	agent.state = candidate
	agent.currentTerm++
	agent.numVotes = 1
	agent.requestVotes()
}

func (agent Agent) requestVotes() {
	voteRequest := VoteRequest{
		term:         agent.currentTerm,
		candidateID:  agent.id,
		lastLogIndex: agent.getLastLogIndex(),
		lastLogTerm:  agent.getLastLogTerm(),
	}
	for _, otherAgent := range agent.agentCallbacks {
		otherAgent.requestVote(voteRequest)
	}
}

func (agent Agent) handleAppendEntriesRPC(request AppendEntriesRequest) AppendEntriesResponse {
	fmt.Printf("handleAppendEntries: %d\n", agent.ID())
	agent.resetTimeout()
	// TODO finish implementation of handling AppendLogsRPC
	if request.term < agent.currentTerm {
		return AppendEntriesResponse{agent.currentTerm, false}
	}
	if (len(agent.log) - 1) > request.prevLogIndex {
		return AppendEntriesResponse{agent.currentTerm, false}
	}
	agent.appendLogs(request.prevLogIndex, request.entries)
	return AppendEntriesResponse{}
}

func (agent Agent) appendLogs(prevLogIndex int, entries []LogEntry) {
	index := prevLogIndex
	for _, entry := range entries {
		agent.addToLog(index, entry)
		index++
	}
}

func (agent Agent) addToLog(index int, entry LogEntry) {
	if len(agent.log) <= index {
		agent.log = append(agent.log, entry)
	} else {
		agent.log[index] = entry
	}
}

func (agent Agent) handleAppendEntriesResponse(response AppendEntriesResponse) {

}

func (agent Agent) getLastLogIndex() int {
	return len(agent.log)
}

func (agent Agent) getLastLogTerm() int {
	logLen := len(agent.log)
	if logLen == 0 {
		return 0
	}
	return agent.log[logLen-1].Term
}

func (agent Agent) numAgents() int {
	return len(agent.agentCallbacks) + 1
}

//AddCallback : use this to add to the callbacks slice for an agent to communicate
func (agent Agent) AddCallback(a AgentCallback) {
	agent.agentCallbacks = append(agent.agentCallbacks, a)
}
