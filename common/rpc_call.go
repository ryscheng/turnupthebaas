package common

import (
	"net/rpc"
)

// RPCCall generalizes an RPC call that caches a client
// If client is not specified (nil), it will try to dial a new client
// If a client is specified, it will reuse the same client
// Returns:
// - client if one exists, nil otherwise
// - nil on success, error on failure
func RPCCall(client *rpc.Client, addr string, methodName string, args interface{}, reply interface{}) (*rpc.Client, error) {
	// Get address
	var err error

	// Setup connection
	if client == nil {
		client, err = rpc.Dial("tcp", addr)
		if err != nil {
			//c.log.Error.Printf("rpc dialing failed: %v\n", err)
			return nil, err
		}
		//defer client.Close()
	}

	// Do RPC
	err = client.Call(methodName, args, reply)
	if err != nil {
		//c.log.Error.Printf("rpc error: %v", err)
		// @todo should we close the client and return a nil client?
		return client, err
	}

	//l.log.Printf("%s.Call(): %v, %v, %v\n", addr, args, reply)
	return client, nil
}
