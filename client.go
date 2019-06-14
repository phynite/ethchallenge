package main

import (
	"context"
	"encoding/json"
	"ethc/contracts/challenge"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/RTradeLtd/rtfs"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"

	"github.com/ethereum/go-ethereum/crypto"
)

const (
	ipfsHost = "/ip4/127.0.0.1/tcp/5001"
)

func main() {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	genesisAlloc := map[common.Address]core.GenesisAccount{
		auth.From: {
			Balance: big.NewInt(1000000000),
		},
	}

	gasLimit := uint64(250000)

	fakeClient := backends.NewSimulatedBackend(genesisAlloc, gasLimit)

	contractAddress, _, instance, err := challenge.DeployChallenge(
		auth,
		fakeClient,
	)

	fakeClient.Commit()

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Contract deployed @ 0x%x\n", contractAddress)

	mgr, err := rtfs.NewManager(ipfsHost, "", time.Minute*3)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Connected to IPFS @ %s\n", ipfsHost)

	ch := make(chan string, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for val := range ch {
			log.Printf("New IPLD Object @ %s\n", val)
		}
	}()

	/******************************
	 * Calling contract functions *
	 ******************************/
	txSetA, err := instance.SetA(&bind.TransactOpts{
		From:   auth.From,
		Signer: auth.Signer,
		Value:  nil,
	})

	fakeClient.Commit()

	receiptA, err := fakeClient.TransactionReceipt(context.Background(), txSetA.Hash())
	if err != nil {
		log.Fatal(err)
	}
	if receiptA == nil {
		log.Fatal("receiptA is nil. Forgot to commit?")
	}

	if err = ipfsStore(mgr, ch, receiptA); err != nil {
		log.Fatal(err)
	}

	txGetA, err := instance.GetA(nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("address value: 0x%x\n", txGetA)

	txSetB, err := instance.SetB(&bind.TransactOpts{
		From:   auth.From,
		Signer: auth.Signer,
		Value:  nil,
	}, big.NewInt(1337))

	if err != nil {
		log.Fatal(err)
	}

	fakeClient.Commit()

	receiptB, err := fakeClient.TransactionReceipt(context.Background(), txSetB.Hash())
	if err != nil {
		log.Fatal(err)
	}
	if receiptB == nil {
		log.Fatal("receiptB is nil. Forgot to commit?")
	}

	if err = ipfsStore(mgr, ch, receiptB); err != nil {
		log.Fatal(err)
	}

	txGetB, err := instance.GetB(nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("b value: %v\n", txGetB)

	close(ch)
	wg.Wait()
}

func ipfsStore(mgr *rtfs.IpfsManager, ch chan string, receipt *types.Receipt) error {
	type storeObj struct {
		Receipt *types.Receipt `json:"receipt"`
	}

	if mgr == nil {
		return fmt.Errorf("ipfsStore: IPFS manager is nil")
	}

	obj := storeObj{Receipt: receipt}
	marshaled, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("ipfsStore: %v", err)
	}

	s, err := mgr.DagPut(marshaled, "json", "cbor")
	if err != nil {
		return fmt.Errorf("ipfsStore: %v", err)
	}

	ch <- s
	return nil
}
