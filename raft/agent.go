package raft

import (
	"fmt"
	"log"
	"math"
	"math/rand"
)

type agentState uint

const (
	leader    agentState = 0
	candidate agentState = 1
	follower  agentState = 2
)

const heartBeatFrequency int64 = 50

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
	tick()
	AddCallback(AgentRPC)
	ID() int
}

//Agent type
type Agent struct {
	id        int
	agentRPCs []AgentRPC
	state     agentState
	timeout   int64
	//Persistent state on all servers
	currentTerm int
	votedFor    int
	log         AgentLog
	// Volatile state on all servers
	commitIndex int
	lastApplied int
	// Volatile state on leaders
	nextIndex  map[int]int
	matchIndex map[int]int
	// State to maintain election status
	numVotes int
}

//NewAgent creates a new agent
func NewAgent(id int) *Agent {
	duration := generateTimeoutDuration()
	agent := Agent{
		id:          id,
		agentRPCs:   make([]AgentRPC, 0),
		state:       follower,
		timeout:     duration,
		currentTerm: 0,
		votedFor:    -1,
		log:         newLog(),
		commitIndex: 0,
		lastApplied: 0,
		nextIndex:   make(map[int]int),
		matchIndex:  make(map[int]int),
		numVotes:    0,
	}
	log.Printf("New agent created timeout=%d", agent.timeout)
	return &agent
}

//ID of the agent
func (agent *Agent) ID() int {
	return agent.id
}

func (agent *Agent) tick() {
	agent.timeout--
	if agent.timeout == 0 {
		agent.handleTimeout()
	}
}

func (agent *Agent) handleTimeout() {
	agent.logEvent("act=handleTimeout")
	if agent.state == candidate {
		agent.beginElection()
	} else if agent.state == follower &&
		!agent.grantedVote() {
		agent.beginElection()
	} else if agent.state == leader {
		agent.sendHeartBeat()
	}
	agent.resetTimeout()
}

func (agent *Agent) grantedVote() bool {
	return agent.votedFor != -1
}

func (agent *Agent) beginElection() {
	agent.logEvent("act=beginElection")
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
		lastLogIndex: agent.log.getLastLogIndex(),
		lastLogTerm:  agent.log.getLastLogTerm(),
	}
	for _, otherAgent := range agent.agentRPCs {
		otherAgent.requestVote(voteRequest)
	}
}

func (agent *Agent) handleRequestVoteResponse(response VoteResponse) {
	agent.logEvent("act=handleRequestVoteResponse response=%+v\n", response)
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
	agent.logEvent("act=becomeLeader")
	if agent.state == candidate {
		agent.logEvent("act=becomeLeader transitioning to leader")
		agent.state = leader
		//TODO FINISH THIS
		agent.initialiseNextIndex()
		agent.sendHeartBeat()
	}
}

func (agent *Agent) initialiseNextIndex() {
	for id := range agent.nextIndex {
		agent.nextIndex[id] = agent.log.length()
	}
}

func (agent *Agent) sendHeartBeat() {
	agent.logEvent("act=sendHeartBeat")
	for _, otherAgent := range agent.agentRPCs {
		//TODO finish
		nextIndex := agent.nextIndex[otherAgent.ID()]
		request := agent.createAppendEntriesRequest([]LogEntry{}, nextIndex)
		otherAgent.appendEntries(request)
	}
}

func (agent *Agent) sendAppendEntries() {
	agent.logEvent("act=sendAppendEntries")
	for _, otherAgent := range agent.agentRPCs {
		//TODO finish
		nextIndex := agent.nextIndex[otherAgent.ID()]
		if nextIndex >= agent.log.length() {
			continue
		}
		entries := agent.log.getEntries(nextIndex)
		request := agent.createAppendEntriesRequest(entries, nextIndex)
		otherAgent.appendEntries(request)
	}
}

func (agent *Agent) createAppendEntriesRequest(entries []LogEntry, nextIndex int) AppendEntriesRequest {
	prevLogIndex := 0
	if nextIndex > 0 {
		prevLogIndex = nextIndex - 1
	}
	agent.logEvent("act=sendHeartBeat prevLogIndex=%d logsize=%d", prevLogIndex, agent.log.length())
	prevLogTerm := agent.log.entries[prevLogIndex].Term
	return AppendEntriesRequest{agent.currentTerm, agent.id, prevLogIndex, prevLogTerm, entries, agent.commitIndex}
}

