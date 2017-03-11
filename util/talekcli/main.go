package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/libtalek"
)

var config = flag.String("config", "talek.json", "Talek configuration.")
var create = flag.Bool("create", false, "Create a new Talek log.")
var share = flag.String("share", "", "Generate a read-only version of a log.")
var log = flag.String("log", "topic.json", "The Talek log to act upon.")
var write = flag.String("write", "", "The message to write to the log.")
var read = flag.Bool("read", false, "Read from an existing talek log.")

// Talekcli provides a minimal command line client interface to talek for
// writing and reading individual items in the database. This interaction mode
// is not resistant to traffic analysis attacks, because a single request is
// made to perform the operations, rather than scheduling them within a
// consistant stream of requests to mask underlying activity.
func main() {
	flag.Parse()

	// Create a new log.
	if *create {
		handle, err := libtalek.NewTopic()
		if err != nil {
			fmt.Printf("Could not generate new topic: %v\n", err)
			return
		}
		if len(*log) == 0 {
			fmt.Printf("-create cannot be used without specifying a -log to save to.")
			return
		}
		handleraw, err := json.Marshal(handle)
		if err != nil {
			fmt.Printf("Could not flatten log: %v\n", err)
			return
		}
		err = ioutil.WriteFile(*log, handleraw, 0640)
		if err != nil {
			fmt.Printf("Failed to write log state: %v\n", err)
			return
		}
		fmt.Printf("New log created at %s\n", *log)
		return
	}

	if len(*share) > 0 {
		dat, readerr := ioutil.ReadFile(*log)
		if readerr != nil {
			fmt.Fprintf(os.Stderr, "Could not read %s: %v", *log, readerr)
			return
		}
		var handle libtalek.Topic
		err := json.Unmarshal(dat, &handle)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse %s: %v", *log, err)
			return
		}
		rodat, err := json.Marshal(&handle.Subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not serialize log: %v", err)
			return
		}
		err = ioutil.WriteFile(*share, rodat, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not write read-only log %s: %v", *share, err)
		}
		return
	}

	var conf libtalek.ClientConfig
	confraw, err := ioutil.ReadFile(*config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read config file: %v\n", err)
		return
	}
	err = json.Unmarshal(confraw, &conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse config file: %v\n", err)
		return
	}

	dat, readerr := ioutil.ReadFile(*log)
	if readerr != nil {
		fmt.Fprintf(os.Stderr, "Could not read %s: %v", *log, readerr)
		return
	}
	var handle libtalek.Topic
	err = json.Unmarshal(dat, &handle)
	if err != nil || handle.SharedSecret == nil {
		fmt.Fprintf(os.Stderr, "Could not parse log state: %v\n", err)
		return
	}

	leaderRPC := common.NewLeaderRpc("RPC", conf.TrustDomains[0])
	cli := libtalek.NewClient("talekcli", conf, leaderRPC)

	// Read a message.
	if *read {
		msgchan := cli.Poll(&handle.Subscription)
		msg := <-msgchan
		// print the result.
		fmt.Printf("%s\n", msg)
		cli.Done(&handle.Subscription)
	} else if len(*write) > 0 {
		if handle.SigningPrivateKey == nil {
			fmt.Fprintf(os.Stderr, "Cannot write. Only provided read capability for log.")
			return
		}
		cli.Publish(&handle, []byte(*write))
		time.Sleep(conf.WriteInterval)
	} else {
		fmt.Printf("No operation to perform.")
		return
	}

	cli.Kill()

	// write updated state back to log.
	rawstate, err := json.Marshal(&handle)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not serialize updated log: %v\n", err)
	}
	err = ioutil.WriteFile(*log, rawstate, 0640)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not update log: %v\n", err)
	}
}
