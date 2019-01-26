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

func assertEqual(t *testing.T, expected interface{}, actual interface{}, message string) bool {
	if expected == actual {
		return true
	}
	message = fmt.Sprintf("Expected: %v != Actual: %v\n%s", expected, actual, message)
	t.Fatal(message)
	return true
}

func assertLogEntryEqual(t *testing.T, expected *LogEntry, actual *LogEntry, message string) {
	if expected.Term == actual.Term &&
		expected.Command == actual.Command {
		return
	}
	message = fmt.Sprintf("Expected: %v != Actual: %v\n%s", expected, actual, message)
	t.Fatal(message)
}

func assertAppendEntriesRequestEqual(t *testing.T, expected AppendEntriesRequest, actual AppendEntriesRequest, message string) {
	if expected.Term == actual.Term &&
		expected.LeaderId == actual.LeaderId &&
		expected.PrevLogIndex == actual.PrevLogIndex &&
		expected.PrevLogTerm == actual.PrevLogTerm &&
		expected.LeaderCommit == actual.LeaderCommit {
		for ii := 0; ii < len(expected.Entries); ii++ {
			assertLogEntryEqual(t, expected.Entries[ii], actual.Entries[ii], message)
		}
		return
	}
	message = fmt.Sprintf("Expected: %v != Actual: %v\n%s", expected, actual, message)
	t.Fatal(message)
}

func assertAppendEntriesResponseEqual(t *testing.T, expected AppendEntriesResponse, actual AppendEntriesResponse, message string) {
	if expected.Term == actual.Term &&
		expected.Success == actual.Success &&
		expected.Id == actual.Id &&
		expected.NextIndex == actual.NextIndex {
		return
	}
	message = fmt.Sprintf("Expected: %v != Actual: %v\n%s", expected, actual, message)
	t.Fatal(message)
}

func generateTestLog(commandsPerTerm []int64) AgentLog {
	log := make([]*LogEntry, 0)
	for term, numCommands := range commandsPerTerm {
		for command := int64(0); command < numCommands; command++ {
			entry := &LogEntry{
				Term:    int64(term),
				Command: int64(command),
			}
			log = append(log, entry)
		}
	}
	return AgentLog{log}
}

func addLogTest(t *testing.T, commandsPerTerm []int64, entry *LogEntry, index int64, expectedLength int64) {
	//Given
	logs := generateTestLog([]int64{3, 4})
	agent := Agent{log: logs}
	//When
	agent.log.addToLog(index, entry)
	//Then
	assertEqual(t, expectedLength, agent.log.length(), "")
	assertLogEntryEqual(t, agent.log.entries[index], entry, "")
}
func TestAddLogAppends(t *testing.T) {
	entry := &LogEntry{
		Term:    3,
		Command: 1,
	}
	addLogTest(t, []int64{3, 4}, entry, 7, 8)
}

func TestAddLogOverwritesAtEnd(t *testing.T) {
	entry := &LogEntry{
		Term:    3,
		Command: 1,
	}
	addLogTest(t, []int64{3, 4}, entry, 6, 7)
}
func TestAddLogOverwrites(t *testing.T) {
	entry := &LogEntry{
		Term:    3,
		Command: 1,
	}
	addLogTest(t, []int64{3, 4}, entry, 3, 7)
}

func TestAddEntriesToLogDeletes(t *testing.T) {
	//Given
	agent := NewAgent(0)
	agent.log = generateTestLog([]int64{1, 2})
	//When
	entry := &LogEntry{
		Term:    0,
		Command: 0,
	}
	agent.log.addEntriesToLog(0, []*LogEntry{entry})
	//Then
	assertEqual(t, int64(2), agent.log.length(), "")
	assertLogEntryEqual(t, entry, agent.log.entries[1], "")
}
func TestAddEntriesToLogAppends(t *testing.T) {
	//Given
	agent := NewAgent(0)
	agent.log = generateTestLog([]int64{1, 2})
	//When
	entry := &LogEntry{
		Term:    0,
		Command: 0,
	}
	agent.log.addEntriesToLog(2, []*LogEntry{entry})
	//Then
	assertEqual(t, int64(4), agent.log.length(), "")
	assertLogEntryEqual(t, entry, agent.log.entries[3], "")
}

