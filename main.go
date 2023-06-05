package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/types"
	"github.com/thanhpn/jupiter/pkg/model"
)

var jupiterBaseUrl = "https://quote-api.jup.ag/v5"

func main() {
	// get quote
	res, err := getQuote()
	fmt.Println(res, err)

	// build swap data
	jupiterBuildSwapRouteRequest := &model.JupiterBuildSwapRouteRequest{
		QuoteResponse:                 res,
		UserPublicKey:                 "E9naYkA74q8xPmNYb8To9dba6Fz6xCFZTE4SFz6Quv43",
		WrapAndUnwrapSol:              true,
		ComputeUnitPriceMicroLamports: "auto",
	}
	swapData, err := getSwap(jupiterBuildSwapRouteRequest)
	fmt.Println("-----------------------------")
	fmt.Println(swapData, err)
	fmt.Println("-----------------------------")

	// submit swap transaction
	rpcClient := client.NewClient("https://api.mainnet-beta.solana.com")
	privateKey := ""
	centralAcc, _ := types.AccountFromBase58(privateKey)

	// decode transaction
	rawTx, err := base64.StdEncoding.DecodeString(swapData.SwapTransaction)
	if err != nil {
		fmt.Errorf("failed to base64 decode data, err: %v", err)
	}
	tx, err := types.TransactionDeserialize(rawTx)
	if err != nil {
		fmt.Errorf("failed to deserialize transaction, err: %v", err)
	}

	blockhash, err := rpcClient.GetLatestBlockhash(context.Background())
	if err != nil {
		fmt.Errorf("client.GetLatestBlockhash() failed", err)
	}
	tx.Message.RecentBlockHash = blockhash.Blockhash

	// sign transaction
	data, err := tx.Message.Serialize()
	if err != nil {
		fmt.Errorf("failed to serialize message, err: %v", err)
	}
	tx.Signatures[0] = centralAcc.Sign(data)

	txHash, err := rpcClient.SendTransaction(context.Background(), tx)
	fmt.Println("txHash: ", txHash)
	fmt.Println("err: ", err)
	fmt.Println("-----------------------------")
	if err != nil {
		fmt.Errorf("[Swap] client.SendTransaction() failed %v", err)
	}
	fmt.Println("-----------------------------")
}

func getQuote() (*model.JupiterSwapRoutesSol, error) {
	var client = &http.Client{}
	fromAddress := "So11111111111111111111111111111111111111112" // SOL
	toAddress := "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"  // USDC
	amount := "10000"

	request, err := http.NewRequest("GET", fmt.Sprintf("%s/quote?inputMint=%s&outputMint=%s&amount=%s&slippageBps=50&onlyDirectRoutes=false", jupiterBaseUrl, fromAddress, toAddress, amount), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(request)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := &model.JupiterSwapRoutesSol{}
	err = json.Unmarshal(resBody, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func getSwap(req *model.JupiterBuildSwapRouteRequest) (*model.JupiterBuildRoute, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	jsonBody := bytes.NewBuffer(body)

	var client = &http.Client{}
	url := fmt.Sprintf("%s/swap", jupiterBaseUrl)
	fmt.Println("jsonBody: ", jsonBody)
	httpRequest, err := http.NewRequest("POST", url, jsonBody)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := &model.JupiterBuildRoute{}
	err = json.Unmarshal(resBody, res)

	if err != nil {
		return nil, err
	}

	return res, nil
}
