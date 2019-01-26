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
	notVoted  int64      = -1
)

const heartBeatFrequency int64 = 50

func (state agentState) String() string {
	names := [...]string{
		"leader",
		"candidate",
		"follower"}

	return names[state]
}

//AgentRPC interface defines callbacks
type AgentRPC interface {
	requestVote(vote VoteRequest)
	appendEntries(request AppendEntriesRequest)
	ID() int64
}

//Agent type
type Agent struct {
	id        int64
	agentRPCs []AgentRPC
	state     agentState
	timeout   int64
	//Persistent state on all servers
	currentTerm int64
	votedFor    int64
	log         AgentLog
	// Volatile state on all servers
	commitIndex int64
	lastApplied int64
	// Volatile state on leaders
	nextIndex  map[int64]int64
	matchIndex map[int64]int64
	// State to maintain election status
	numVotes int64
}

//NewAgent creates a new agent
func NewAgent(id int64) *Agent {
	duration := generateTimeoutDuration()
	agent := Agent{
		id:          id,
		agentRPCs:   make([]AgentRPC, 0),
		state:       follower,
		timeout:     duration,
		currentTerm: 0,
		votedFor:    notVoted,
		log:         newLog(),
		commitIndex: 0,
		lastApplied: 0,
		nextIndex:   make(map[int64]int64),
		matchIndex:  make(map[int64]int64),
		numVotes:    0,
	}
	log.Printf("New agent created timeout=%d", agent.timeout)
	return &agent
}

func (agent *Agent) ClientAction(action int64) {
	entry := &LogEntry{
		Term:    agent.currentTerm,
		Command: action,
	}
	agent.log.appendToLog(entry)
	agent.sendAppendEntries()
}

//ID of the agent
func (agent *Agent) ID() int64 {
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
	} else if agent.state == follower {
		agent.beginElection()
	} else if agent.state == leader {
		agent.sendHeartBeat()
	}
	agent.resetTimeout()
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
		Term:         agent.currentTerm,
		CandidateId:  agent.id,
		LastLogIndex: agent.log.getLastLogIndex(),
		LastLogTerm:  agent.log.getLastLogTerm(),
	}
	for _, otherAgent := range agent.agentRPCs {
		otherAgent.requestVote(voteRequest)
	}
}

func (agent *Agent) handleRequestVoteResponse(response VoteResponse) {
	agent.logEvent("act=handleRequestVoteResponse response=%+v\n", response)
	if response.VotedFor {
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
		agent.resetTimeout()
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
		request := agent.createAppendEntriesRequest([]*LogEntry{}, nextIndex)
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
		agent.logEvent("act=sendAppendEntries request=%+v", request)
		otherAgent.appendEntries(request)
	}
}

func (agent *Agent) createAppendEntriesRequest(entries []*LogEntry, nextIndex int64) AppendEntriesRequest {
	var prevLogIndex int64 = 0
	if nextIndex > 0 {
		prevLogIndex = nextIndex - 1
	}
	agent.logEvent("act=sendHeartBeat prevLogIndex=%d logsize=%d", prevLogIndex, agent.log.length())
	prevLogTerm := agent.log.entries[prevLogIndex].Term
	return AppendEntriesRequest{
		Term:         agent.currentTerm,
		LeaderId:     agent.id,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		LeaderCommit: agent.commitIndex,
		Entries:      entries,
	}
}

func (agent *Agent) handleRequestVoteRPC(request VoteRequest) VoteResponse {
	agent.logEvent("act=handleRequestVoteRPC request=%+v", request)
	if request.Term < agent.currentTerm {
		return VoteResponse{
			Term:     agent.currentTerm,
			VotedFor: false,
			Id:       agent.id,
		}
	}
	agent.updateTerm(request.Term)
	if agent.votedFor == -1 &&
		request.LastLogTerm >= agent.log.getLastLogTerm() &&
		request.LastLogIndex >= agent.log.getLastLogIndex() {

		agent.logEvent("act=handleRequestVoteRPC request=%+v granting vote", request)
		agent.resetTimeout()
		agent.votedFor = request.CandidateId
		return VoteResponse{
			Term:     agent.currentTerm,
			VotedFor: true,
			Id:       agent.id,
		}
	}
	return VoteResponse{
		Term:     agent.currentTerm,
		VotedFor: false,
		Id:       agent.id,
	}
}

func (agent *Agent) handleAppendEntriesRPC(request AppendEntriesRequest) AppendEntriesResponse {
	agent.logEvent("act=handleAppendEntries request=%+v", request)
	agent.resetTimeout()
	agent.updateTerm(request.Term)

	if request.Term < agent.currentTerm {
		return AppendEntriesResponse{
			Term:      agent.currentTerm,
			Success:   false,
			Id:        agent.id,
			NextIndex: -1,
		}
	}
	if agent.log.getLastLogIndex() < request.PrevLogIndex ||
		agent.log.entries[request.PrevLogIndex].Term != request.PrevLogTerm {
		return AppendEntriesResponse{
			Term:      agent.currentTerm,
			Success:   false,
			Id:        agent.id,
			NextIndex: -1,
		}
	}

	//Else its good to append the logs
	agent.logEvent("act=handleAppendEntries request=%+v\n Appending Entries", request)

	agent.log.addEntriesToLog(request.PrevLogIndex, request.Entries)
	agent.followerUpdateCommitIndex(request.LeaderCommit)
	agent.logEvent("act=handleAppendEntriesComplete log=%+v", agent.log)
	return AppendEntriesResponse{
		Term:      agent.currentTerm,
		Success:   true,
		Id:        agent.id,
		NextIndex: agent.log.length(),
	}
}

func (agent *Agent) handleAppendEntriesResponse(response AppendEntriesResponse) {
	agent.logEvent("act=handleAppendEntriesResponse response=%+v", response)
	if response.Success {
		agent.nextIndex[response.Id] = response.NextIndex
		majorityThreshold := int(math.Ceil(float64(agent.numAgents()) / float64(2)))
		replicated := 1
		for _, nextIndex := range agent.nextIndex {
			if nextIndex >= response.NextIndex {
				replicated++
			}
		}
		if replicated >= majorityThreshold {
			agent.commit(response.NextIndex)
		}
	} else {
		agent.nextIndex[response.Id]--
	}
}

func (agent *Agent) commit(commitUpto int64) {
	if agent.log.getTerm(commitUpto-1) != agent.currentTerm {
		return
	}
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

func (agent *Agent) apply(index int64) bool {
	return true
}

func (agent *Agent) updateTerm(newTerm int64) {
	if newTerm > agent.currentTerm {
		agent.state = follower
		agent.votedFor = -1
		agent.currentTerm = newTerm
		agent.resetTimeout()
	}
}

func (agent *Agent) followerUpdateCommitIndex(leaderCommit int64) {
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
		duration = 20
	}
	agent.logEvent("act=resetTimeout duration=%d", duration)
	agent.timeout = duration
}

func generateTimeoutDuration() int64 {
	return rand.Int63n(150) + 150
}

func (agent *Agent) numAgents() int64 {
	return int64(len(agent.agentRPCs) + 1)
}

//AddCallback : use this to add to the callbacks slice for an agent to communicate
func (agent *Agent) AddCallback(rpc AgentRPC) {
	agent.agentRPCs = append(agent.agentRPCs, rpc)
}

func (agent *Agent) logEvent(format string, args ...interface{}) {
	log.Printf("id=%d state=%s currentTerm=%d %s", agent.id, agent.state, agent.currentTerm, fmt.Sprintf(format, args...))
}

func (agent *Agent) IsLeader() bool {
	return agent.state == leader
}
