package paxos

import (
	"fmt"
)

//Agent type
type Agent struct {
	ID           uint
	NumAgents    uint
	numProposals uint
}

// Proposal Type
type Proposal struct {
	ID   uint
	Data int
}

func (agent *Agent) accept(proposal Proposal) {

}

func (agent *Agent) generateProposalID() (id uint) {
	id = agent.NumAgents*agent.numProposals + agent.ID
	agent.numProposals++
	return id
}

func (agent *Agent) propose() {
	id := agent.generateProposalID()
	fmt.Printf("%d", id)

}
