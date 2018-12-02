package raft

import (
	"testing"
)

type MockCallback struct {
	// Returns for requestVote
	term        int
	voteGranted bool
	// Returns for appendEntries
	success bool
}

func newVoterMock(term int, voteGranted bool, success bool) MockCallback {
	return MockCallback{
		term:        term,
		voteGranted: voteGranted,
		success:     success,
	}
}

func (a MockCallback) requestVote(request VoteRequest) {
}

func (a MockCallback) appendEntries(request AppendEntriesRequest) {
}

func generateTestLog(commandsPerTerm []int) []LogEntry {
	log := make([]LogEntry, 0)
	for term, numCommands := range commandsPerTerm {
		for command := 0; command < numCommands; command++ {
			log = append(log, LogEntry{command, term})
		}
	}
	return log
}

func addLogTest(t *testing.T, commandsPerTerm []int, entry LogEntry, index int, expectedLength int) {
	logs := generateTestLog([]int{3, 4})
	agent := Agent{log: logs}
	agent.addToLog(index, entry)

	if len(agent.log) != expectedLength {
		t.Error("Expected length: ", expectedLength, ", actual: ", len(agent.log))
	}
	if agent.log[index] != entry {
		t.Error("Expected: ", entry, " Actual: ", agent.log[index])
	}
}
func TestAddLogAppends(t *testing.T) {
	entry := LogEntry{3, 5}
	addLogTest(t, []int{3, 4}, entry, 7, 8)
}

func TestAddLogOverwritesAtEnd(t *testing.T) {
	entry := LogEntry{3, 5}
	addLogTest(t, []int{3, 4}, entry, 6, 7)
}
func TestAddLogOverwrites(t *testing.T) {
	entry := LogEntry{3, 5}
	addLogTest(t, []int{3, 4}, entry, 3, 7)
}
