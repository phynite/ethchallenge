package challenge_test

import (
	"ethc/contracts/challenge"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"

	"github.com/ethereum/go-ethereum/crypto"
)

func TestDeployChallenge(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	genesisAlloc := map[common.Address]core.GenesisAccount{
		auth.From: {
			Balance: big.NewInt(1000000000),
		},
	}

	gasLimit := uint64(323167)

	fakeClient := backends.NewSimulatedBackend(genesisAlloc, gasLimit)

	address, _, _, err := challenge.DeployChallenge(
		auth,
		fakeClient,
	)

	fakeClient.Commit()
	if err != nil {
		t.Fatal(err)
	}

	if len(address.Bytes()) == 0 {
		t.Error("Not a valid deployment address")
	}
}

func TestSetAddress(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	genesisAlloc := map[common.Address]core.GenesisAccount{
		auth.From: {
			Balance: big.NewInt(1000000000),
		},
	}

	gasLimit := uint64(323167)

	fakeClient := backends.NewSimulatedBackend(genesisAlloc, gasLimit)

	fakeClient.Commit()

	_, _, contract, err := challenge.DeployChallenge(
		auth,
		fakeClient,
	)

	_, err = contract.SetA(&bind.TransactOpts{
		From:   auth.From,
		Signer: auth.Signer,
		Value:  nil,
	})

	fakeClient.Commit()

	if err != nil {
		t.Fatal(err)
	}
}

func TestGetAddress(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	genesisAlloc := map[common.Address]core.GenesisAccount{
		auth.From: {
			Balance: big.NewInt(1000000000),
		},
	}

	gasLimit := uint64(323167)

	fakeClient := backends.NewSimulatedBackend(genesisAlloc, gasLimit)

	_, _, contract, err := challenge.DeployChallenge(
		auth,
		fakeClient,
	)

	fakeClient.Commit()

	contract.SetA(&bind.TransactOpts{
		From:   auth.From,
		Signer: auth.Signer,
		Value:  nil,
	})

	fakeClient.Commit()

	got, err := contract.GetA(nil)

	if err != nil {
		t.Fatal(err)
	}

	if got != auth.From {
		t.Fatalf("The addresses do not match; got 0x%x, want 0x%x", got, auth.From)
	}
}
