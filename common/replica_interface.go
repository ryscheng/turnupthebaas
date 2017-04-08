package common

// ReplicaInterface dictates the methods used for server-server communication
// in the Talek system
type ReplicaInterface interface {
	Write(args *ReplicaWriteArgs, reply *ReplicaWriteReply) error
	BatchRead(args *BatchReadRequest, reply *BatchReadReply) error
	GetUpdates(args *GetUpdatesArgs, reply *GetUpdatesReply) error
}
