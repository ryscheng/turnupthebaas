package common

type FollowerInterface interface {
	GetName() string
	Ping(args *PingArgs, reply *PingReply) error
	Write(args *WriteArgs, reply *WriteReply) error
	BatchRead(args *BatchReadRequest, reply *BatchReadReply) error
	GetUpdates(args *GetUpdatesArgs, reply *GetUpdatesReply) error
}
