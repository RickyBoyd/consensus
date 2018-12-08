package raft

import (
	"fmt"
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

func assertEqual(t *testing.T, expected interface{}, actual interface{}, message string) {
	if expected == actual {
		return
	}
	message = fmt.Sprintf("Expected: %v != Actual: %v\n%s", expected, actual, message)
	t.Fatal(message)
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
	assertEqual(t, expectedLength, len(agent.log), "")
	assertEqual(t, agent.log[index], entry, "")
}
func TestAddLogAppends(t *testing.T) {
	entry := LogEntry{3, 1}
	addLogTest(t, []int{3, 4}, entry, 7, 8)
}

func TestAddLogOverwritesAtEnd(t *testing.T) {
	entry := LogEntry{3, 1}
	addLogTest(t, []int{3, 4}, entry, 6, 7)
}
func TestAddLogOverwrites(t *testing.T) {
	entry := LogEntry{3, 1}
	addLogTest(t, []int{3, 4}, entry, 3, 7)
}

func TestNumAgentsIsOne(t *testing.T) {
	agent := NewAgent(0)
	assertEqual(t, 1, agent.numAgents(), "")
}
func TestNumAgents(t *testing.T) {
	agent := NewAgent(0)
	agent.AddCallback(AgentChannelRPC{})
	assertEqual(t, 2, agent.numAgents(), "")
}
