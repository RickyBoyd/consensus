package raft

import "testing"

type MockCallback struct {
	// Returns for requestVote
	term        int
	voteGranted bool
	// Returns for appendEntries
	success bool
}

func newMock(term int, voteGranted bool, success bool) MockCallback {
	return MockCallback{
		term:        term,
		voteGranted: voteGranted,
		success:     success,
	}
}

func (a MockCallback) requestVote(term int,
	candidateID int,
	lastLogIndex int,
	lastLogTerm int) RequestVoteResponse {

	termResponse := make(chan int, 1)
	termResponse <- a.term

	voteResponse := make(chan bool, 1)
	voteResponse <- a.voteGranted

	return RequestVoteResponse{
		term:     termResponse,
		votedFor: voteResponse,
	}
}

func (a MockCallback) appendEntries(term int,
	leaderID int,
	prevLogIndex int,
	prevLogTerm int,
	entries []LogEntry,
	leaderCommit int) (int, bool) {

	return a.term, a.success
}

func TestRequestVotes(t *testing.T) {
	testAgent := NewAgent()
	testAgent.AddCallback(newMock(0, true, false))
	testAgent.AddCallback(newMock(0, false, false))

	votes := testAgent.requestVotes()
	count := testAgent.countVotes(votes)
	if count != 2 {
		t.Error("Expected 1 vote got ", votes)
	}
}
