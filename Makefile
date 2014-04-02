all:
	go install github.com/nilangshah/Raft
	go test
	rm -rf Raftlog*
	rm -rf /home/nilang/Documents/go-space/go/src/github.com/nilangshah/Raft/log1 /home/nilang/Documents/go-space/go/src/github.com/nilangshah/Raft/log2 /home/nilang/Documents/go-space/go/src/github.com/nilangshah/Raft/log3 /home/nilang/Documents/go-space/go/src/github.com/nilangshah/Raft/log4 /home/nilang/Documents/go-space/go/src/github.com/nilangshah/Raft/log5 
