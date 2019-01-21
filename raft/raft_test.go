package raft

import (
	"testing"
)

func TestAlwaysHasOneLeader(t *testing.T) {
	as := ConstructRaftChanInstance(3)

	for j := 0; j < 2000; j++ {
		numLeaders := 0
		for ii := 0; ii < len(as); ii++ {
			if as[ii].Agent.IsLeader() {
				numLeaders++
			}
			as[ii].Tick()
		}
		if j > 300 {
			assertEqual(t, 1, numLeaders, "Must have exactly one leader")
		}
	}
}

func TestAlwaysHasOneLeaderWhenElectionForced(t *testing.T) {
	as := ConstructRaftChanInstance(3)

	for j := 0; j < 2000; j++ {
		numLeaders := 0
		for ii := 0; ii < len(as); ii++ {
			if as[ii].Agent.IsLeader() {
				numLeaders++
				if j == 650 {
					as[ii].Agent.state = follower
				}
			}
			as[ii].Tick()
		}
		if j > 700 {
			assertEqual(t, 1, numLeaders, "Must have exactly one leader")
		}
	}
}

func TestLogEntryIsReplicated(t *testing.T) {
	as := ConstructRaftChanInstance(3)

	for ticks := 0; ticks < 2000; ticks++ {
		for ii := 0; ii < len(as); ii++ {
			if as[ii].Agent.IsLeader() {
				if ticks == 400 {
					as[ii].Agent.ClientAction(3)
				}
			}
			as[ii].Tick()
		}
	}
	assertEqual(t, LogEntry{1, 3}, as[0].Agent.log.entries[1], "Entry was not entered")
	assertEqual(t, LogEntry{1, 3}, as[1].Agent.log.entries[1], "Entry was not entered")
	assertEqual(t, LogEntry{1, 3}, as[2].Agent.log.entries[1], "Entry was not entered")

	assertEqual(t, 1, as[0].Agent.commitIndex, "Commit index id 0 did not get raised")
	assertEqual(t, 1, as[1].Agent.commitIndex, "Commit index id 1 did not get raised")
	assertEqual(t, 1, as[2].Agent.commitIndex, "Commit index id 2 did not get raised")
}
