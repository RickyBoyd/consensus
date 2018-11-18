package paxos

import "testing"

func TestGenerateProposalID(t *testing.T) {
	agent := Agent{ID: 3, NumAgents: 4}
	proposal1 := agent.generateProposalID()
	proposal2 := agent.generateProposalID()
	if proposal1 != 3 {
		t.Error("Expected 3, got ", proposal1)
	}
	if proposal2 != 7 {
		t.Error("Expected 7, got ", proposal2)
	}
}
