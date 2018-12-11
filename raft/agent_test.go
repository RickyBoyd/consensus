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
	//Given
	logs := generateTestLog([]int{3, 4})
	agent := Agent{log: logs}
	//When
	agent.addToLog(index, entry)
	//Then
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

func TestAddEntriesToLogDeletes(t *testing.T) {
	//Given
	agent := NewAgent(0)
	agent.log = generateTestLog([]int{1, 2})
	//When
	agent.addEntriesToLog(0, []LogEntry{LogEntry{0, 0}})
	//Then
	assertEqual(t, 2, len(agent.log), "")
	assertEqual(t, LogEntry{0, 0}, agent.log[1], "")
}
func TestAddEntriesToLogAppends(t *testing.T) {
	//Given
	agent := NewAgent(0)
	agent.log = generateTestLog([]int{1, 2})
	//When
	agent.addEntriesToLog(2, []LogEntry{LogEntry{0, 0}})
	//Then
	assertEqual(t, 4, len(agent.log), "")
	assertEqual(t, LogEntry{0, 0}, agent.log[3], "")
}

func TestAddEntriesToLogOverwritesAndAppends(t *testing.T) {
	//Given
	agent := NewAgent(0)
	agent.log = generateTestLog([]int{1, 2})
	//When
	agent.addEntriesToLog(1, []LogEntry{LogEntry{0, 0}, LogEntry{0, 1}})
	//Then
	assertEqual(t, 4, len(agent.log), "")
	assertEqual(t, LogEntry{0, 0}, agent.log[2], "")
	assertEqual(t, LogEntry{0, 1}, agent.log[3], "")
}

func TestNumAgentsIsOne(t *testing.T) {
	//Given
	agent := NewAgent(0)
	//When
	//Then
	assertEqual(t, 1, agent.numAgents(), "")
}
func TestNumAgents(t *testing.T) {
	//Given
	agent := NewAgent(0)
	//When
	agent.AddCallback(AgentChannelRPC{})
	//Then
	assertEqual(t, 2, agent.numAgents(), "")
}

func TestInitializeNextIndex(t *testing.T) {
	//Given
	agent := NewAgent(0)
	agent.log = generateTestLog([]int{1, 2})
	//When
	agent.initialiseNextIndex()
	//Then
	for _, index := range agent.nextIndex {
		assertEqual(t, 3, index, "")
	}
}

func TestHandleAppendEntries(t *testing.T) {
	//Given
	agent := NewAgent(0)
	newLog := LogEntry{1, 1}
	request := AppendEntriesRequest{1, 1, 0, 0, []LogEntry{newLog}, 0}
	//When
	response := agent.handleAppendEntriesRPC(request)

	//Then
	assertEqual(t, 2, len(agent.log), "")
	assertEqual(t, newLog, agent.log[1], "")

	assertEqual(t, AppendEntriesResponse{1, true, 0}, response, "")
}

func TestHandleAppendEntriesMultipleLogs(t *testing.T) {
	//Given
	agent := NewAgent(0)
	newEntries := []LogEntry{LogEntry{1, 1}, LogEntry{1, 2}}
	request := AppendEntriesRequest{1, 1, 0, 0, newEntries, 0}
	//When
	response := agent.handleAppendEntriesRPC(request)

	//Then
	assertEqual(t, 3, len(agent.log), "")
	assertEqual(t, newEntries[0], agent.log[1], "")
	assertEqual(t, newEntries[1], agent.log[2], "")

	assertEqual(t, AppendEntriesResponse{1, true, 0}, response, "")
}

func TestHandleAppendEntriesMultipleRequests(t *testing.T) {
	//Given
	agent := NewAgent(0)
	newEntries := []LogEntry{LogEntry{1, 1}, LogEntry{1, 2}}
	request1 := AppendEntriesRequest{1, 1, 0, 0, newEntries[:1], 0}
	request2 := AppendEntriesRequest{1, 1, 1, 1, newEntries[1:], 0}
	//When
	response1 := agent.handleAppendEntriesRPC(request1)
	response2 := agent.handleAppendEntriesRPC(request2)

	//Then
	assertEqual(t, 3, len(agent.log), "")
	assertEqual(t, newEntries[0], agent.log[1], "")
	assertEqual(t, newEntries[1], agent.log[2], "")

	assertEqual(t, AppendEntriesResponse{1, true, 0}, response1, "")
	assertEqual(t, AppendEntriesResponse{1, true, 0}, response2, "")
}