func TestAddEntriesToLogOverwritesAndAppends(t *testing.T) {
	//Given
	agent := NewAgent(0)
	agent.log = generateTestLog([]int64{1, 2})
	//When
	entry0 := &LogEntry{
		Term:    0,
		Command: 0,
	}
	entry1 := &LogEntry{
		Term:    0,
		Command: 0,
	}
	agent.log.addEntriesToLog(1, []*LogEntry{entry0, entry1})
	//Then
	assertEqual(t, int64(4), agent.log.length(), "")
	assertLogEntryEqual(t, entry0, agent.log.entries[2], "")
	assertLogEntryEqual(t, entry1, agent.log.entries[3], "")
}

func TestNumAgentsIsOne(t *testing.T) {
	//Given
	agent := NewAgent(0)
	//When
	//Then
	assertEqual(t, int64(1), agent.numAgents(), "")
}
func TestNumAgents(t *testing.T) {
	//Given
	agent := NewAgent(0)
	//When
	agent.AddCallback(AgentChannelRPC{})
	//Then
	assertEqual(t, int64(2), agent.numAgents(), "")
}

func TestInitializeNextIndex(t *testing.T) {
	//Given
	agent := NewAgent(0)
	agent.log = generateTestLog([]int64{1, 2})
	//When
	agent.initialiseNextIndex()
	//Then
	for _, index := range agent.nextIndex {
		assertEqual(t, int64(3), index, "")
	}
}

func TestHandleAppendEntries(t *testing.T) {
	//Given
	agent := NewAgent(0)
	newLog := &LogEntry{
		Term:    1,
		Command: 1,
	}
	request := AppendEntriesRequest{
		Term:         1,
		LeaderId:     1,
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		LeaderCommit: 0,
		Entries:      []*LogEntry{newLog},
	}
	//When
	response := agent.handleAppendEntriesRPC(request)

	//Then
	assertEqual(t, int64(2), agent.log.length(), "")
	assertEqual(t, newLog, agent.log.entries[1], "")

	expectedResponse := AppendEntriesResponse{
		Term:      1,
		Success:   true,
		Id:        0,
		NextIndex: 2,
	}
	assertAppendEntriesResponseEqual(t, expectedResponse, response, "")

	assertEqual(t, int64(0), agent.commitIndex, "")
}

func TestHandleAppendEntriesOverwrites(t *testing.T) {
	//Given
	agent := NewAgent(0)
	newLog := &LogEntry{
		Term:    1,
		Command: 1,
	}
	request := AppendEntriesRequest{
		Term:         1,
		LeaderId:     1,
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		LeaderCommit: 0,
		Entries:      []*LogEntry{newLog},
	}

	overwritingLog := &LogEntry{
		Term:    2,
		Command: 2,
	}
	overwritingRequest := AppendEntriesRequest{
		Term:         1,
		LeaderId:     1,
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		LeaderCommit: 0,
		Entries:      []*LogEntry{overwritingLog},
	}

	//When
	response1 := agent.handleAppendEntriesRPC(request)
	response2 := agent.handleAppendEntriesRPC(overwritingRequest)

	//Then
	assertEqual(t, int64(2), agent.log.length(), "")
	assertEqual(t, overwritingLog, agent.log.entries[1], "")

	expectedResponse := AppendEntriesResponse{
		Term:      1,
		Success:   true,
		Id:        0,
		NextIndex: 2,
	}
	assertAppendEntriesResponseEqual(t, expectedResponse, response1, "")
	assertAppendEntriesResponseEqual(t, expectedResponse, response2, "")

	assertEqual(t, int64(0), agent.commitIndex, "")
}

