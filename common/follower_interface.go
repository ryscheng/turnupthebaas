package common

type FollowerInterface interface {
	GetName(args *interface{}, reply *string) error
	Ping(args *PingArgs, reply *PingReply) error
	Write(args *WriteArgs, reply *WriteReply) error
	BatchRead(args *BatchReadRequest, reply *BatchReadReply) error
	GetUpdates(args *GetUpdatesArgs, reply *GetUpdatesReply) error
}
