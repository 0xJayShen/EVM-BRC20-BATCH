package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"strconv"
	"strings"
	"time"
)

func getNonce(node, priv string, output *widget.Entry) (uint64, error) {
	privateKey, err := crypto.HexToECDSA(priv)
	if err != nil {
		appendOutput(output, fmt.Sprintf("Failed to convert private key: %v\n", err))
		return 0, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		appendOutput(output, "Error casting public key to ECDSA\n")
		return 0, err
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	client, err := ethclient.Dial(node)
	if err != nil {
		appendOutput(output, fmt.Sprintf("Failed to connect to the Ethereum client: %v\n", err))
		return 0, err
	}
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		appendOutput(output, fmt.Sprintf("Failed to get nonce: %v\n", err))
		return 0, err
	}
	appendOutput(output, fmt.Sprintf("Nonce obtained: %d\n", nonce))
	return nonce, nil
}

func getChainID(node string, output *widget.Entry) (*big.Int, error) {
	client, err := ethclient.Dial(node)
	if err != nil {
		appendOutput(output, fmt.Sprintf("Failed to connect to the Ethereum client: %v\n", err))
		return nil, err
	}
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		appendOutput(output, fmt.Sprintf("Failed to get chain ID: %v\n", err))
		return nil, err
	}
	appendOutput(output, fmt.Sprintf("Chain ID obtained: %s\n", chainID.String()))
	return chainID, nil
}

func do(node, priv, msg string, nonce uint64, chainID *big.Int, gasPrice_ int64, output *widget.Entry) error {
	client, err := ethclient.Dial(node)
	if err != nil {
		appendOutput(output, fmt.Sprintf("Failed to connect to the Ethereum client: %v\n", err))
		return err
	}

	privateKey, err := crypto.HexToECDSA(priv)
	if err != nil {
		appendOutput(output, fmt.Sprintf("Failed to convert private key: %v\n", err))
		return err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		appendOutput(output, "Error casting public key to ECDSA\n")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	value := big.NewInt(0)
	gasLimit := uint64(22000)
	gasPrice := big.NewInt(gasPrice_)

	data := []byte(msg)
	tx := types.NewTransaction(nonce, fromAddress, value, gasLimit, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		appendOutput(output, fmt.Sprintf("Failed to sign transaction: %v\n", err))
		return err
	}
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		appendOutput(output, fmt.Sprintf("Failed to send transaction: %v\n", err))
		return err
	}

	appendOutput(output, fmt.Sprintf("Transaction sent! TX Hash: %s\n", signedTx.Hash().Hex()))
	return nil
}

func startSending(node, priv, msg string, gas, loopTotal, sleepTime int64, output *widget.Entry) error {
	nonceNow, err := getNonce(node, priv, output)
	if err != nil {
		return err
	}

	chainID, err := getChainID(node, output)
	if err != nil {
		return err
	}

	for i := 0; i < int(loopTotal); i++ {
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)
		if err := do(node, priv, msg, nonceNow, chainID, gas, output); err != nil {
			return err
		}
		nonceNow++
	}

	appendOutput(output, "All transactions sent successfully!\n")
	return nil
}

func appendOutput(output *widget.Entry, text string) {
	lines := strings.Split(output.Text, "\n")
	if len(lines) > 500 {
		lines = lines[len(lines)-500:] // Keep only the last 500 lines
	}
	lines = append(lines, text)
	output.SetText(strings.Join(lines, "\n"))
	output.Refresh() // Refresh the widget to update the text
}

func main() {
	a := app.New()
	w := a.NewWindow("Ethereum Transaction Sender")
	nodeEntry := widget.NewEntry()
	nodeEntry.SetPlaceHolder("https://rpc.ankr.com/bsc/xxxxxxxx")

	privEntry := widget.NewPasswordEntry()
	privEntry.SetPlaceHolder("your priv")

	msgEntry := widget.NewEntry()
	msgEntry.SetPlaceHolder(`data:,{"p":"bsc-20","op":"mint","tick":"bsci","amt":"1000"}`)

	gasEntry := widget.NewEntry()
	gasEntry.SetPlaceHolder("6000000000")

	loopTotalEntry := widget.NewEntry()
	loopTotalEntry.SetPlaceHolder("10")

	sleepTimeEntry := widget.NewEntry()
	sleepTimeEntry.SetPlaceHolder("100")
	outputEntry := widget.NewMultiLineEntry()
	outputEntry.Wrapping = fyne.TextWrapWord

	scrollContainer := container.NewVScroll(outputEntry)
	scrollContainer.SetMinSize(fyne.NewSize(800, 300))

	startButton := widget.NewButton("Start", func() {
		node := nodeEntry.Text
		priv := privEntry.Text
		msg := msgEntry.Text
		gas, err := strconv.ParseInt(gasEntry.Text, 10, 64)
		if err != nil {
			appendOutput(outputEntry, fmt.Sprintf("Invalid gas price: %v\n", err))
			return
		}
		loopTotal, err := strconv.ParseInt(loopTotalEntry.Text, 10, 64)
		if err != nil {
			appendOutput(outputEntry, fmt.Sprintf("Invalid loop total: %v\n", err))
			return
		}
		sleepTime, err := strconv.ParseInt(sleepTimeEntry.Text, 10, 64)
		if err != nil {
			appendOutput(outputEntry, fmt.Sprintf("Invalid sleep time: %v\n", err))
			return
		}

		if err := startSending(node, priv, msg, gas, loopTotal, sleepTime, outputEntry); err != nil {
			appendOutput(outputEntry, fmt.Sprintf("Error: %s", err.Error()))
		}
	})

	w.SetContent(container.NewVBox(
		widget.NewLabel("Node URL"),
		nodeEntry,
		widget.NewLabel("Private Key"),
		privEntry,
		widget.NewLabel("Message"),
		msgEntry,
		widget.NewLabel("Gas Price"),
		gasEntry,
		widget.NewLabel("Loop Total"),
		loopTotalEntry,
		widget.NewLabel("Sleep Time (ms)"),
		sleepTimeEntry,
		startButton,
		scrollContainer,
	))

	w.ShowAndRun()
}
