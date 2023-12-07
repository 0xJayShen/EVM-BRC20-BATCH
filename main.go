package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"
)

func getNonce(node, priv string) (uint64, error) {
	privateKey, err := crypto.HexToECDSA(priv)
	if err != nil {
		log.Fatalf("Failed to convert private key: %v", err)
		return 0, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Error casting public key to ECDSA")
		return 0, err
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	client, err := ethclient.Dial(node)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		return 0, err
	}
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
		return 0, err
	}
	fmt.Println("nonce now is ------- ", nonce)
	return nonce, nil
}
func getChainID(node string) (*big.Int, error) {
	client, err := ethclient.Dial(node)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		return nil, err
	}
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
		return nil, err
	}
	fmt.Println("chain id is ------- ", chainID)
	return chainID, nil
}
func do(node string, priv string, msg string, nonce uint64, chainID *big.Int, gasPrice_ int64) error {
	client, err := ethclient.Dial(node)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		return err
	}

	privateKey, err := crypto.HexToECDSA(priv)
	if err != nil {
		log.Fatalf("Failed to convert private key: %v", err)
		return err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	value := big.NewInt(0)
	gasLimit := uint64(22000)
	gasPrice := big.NewInt(gasPrice_)
	if err != nil {
		log.Fatalf("Failed to suggest gas price: %v", err)
		return err
	}
	data := []byte(msg)
	tx := types.NewTransaction(nonce, fromAddress, value, gasLimit, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
		return err
	}
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
		return err
	}

	fmt.Printf("Transaction sent! TX Hash: %s\n", signedTx.Hash().Hex())
	return nil
}
func main() {
	var (
		Node      = os.Args[1]
		Priv      = os.Args[2]
		Msg       = os.Args[3]
		Gas       = os.Args[4]
		LoopTotal = os.Args[5]
		SleepTime = os.Args[6]
	)

	nonceNow, err := getNonce(Node, Priv)
	if err != nil {
		panic(err)
	}

	chainID, err := getChainID(Node)
	if err != nil {
		panic(err)
	}

	gas, err := strconv.ParseInt(Gas, 10, 64)
	if err != nil {
		panic(err)
	}

	loopTotal_, err := strconv.ParseInt(LoopTotal, 10, 64)
	if err != nil {
		panic(err)
	}

	sleep_, err := strconv.ParseInt(SleepTime, 10, 64)
	if err != nil {
		panic(err)
	}
	for i := 0; i < int(loopTotal_); i++ {
		time.Sleep(time.Duration(sleep_) * time.Millisecond)
		if err := do(Node, Priv, Msg, nonceNow, chainID, gas); err != nil {
			panic(err)
		}
	}
}
