package prototest

//go:generate protoc -I=$GOPATH/src/  --go_out=$GOPATH/src/ github.com/philpearl/plenc/prototest/test.proto
