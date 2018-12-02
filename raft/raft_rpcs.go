package raft

//VoteRequest type for a vote request
type VoteRequest struct {
	term         int
	candidateID  int
	lastLogIndex int
	lastLogTerm  int
}

//VoteResponse type
type VoteResponse struct {
	term     int
	votedFor bool
}

//AppendEntriesRequest type for AppendEntriesRPC
type AppendEntriesRequest struct {
	term         int
	leaderID     int
	prevLogIndex int
	prevLogTerm  int
	entries      []LogEntry
	leaderCommit int
}

//AppendEntriesResponse type for response
type AppendEntriesResponse struct {
	term    int
	success bool
}

//AgentCallback interface defines callbacks
type AgentCallback interface {
	requestVote(vote VoteRequest)
	appendEntries(request AppendEntriesRequest)
}
