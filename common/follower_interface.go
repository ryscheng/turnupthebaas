package common

// FollowerInterface dictates the methods used for server-server communication
// in the Talek system
type FollowerInterface interface {
	GetName(args *interface{}, reply *string) error
	Write(args *WriteArgs, reply *WriteReply) error
	NextEpoch(args *uint64, reply *interface{}) error
	BatchRead(args *BatchReadRequest, reply *BatchReadReply) error
	GetUpdates(args *GetUpdatesArgs, reply *GetUpdatesReply) error
}
