package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
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
	follow := pflag.Bool("follow", false, "Keep reading until interrupt")
	randSeed := pflag.Int("randSeed", 0, "Use a deterministic random seed. [Dangerous!]")
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
		err = topic.UnmarshalText(topicdata)
		if err != nil {
			// if it is a read-only share, unmarshal only the Handle part
			if *read {
				topic.SigningPrivateKey = new([64]byte)
				err = topic.Handle.UnmarshalText(topicdata)
				if err != nil {
					panic(err)
				}
			} else {
				panic(err)
			}
		}
	}

	if len(*share) > 0 {
		handle := topic.Handle
		handleBytes, handleerr := handle.MarshalText()
		if handleerr != nil {
			panic(handleerr)
		}
		ioutil.WriteFile(*share, handleBytes, 0640)
		fmt.Fprintf(os.Stderr, "Read-only topic written to %s\n", *share)
		return
	}

	// Client-connected activity below.
	frontendRPC := common.NewFrontendRPC("RPC", config.FrontendAddr)

	client := libtalek.NewClient("Client", *config, frontendRPC)
	if client == nil {
		panic("could not create talek client")
	} else if *verbose {
		fmt.Fprintln(os.Stderr, "Connection to RPC established.")
	}
	if *randSeed != 0 {
		fmt.Fprintln(os.Stderr, "Using deterministic random seed. This can cause replay of nonces, and should not be used in production.")
		r := rand.New(rand.NewSource(int64(*randSeed)))
		client.Rand = r
	}
	client.Verbose = *verbose

	if *read == false && len(*write) > 0 {
		if err = client.Publish(&topic, []byte(*write)); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to publish: %s\n", err)
			panic(err)
		}
		client.Flush()
	} else if *read == false {
		if *create {
			if *verbose {
				fmt.Fprintf(os.Stderr, "New topic written to %s.\n", *handlePath)
			}
		} else {
			fmt.Fprintf(os.Stderr, "No Read or Write operation requested. Closing.\n")
		}
	} else {
		if len(*write) > 0 {
			fmt.Fprintf(os.Stderr, "Cannot read and write at the same time.\n")
			return
		}
		msgs := client.Poll(&topic.Handle)

		if *follow {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
		forever:
			for {
				select {
				case next := <-msgs:
					fmt.Fprintf(os.Stdout, "%s\n", next)
				case <-c:
					break forever
				}
			}
		} else {
			timeout := time.After(config.ReadInterval * readTimeoutMultiple)
			select {
			case next := <-msgs:
				fmt.Fprintf(os.Stdout, "%s\n", next)
				break
			case <-timeout:
				if *verbose {
					fmt.Fprintln(os.Stderr, "Timed out waiting for new message.")
				}
				return
			}
		}
		client.Done(&topic.Handle)
	}

	updatedTopic, err := topic.MarshalText()
	if err != nil {
		panic(err)
	}
	if err = ioutil.WriteFile(*handlePath, updatedTopic, 0600); err != nil {
		panic(err)
	}
}
