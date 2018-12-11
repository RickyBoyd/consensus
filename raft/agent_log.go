package raft

type AgentLog struct {
	entries []LogEntry
}

func (log *AgentLog) addEntriesToLog(prevLogIndex int, entries []LogEntry) {
	index := prevLogIndex + 1
	for _, entry := range entries {
		log.addToLog(index, entry)
		index++
	}
	log.entries = log.entries[:index]
}

func (log *AgentLog) addToLog(index int, entry LogEntry) {
	if index >= log.length() {
		log.entries = append(log.entries, entry)
	} else {
		log.entries[index] = entry
	}
}

func (log *AgentLog) getLastLogIndex() int {
	return len(log.entries) - 1
}

func (log *AgentLog) getLastLogTerm() int {
	logLen := log.length()
	if logLen == 0 {
		return 0
	}
	return log.entries[logLen-1].Term
}

func (log *AgentLog) length() int {
	return len(log.entries)
}
