package main

import (
	"consensus/raft"
	"log"
	"os"
)

func main() {
	f, err := os.OpenFile("info.log", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Main")

	as := raft.ConstructRaftChanInstance(3)

	for j := 0; j < 750; j++ {
		for ii := 0; ii < len(as); ii++ {
			if j == 500 {
				if as[ii].Agent.IsLeader() {
					as[ii].Agent.ClientAction(2)
				}
			}
			if j == 600 {
				if as[ii].Agent.IsLeader() {
					as[ii].Agent.ClientAction(3)
				}
			}
			as[ii].Tick()
		}
	}
}
