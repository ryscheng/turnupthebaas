package intvec

// Interface is the interface for getting interest vectors
type Interface interface {
	GetIntVec(args *Args, reply *Reply) error
}
