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

func (state agentState) String() string {
	names := [...]string{
		"leader",
		"candidate",
		"follower"}

	return names[state]
}

//RequestVoteResponse type
type RequestVoteResponse struct {
	term     chan int
	votedFor chan bool
}

func (res *RequestVoteResponse) getTerm() int {
	return <-res.term
}

func (res *RequestVoteResponse) getVotedFor() bool {
	return <-res.votedFor
}

//AgentCallback interface defines callbacks
type AgentCallback interface {
	requestVote(term int, candidateID int, lastLogIndex int, lastLogTerm int) RequestVoteResponse
	appendEntries(term int, leaderID int, prevLogIndex int,
		prevLogTerm int, entries []LogEntry, leaderCommit int) (int, bool)
}

//LogEntry type
type LogEntry struct {
	Term    int
	Command int
}

//Agent type
type Agent struct {
	ID             int
	agentCallbacks []AgentCallback
	state          agentState
	timeout        *time.Timer
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
}

//NewAgent creates a new agent
func NewAgent() *Agent {
	agent := Agent{
		ID:          0,
		state:       follower,
		currentTerm: 0,
		votedFor:    -1,
		log:         make([]LogEntry, 0),
	}
	//agent.startTimeout()
	return &agent
}

func generateTimeoutDuration() time.Duration {
	return time.Duration(rand.Int63n(150) + 150)
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

//AddCallback : use this to add to the callbacks slice for an agent to communicate
func (agent *Agent) AddCallback(a AgentCallback) {
	agent.agentCallbacks = append(agent.agentCallbacks, a)
}

func (agent *Agent) requestVotes() []RequestVoteResponse {
	responses := make([]RequestVoteResponse, len(agent.agentCallbacks))
	for i, otherAgent := range agent.agentCallbacks {
		responses[i] = otherAgent.requestVote(agent.currentTerm, agent.ID, agent.getLastLogIndex(), agent.getLastLogTerm())
	}
	return responses
}

func (agent *Agent) countVotes(votes []RequestVoteResponse) int {
	numVotes := 1 //Vote for self
	for _, vote := range votes {
		votedForMe := vote.getVotedFor()
		term := vote.getTerm()
		if term > agent.currentTerm {
			// Then give up on election?
		}
		if votedForMe {
			numVotes++
		}
	}
	return numVotes
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

func (agent *Agent) sendHeartBeat() {

}

func (agent *Agent) beginElection() {
	agent.state = candidate
	agent.currentTerm++
	fmt.Printf("BEGIN ELECTION\n")
	//vote for self
	agent.votedFor = agent.ID
	//issue vote requests
	votes := agent.requestVotes()
	numVotes := agent.countVotes(votes)
	if numVotes > agent.numAgents()/2 {
		agent.state = leader
		//start some leader hearthbeat thing
	}
}

func (agent *Agent) requestVoteResponse(term int,
	candidateID int,
	lastLogIndex int,
	lastLogTerm int) (int, bool) {

	if term < agent.currentTerm {
		return agent.currentTerm, false
	}
	if agent.votedFor != -1 {
		// Now check candidates log is atleast
		// up to do as receivers log then grant vote
	}
	//TODO implement
	return 0, false
}

func (agent *Agent) numAgents() int {
	return len(agent.agentCallbacks)
}
