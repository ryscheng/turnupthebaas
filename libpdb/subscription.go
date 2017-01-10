type Subscription struct {
  // for PIR
	drbg *drbg.HashDrbg

  // Last seen sequence number
  Seqno      uint64
  Updates    chan []byte
}

//TODO: what sort of topic knowledge does a subscription need?
func NewSubscription(approximateSeqNo uint64) (*Subscription, error) {
  s := &Subscription{}
  s.Seqno = approximateSeqNo
  s.Updates := make(chan []bytes)

	hashDrbg, drbgErr := drbg.NewHashDrbg(nil)
  if drbgErr != nil {
		return nil, drbgErr
	}
	s.drbg = hashDrbg

  return s
}

func (s *Subscription) generatePoll(config *ClientConfig, seqNo uint64) (*common.ReadArgs, *common.ReadArgs, error) {
	args := make([]*common.ReadArgs, 2)
	seqNoBytes := make([]byte, 12)
	_ = binary.PutUvarint(seqNoBytes, seqNo)

	args[0] = &common.ReadArgs{}
	args[0].ForTd = make([]common.PirArgs, len(config.TrustDomains))
	for j := 0; j < len(config.TrustDomains); j++ {
		args[0].ForTd[j].RequestVector = make([]byte, config.CommonConfig.NumBuckets/8+1)
		s.drbg.FillBytes(args[0].ForTd[j].RequestVector)
		args[0].ForTd[j].PadSeed = make([]byte, drbg.SeedLength)
		s.drbg.FillBytes(args[0].ForTd[j].PadSeed)
	}
	// @todo - XOR topic info into request?
	//k0, k1 := t.Seed1.KeyUint128()
	//bucket1 := siphash.Hash(k0, k1, seqNoBytes) % globalConfig.NumBuckets

	args[1] = &common.ReadArgs{}
	args[1].ForTd = make([]common.PirArgs, len(config.TrustDomains))
	for j := 0; j < len(config.TrustDomains); j++ {
		args[1].ForTd[j].RequestVector = make([]byte, config.CommonConfig.NumBuckets/8+1)
		s.drbg.FillBytes(args[1].ForTd[j].RequestVector)
		args[1].ForTd[j].PadSeed = make([]byte, drbg.SeedLength)
		s.drbg.FillBytes(args[1].ForTd[j].PadSeed)
	}
	// @todo - XOR sopic info into request?
	//k0, k1 = t.Seed2.KeyUint128()
	//bucket2 := siphash.Hash(k0, k1, seqNoBytes) % globalConfig.NumBuckets

	return args[0], args[1], nil
}

func (s *Subscription) OnResponse(args *common.ReadArgs, reply *common.ReadReply) {
	msg := retrieveResponse(args, reply)
	if msg != nil && updates != nil {
		s.Updates <- msg
	}
}

// TODO: checksum msgs at topic level so if something random comes back it is filtered out.
func (s *Subscription) retrieveResponse(args *common.ReadArgs, reply *common.ReadReply) []byte {
	data := reply.Data

	for i := 0; i < len(args.ForTd); i++ {
		pad := make([]byte, len(data))
		seed, _ := drbg.ImportSeed(args.ForTd[i].PadSeed)
		hashDrbg, _ := drbg.NewHashDrbg(seed)
		hashDrbg.FillBytes(pad)
		for j := 0; j < len(data); j++ {
			data[j] ^= pad[j]
		}
	}
	return data
}