func TestHandleAppendEntriesMultipleLogs(t *testing.T) {
	//Given
	agent := NewAgent(0)
	newEntries := []*LogEntry{&LogEntry{Term: 1, Command: 1}, &LogEntry{Term: 1, Command: 2}}
	request := AppendEntriesRequest{
		Term:         1,
		LeaderId:     1,
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		LeaderCommit: 3,
		Entries:      newEntries,
	}
	//When
	response := agent.handleAppendEntriesRPC(request)

	//Then
	assertEqual(t, int64(3), agent.log.length(), "")
	assertEqual(t, newEntries[0], agent.log.entries[1], "")
	assertEqual(t, newEntries[1], agent.log.entries[2], "")

	assertAppendEntriesResponseEqual(t, AppendEntriesResponse{Term: 1, Success: true, Id: 0, NextIndex: 3}, response, "")

	// Test that commitIndex = min(eaderCommit, lastIndex)
	assertEqual(t, int64(2), agent.commitIndex, "")
}

func TestHandleAppendEntriesMultipleRequests(t *testing.T) {
	//Given
	agent := NewAgent(0)
	newEntries := []*LogEntry{&LogEntry{Term: 1, Command: 1}, &LogEntry{Term: 1, Command: 2}}
	request1 := AppendEntriesRequest{Term: 1, LeaderId: 1, PrevLogIndex: 0, PrevLogTerm: 0, LeaderCommit: 0, Entries: newEntries[:1]}
	request2 := AppendEntriesRequest{Term: 1, LeaderId: 1, PrevLogIndex: 1, PrevLogTerm: 1, LeaderCommit: 1, Entries: newEntries[1:]}
	//When
	response1 := agent.handleAppendEntriesRPC(request1)
	response2 := agent.handleAppendEntriesRPC(request2)

	//Then
	assertEqual(t, int64(3), agent.log.length(), "")
	assertEqual(t, newEntries[0], agent.log.entries[1], "")
	assertEqual(t, newEntries[1], agent.log.entries[2], "")

	assertAppendEntriesResponseEqual(t, AppendEntriesResponse{Term: 1, Success: true, Id: 0, NextIndex: 2}, response1, "")
	assertAppendEntriesResponseEqual(t, AppendEntriesResponse{Term: 1, Success: true, Id: 0, NextIndex: 3}, response2, "")

	assertEqual(t, int64(1), agent.commitIndex, "")
}

func TestHandleAppendEntriesRejectsBadIndex(t *testing.T) {
	//Given
	agent := NewAgent(0)
	newEntries := []*LogEntry{&LogEntry{Term: 1, Command: 1}, &LogEntry{Term: 1, Command: 2}}
	request := AppendEntriesRequest{Term: 1, LeaderId: 1, PrevLogIndex: 1, PrevLogTerm: 0, LeaderCommit: 0, Entries: newEntries}
	//When
	agent.handleAppendEntriesRPC(request)

	//Then
	assertEqual(t, int64(1), agent.log.length(), "")
	assertLogEntryEqual(t, &LogEntry{Term: 0, Command: 0}, agent.log.entries[0], "")

	assertEqual(t, int64(0), agent.commitIndex, "")
}
func TestHandleAppendEntriesRejectsBadTerm(t *testing.T) {
	//Given
	agent := NewAgent(0)
	newEntries := []*LogEntry{&LogEntry{Term: 1, Command: 1}, &LogEntry{Term: 1, Command: 2}}
	request := AppendEntriesRequest{Term: 1, LeaderId: 1, PrevLogIndex: 0, PrevLogTerm: 2, LeaderCommit: 0, Entries: newEntries}
	//When
	agent.handleAppendEntriesRPC(request)

	//Then
	assertEqual(t, int64(1), agent.log.length(), "")
	assertLogEntryEqual(t, &LogEntry{Term: 0, Command: 0}, agent.log.entries[0], "")

	assertEqual(t, int64(0), agent.commitIndex, "")
}
