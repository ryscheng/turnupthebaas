package common

type LeaderInterface interface {
	GetName(args *interface{}, reply *string) error
	Ping(args *PingArgs, reply *PingReply) error
	Write(args *WriteArgs, reply *WriteReply) error
	Read(args *EncodedReadArgs, reply *ReadReply) error
	GetUpdates(args *GetUpdatesArgs, reply *GetUpdatesReply) error
}
