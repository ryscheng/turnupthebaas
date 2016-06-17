package common

type PingArgs struct {
	Msg string
}

type PingReply struct {
	Err string
	Msg string
}

type AppendArgs struct {
	Bucket uint64
	Data   []byte
}

type AppendReply struct {
	Err string
}

type PirArgs struct {
	requestVector []byte
}

type PirReply struct {
	Err  string
	Data []byte
}
