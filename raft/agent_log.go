package raft

//AgentLog is the type for holding the agents log and related helper methods
type AgentLog struct {
	entries []*LogEntry
}

func newLog() AgentLog {
	entry := &LogEntry{
		Term:    0,
		Command: 0,
	}
	return AgentLog{
		entries: []*LogEntry{entry},
	}
}

func (log *AgentLog) appendToLog(entry *LogEntry) {
	log.entries = append(log.entries, entry)
}

func (log *AgentLog) addEntriesToLog(prevLogIndex int64, entries []*LogEntry) {
	index := prevLogIndex + 1
	for _, entry := range entries {
		log.addToLog(index, entry)
		index++
	}
	log.entries = log.entries[:index]
}

func (log *AgentLog) addToLog(index int64, entry *LogEntry) {
	if index >= log.length() {
		log.entries = append(log.entries, entry)
	} else {
		log.entries[index] = entry
	}
}

func (log *AgentLog) getLastLogIndex() int64 {
	return int64(len(log.entries) - 1)
}

func (log *AgentLog) getLastLogTerm() int64 {
	logLen := log.length()
	if logLen == 0 {
		return 0
	}
	return log.entries[logLen-1].Term
}

func (log *AgentLog) getEntries(index int64) []*LogEntry {
	if index < log.length() {
		return log.entries[index:]
	}
	return []*LogEntry{}
}

func (log *AgentLog) getTerm(index int64) int64 {
	return log.entries[index].Term
}

func (log *AgentLog) length() int64 {
	return int64(len(log.entries))
}
