package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/privacylab/talek/common"
)

var server = flag.Bool("server", false, "Create configuration for a talek server.")
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

	if !*server {
		fmt.Println("Talekutil doesn't handle client configuration yet.")
		return
	}

	var td *common.TrustDomainConfig
	var err error

	if len(*infile) > 0 {
		dat, readerr := ioutil.ReadFile(*infile)
		if readerr != nil {
			fmt.Printf("Could not read input file: %v\n", err)
			return
		}
		err = json.Unmarshal(dat, &td)
		if err != nil {
			fmt.Printf("Could not parse input file: %v\n", err)
			return
		}
	} else {
		td = common.NewTrustDomainConfig(*name, *address, true, false)
	}
	td.Name = *name
	td.Address = *address
	td.IsValid = true

	var out []byte
	if *private {
		tdp := td.Private()
		if bytes.Compare(tdp.PrivateKey[:], make([]byte, 32)) == 0 {
			fmt.Printf("Imported configuration did not include key.\n")
			return
		}
		out, err = json.Marshal(tdp)
		if err != nil {
			fmt.Printf("Failed to export config: %v\n", err)
			return
		}
	} else {
		out, err = json.Marshal(td)
		if err != nil {
			fmt.Printf("Failed to export config: %v\n", err)
			return
		}
	}
	err = ioutil.WriteFile(*outfile, out, 0644)
	if err != nil {
		fmt.Printf("Failed to write file: %v\n", err)
		return
	}
}
