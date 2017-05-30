package layout

// Interface is the interface for getting layouts
type Interface interface {
	GetLayout(args *GetLayoutArgs, reply *GetLayoutReply) error
}
