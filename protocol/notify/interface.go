package notify

// Interface is the interface for notifying of new snapshots
type Interface interface {
	Notify(args *Args, reply *Reply) error
}
