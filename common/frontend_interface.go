package common

// FrontendInterface is the interface between libtalek and the frontend
type FrontendInterface interface {
	GetName(args *interface{}, reply *string) error
	GetConfig(args *interface{}, reply *Config) error
	Write(args *WriteArgs, reply *WriteReply) error
	Read(args *EncodedReadArgs, reply *ReadReply) error
	GetUpdates(args *GetUpdatesArgs, reply *GetUpdatesReply) error
}
