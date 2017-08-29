package common

import (
	"bytes"
	"net/http"

	"github.com/gorilla/rpc/json"
)

// RPCCall Makes a JSON RPC client.
func RPCCall(client *http.Client, address string, methodName string, args interface{}, reply interface{}) error {
	var err error
	if client == nil {
		client = &http.Client{}
	}

	// Encode arguments
	message, err := json.EncodeClientRequest(methodName, args)
	if err != nil {
		return err
	}

	// Construct request
	req, err := http.NewRequest("POST", "http://"+address, bytes.NewBuffer(message))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Do RPC
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if err = json.DecodeClientResponse(resp.Body, reply); err != nil {
		return err
	}

	return nil
}
