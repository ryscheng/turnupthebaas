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

const readTimeoutMultiple = 5

var config = flag.String("config", "talek.conf", "Client configuration for talek")
var create = flag.Bool("create", false, "Create a new talek handle")
var share = flag.String("share", "", "Create a read-only version of the topic for sharing")
var handle = flag.String("topic", "talek.handle", "The talek handle to use")
var write = flag.String("write", "", "A message to append to the log (If not specified, the next item will be read.)")
var read = flag.Bool("read", false, "Read from the provided topic")
var verbose = flag.Bool("verbose", false, "Print diagnostic information")

// The CLI client will read or write a single item for talek
func main() {
	flag.Parse()

	// Config
	config := libtalek.ClientConfigFromFile(*config)
	if config == nil {
		fmt.Fprintln(os.Stderr, "Talek Client must be run with --config specifying where the server is.")
		os.Exit(1)
	}
	if config.Config == nil && *verbose {
		fmt.Fprintln(os.Stderr, "Common configuration will be fetched from frontend.")
	}

	topicdata, err := ioutil.ReadFile(*handle)
	if err != nil {
		panic(err)
	}
	topic := libtalek.Topic{}
	if *create {
		nt, newerr := libtalek.NewTopic()
		if newerr != nil {
			panic(err)
		}
		topic = *nt
	} else {
		if err = json.Unmarshal(topicdata, &topic); err != nil {
			panic(err)
		}
	}

	if len(*share) > 0 {
		handle := topic.Handle
		handleBytes, handleerr := json.Marshal(handle)
		if handleerr != nil {
			panic(handleerr)
		}
		ioutil.WriteFile(*share, handleBytes, 0640)
		fmt.Fprintf(os.Stderr, "Read-only topic written to %s\n", *share)
	}

	// Client-connected activity below.
	leaderRPC := common.NewFrontendRPC("RPC", config.FrontendAddr)

	client := libtalek.NewClient("Client", *config, leaderRPC)
	if client == nil {
		panic("could not create talek client")
	} else if *verbose {
		fmt.Fprintln(os.Stderr, "Connection to RPC established.")
	}

	if *read == false && len(*write) > 0 {
		if err = client.Publish(&topic, []byte(*write)); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to publish: %s", err)
			panic(err)
		}
	} else if *read == false {
		fmt.Fprintf(os.Stderr, "No Read or Write operation requested. Closing.")
	} else {
		msgs := client.Poll(&topic.Handle)
		timeout := time.After(config.ReadInterval * readTimeoutMultiple)
		select {
		case next := <-msgs:
			fmt.Println(next)
			break
		case <-timeout:
			if *verbose {
				fmt.Fprintln(os.Stderr, "Timed out waiting for new message.")
			}
			return
		}
		client.Done(&topic.Handle)
	}
	updatedTopic, err := json.Marshal(topic)
	if err != nil {
		panic(err)
	}
	if err = ioutil.WriteFile(*handle, updatedTopic, 0600); err != nil {
		panic(err)
	}
}