func (agent *Agent) handleRequestVoteRPC(request VoteRequest) VoteResponse {
	agent.logEvent("act=handleRequestVoteRPC request=%+v", request)
	if request.term < agent.currentTerm {
		return VoteResponse{agent.currentTerm, false, agent.id}
	}
	agent.updateTerm(request.term)
	if agent.votedFor == -1 &&
		request.lastLogTerm >= agent.log.getLastLogTerm() &&
		request.lastLogIndex >= agent.log.getLastLogIndex() {

		agent.logEvent("act=handleRequestVoteRPC request=%+v granting vote", request)
		agent.resetTimeout()
		agent.votedFor = request.candidateID
		return VoteResponse{agent.currentTerm, true, agent.id}
	}
	return VoteResponse{agent.currentTerm, false, agent.id}
}

func (agent *Agent) handleAppendEntriesRPC(request AppendEntriesRequest) AppendEntriesResponse {
	agent.logEvent("act=handleAppendEntries request=%+v", request)
	agent.resetTimeout()
	agent.updateTerm(request.term)

	if request.term < agent.currentTerm {
		return AppendEntriesResponse{agent.currentTerm, false, agent.id, -1}
	}
	if agent.log.getLastLogIndex() < request.prevLogIndex ||
		agent.log.entries[request.prevLogIndex].Term != request.prevLogTerm {
		return AppendEntriesResponse{agent.currentTerm, false, agent.id, -1}
	}

	//Else its good to append the logs
	agent.logEvent("act=handleAppendEntries request=%+v\n Appending Entries", request)

	agent.log.addEntriesToLog(request.prevLogIndex, request.entries)
	agent.updateTerm(request.term)
	agent.followerUpdateCommitIndex(request.leaderCommit)
	return AppendEntriesResponse{agent.currentTerm, true, agent.id, agent.log.length()}
}

func (agent *Agent) handleAppendEntriesResponse(response AppendEntriesResponse) {
	agent.logEvent("act=handleAppendEntriesResponse response=%+v", response)
	if response.success {
		agent.nextIndex[response.id] = response.nextIndex
		majorityThreshold := int(math.Ceil(float64(agent.numAgents()) / float64(2)))
		replicated := 1
		for _, nextIndex := range agent.nextIndex {
			if nextIndex >= response.nextIndex {
				replicated++
			}
		}
		if replicated >= majorityThreshold {
			agent.commit(response.nextIndex)
		}
	} else {
		agent.nextIndex[response.id]--
	}
}

func (agent *Agent) commit(commitUpto int) {
	for ii := agent.commitIndex + 1; ii < commitUpto; ii++ {
		// apply the entry at ii
		success := agent.apply(ii)
		if success {
			agent.commitIndex = ii
		} else {
			return
		}
	}
}

func (agent *Agent) apply(index int) bool {
	return true
}

func (agent *Agent) updateTerm(newTerm int) {
	if newTerm > agent.currentTerm {
		agent.state = follower
		agent.votedFor = -1
		agent.currentTerm = newTerm
		agent.resetTimeout()
	}
}

func (agent *Agent) followerUpdateCommitIndex(leaderCommit int) {
	commitUpto := agent.commitIndex
	if agent.commitIndex < leaderCommit {
		if leaderCommit < agent.log.getLastLogIndex() {
			commitUpto = leaderCommit
		} else {
			commitUpto = agent.log.getLastLogIndex()
		}
	}
	agent.commit(commitUpto + 1)
}

func (agent *Agent) resetTimeout() {
	duration := generateTimeoutDuration()
	if agent.state == leader {
		duration = 50
	}
	agent.logEvent("act=resetTimeout duration=%d", duration)
	agent.timeout = duration
}

func generateTimeoutDuration() int64 {
	return rand.Int63n(150) + 150
}

func (agent *Agent) numAgents() int {
	return len(agent.agentRPCs) + 1
}

//AddCallback : use this to add to the callbacks slice for an agent to communicate
func (agent *Agent) AddCallback(rpc AgentRPC) {
	agent.agentRPCs = append(agent.agentRPCs, rpc)
}

func (agent *Agent) logEvent(format string, args ...interface{}) {
	log.Printf("id=%d state=%s currentTerm=%d %s", agent.id, agent.state, agent.currentTerm, fmt.Sprintf(format, args...))
}
