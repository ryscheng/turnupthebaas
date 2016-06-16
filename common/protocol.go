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
