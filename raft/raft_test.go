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
