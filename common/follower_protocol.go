package common

type BatchReadArgs struct {
	Args []ReadArgs // Set of Read requests
}

type BatchReadReply struct {
	Err     string
	Replies []ReadReply
}
