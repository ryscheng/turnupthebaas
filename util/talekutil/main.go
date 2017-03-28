package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/libtalek"
	"github.com/privacylab/talek/server"
)

var outputClient = flag.Bool("client", false, "Create configuration for a talek client.")
var outputServer = flag.Bool("server", false, "Create configuration for a talek server.")
var outputTD = flag.Bool("trustdomain", false, "Create raw trustdomain configuration.")
var name = flag.String("name", "talek", "Server Name.")
var address = flag.String("address", "localhost:9000", "Server Address.")
var infile = flag.String("infile", "", "Begin with configuration from file.")
var outfile = flag.String("outfile", "talek.json", "Save configuration to file.")
var private = flag.Bool("private", false, "Include private key configuration.")
var trustdomains = flag.String("trustdomains", "talek.json", "Comma separated list of trust domains.")

// Talekutil is used to generate configuration files for structuring a talek
// deployment.  In particular, creating a set of configuration files for the
// clients and various trust domains.
func main() {
	flag.Parse()

	if !*outputServer && !*outputTD && !*outputClient {
		fmt.Println("Talekutil needs a mode: -client, -server, or -trustdomain.")
		return
	} else if (*outputServer && *outputTD) || (*outputClient && *outputServer) || (*outputClient && *outputTD) {
		fmt.Println("Mode must be one of -server or -trustdomain or -client.")
		return
	}

	if *outputClient {
		clientUtil(*infile, *outfile, *trustdomains)
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
			fmt.Printf("Could not read input file: %v\n", readerr)
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
		// We write a custom version of the server config that is still able to be
		// unmarshaled. In particular, the code below uses the serialized version of
		// the trust domain from above, which may have it's private key stripped
		// (which can't easily be directly specified), and with the pointer to the
		// common config removed.

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

// update/create client configuration with an explicit set of server trust domains.
func clientUtil(infile string, outfile string, trustfiles string) {
	domainPaths := strings.Split(trustfiles, ",")
	trustDomains := make([]*common.TrustDomainConfig, len(domainPaths))
	for i, path := range domainPaths {
		tdString, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("Could not read %s!\n", path)
			return
		}
		trustDomains[i] = new(common.TrustDomainConfig)
		if err := json.Unmarshal(tdString, trustDomains[i]); err != nil {
			log.Printf("Could not parse %s: %v\n", path, err)
			return
		}
	}

	clientconf := libtalek.ClientConfig{
		ReadInterval:  time.Second,
		WriteInterval: time.Second,
	}
	if len(infile) > 0 {
		dat, readerr := ioutil.ReadFile(infile)
		if readerr != nil {
			fmt.Printf("Could not read input file: %v\n", readerr)
			return
		}
		if err := json.Unmarshal(dat, &clientconf); err != nil {
			fmt.Printf("Could not parse input file: %v\n", err)
			return
		}
	}

	clientconf.TrustDomains = trustDomains
	bytes, err := json.MarshalIndent(clientconf, "", "  ")
	if err != nil {
		fmt.Printf("Failed to export config: %v\n", err)
		return
	}
	err = ioutil.WriteFile(outfile, bytes, 0644)
	if err != nil {
		fmt.Printf("Failed to write %s: %v\n", outfile, err)
		return
	}
}
