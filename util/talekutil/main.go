package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/server"
)

var outputServer = flag.Bool("server", false, "Create configuration for a talek server.")
var outputTD = flag.Bool("trustdomain", false, "Create raw trustdomain configuration.")
var name = flag.String("name", "talek", "Server Name.")
var address = flag.String("address", "localhost:9000", "Server Address.")
var infile = flag.String("infile", "", "Begin with configuration from file.")
var outfile = flag.String("outfile", "talek.json", "Save configuration to file.")
var private = flag.Bool("private", false, "Include private key configuration.")

// Talekutil is used to generate configuration files for structuring a talek
// deployment.  In particular, creating a set of configuration files for the
// clients and various trust domains.
func main() {
	flag.Parse()

	if !*outputServer && !*outputTD {
		fmt.Println("Talekutil doesn't handle client configuration yet.")
		return
	} else if *outputServer && *outputTD {
		fmt.Println("Mode must be one of -server or -trustdomain.")
		return
	}

	var sc = server.Config{
		ReadBatch:     8,
		WriteInterval: time.Second,
		ReadInterval:  time.Second,
	}
	var tdc common.TrustDomainConfig
	var err error

	if len(*infile) > 0 {
		dat, readerr := ioutil.ReadFile(*infile)
		if readerr != nil {
			fmt.Printf("Could not read input file: %v\n", err)
			return
		}
		err = json.Unmarshal(dat, &tdc)
		if err != nil {
			fmt.Printf("Could not parse input file: %v\n", err)
			return
		}
		err = json.Unmarshal(dat, &sc)
		if err != nil {
			fmt.Printf("Could not parse input file: %v\n", err)
			return
		}
	} else {
		tdc = *common.NewTrustDomainConfig(*name, *address, true, false)
	}
	if sc.TrustDomain != nil {
		tdc = *sc.TrustDomain
	}
	tdc.Name = *name
	tdc.Address = *address
	tdc.IsValid = true

	var tdb []byte
	if *private {
		tdp := tdc.Private()
		if bytes.Compare(tdp.PrivateKey[:], make([]byte, 32)) == 0 {
			fmt.Printf("Imported configuration did not include key.\n")
			return
		}
		tdb, err = json.MarshalIndent(tdp, "", "  ")
		if err != nil {
			fmt.Printf("Failed to export config: %v\n", err)
			return
		}
	} else {
		tdb, err = json.MarshalIndent(tdc, "", "  ")
		if err != nil {
			fmt.Printf("Failed to export config: %v\n", err)
			return
		}
	}
	if *outputTD {
		err = ioutil.WriteFile(*outfile, tdb, 0640)
		if err != nil {
			fmt.Printf("Failed to write file: %v\n", err)
			return
		}
	} else if *outputServer {
		// first encode
		servraw, err := json.Marshal(sc)
		if err != nil {
			fmt.Printf("Cannot flatten server: %v\n", err)
			return
		}
		// reload both server config and trustdomain config as JSON messages
		var servstruct map[string]interface{}
		err = json.Unmarshal(servraw, &servstruct)
		if err != nil {
			fmt.Printf("Failed to unmarshal server: %v\n", err)
			return
		}
		delete(servstruct, "CommonConfig")
		delete(servstruct, "TrustDomain")
		var tdstruct map[string]interface{}
		err = json.Unmarshal(tdb, &tdstruct)
		if err != nil {
			fmt.Printf("Failed to unmarshal trust domain: %v\n", err)
			return
		}
		servstruct["TrustDomain"] = tdstruct

		servraw, err = json.MarshalIndent(servstruct, "", "  ")
		if err != nil {
			fmt.Printf("Could not flatten combined server config: %v\n", err)
			return
		}
		err = ioutil.WriteFile(*outfile, servraw, 0640)
		if err != nil {
			fmt.Printf("Failed to write file: %v\n", err)
			return
		}
	}
}
