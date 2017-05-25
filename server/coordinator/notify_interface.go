package coordinator

// NotifyInterface is the interface for notifying of new snapshots
type NotifyInterface interface {
	Notify(args *NotifyArgs, reply *NotifyReply) error
}
