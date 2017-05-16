package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/coreos/etcd/pkg/flags"
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/libtalek"
	"github.com/spf13/pflag"
)

const readTimeoutMultiple = 5

// The CLI client will read or write a single item for talek
func main() {
	configPath := pflag.String("config", "talek.conf", "Client configuration for talek")
	create := pflag.Bool("create", false, "Create a new talek handle")
	share := pflag.String("share", "", "Create a read-only version of the topic for sharing")
	handlePath := pflag.String("topic", "talek.handle", "The talek handle to use")
	write := pflag.String("write", "", "A message to append to the log (If not specified, the next item will be read.)")
	read := pflag.Bool("read", false, "Read from the provided topic")
	verbose := pflag.Bool("verbose", false, "Print diagnostic information")
	err := flags.SetPflagsFromEnv(common.EnvPrefix, pflag.CommandLine)
	if err != nil {
		fmt.Printf("Error reading environment variables, %v\n", err)
		return
	}
	pflag.Parse()

	// Config
	config := libtalek.ClientConfigFromFile(*configPath)
	if config == nil {
		fmt.Fprintln(os.Stderr, "Talek Client must be run with --config specifying where the server is.")
		os.Exit(1)
	}
	if config.Config == nil && *verbose {
		fmt.Fprintln(os.Stderr, "Common configuration will be fetched from frontend.")
	}

	topicdata, err := ioutil.ReadFile(*handlePath)
	if err != nil && !*create {
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
		if len(*write) > 0 {
			fmt.Fprintf(os.Stderr, "Cannot read and write at the same time. Ignoring write.")
		}
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
	if err = ioutil.WriteFile(*handlePath, updatedTopic, 0600); err != nil {
		panic(err)
	}
}
