package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/drbg"
	"github.com/privacylab/talek/libtalek"
)

// ReplicasEquivalent makes reads for single cells in single replicas, expecting them to be the same.
func ReplicasEquivalent(config libtalek.ClientConfig, leaderRPC common.FrontendRPC, spotChecks int) bool {
	for cell := uint64(0); cell < config.Config.NumBuckets; cell += config.Config.NumBuckets / uint64(spotChecks) {
		var expectedForCell []byte
		for replica := 0; replica < len(config.TrustDomains); replica++ {
			args := initReadArg(config.Config.NumBuckets, len(config.TrustDomains))
			args.TD[replica].RequestVector[cell/8] ^= (1 << (cell % 8))

			encArgs, _ := args.Encode(config.TrustDomains)
			reply := common.ReadReply{}
			leaderRPC.Read(&encArgs, &reply)
			if replica == 0 {
				expectedForCell = reply.Data
			} else {
				if !bytes.Equal(reply.Data, expectedForCell) {
					fmt.Fprintf(os.Stderr, "Disagreement of value of cell %d between 0 and %d.\n", cell, replica)
					return false
				}
			}
		}
		fmt.Fprintf(os.Stderr, ".")
	}
	fmt.Fprintf(os.Stderr, "\n")
	return true
}

// WritesPersisted writes into individual cells, and then attempts to read back those values
func WritesPersisted(config libtalek.ClientConfig, leaderRPC common.FrontendRPC, spotChecks int) bool {
	// 2. write to DB and make sure replicas propagate that write.
	for cell := uint64(0); cell < config.Config.NumBuckets; cell += config.Config.NumBuckets / uint64(spotChecks) {
		writeItem := make([]byte, config.Config.DataSize)
		for i := 0; i < len(writeItem); i++ {
			writeItem[i] = byte(i)
		}
		writeArgs := common.WriteArgs{
			Bucket1: cell,
			Bucket2: cell,
			Data:    writeItem,
		}
		writeReply := common.WriteReply{}
		leaderRPC.Write(&writeArgs, &writeReply)
		time.Sleep(config.WriteInterval)

		//responses will be overlayed with (in this case the 0) pad for TD privacy by each TD
		seed := make([]byte, drbg.SeedLength)
		for i := 0; i < len(config.TrustDomains); i++ {
			drbg.Overlay(seed, writeItem)
		}

		for replica := 0; replica < len(config.TrustDomains); replica++ {
			args := initReadArg(config.Config.NumBuckets, len(config.TrustDomains))
			args.TD[replica].RequestVector[cell/8] ^= (1 << (cell % 8))

			fmt.Fprintf(os.Stderr, "Pre-encoded Read Request is %v \n", args)
			encArgs, _ := args.Encode(config.TrustDomains)
			reply := common.ReadReply{}
			leaderRPC.Read(&encArgs, &reply)

			// Reply.data is bucket, and may have multiple items.
			found := false
			for i := uint64(0); i < config.Config.BucketDepth; i++ {
				if bytes.Equal(reply.Data[i*config.Config.DataSize:(i+1)*config.Config.DataSize], writeItem) {
					found = true
					break
				}
			}
			if !found {
				fmt.Fprintf(os.Stderr, "read after write failed. cell %d. replica %d.\n", cell, replica)
				return false
			}
		}
		fmt.Fprintf(os.Stderr, ".")
	}
	fmt.Fprintf(os.Stderr, "\n")
	return true
}
