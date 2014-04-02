all:
	go install github.com/nilangshah/Raft
	go test
	rm -rf Raftlog*
	
