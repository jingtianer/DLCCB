package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func createAccount() {
	ks := keystore.NewKeyStore("./account/keystore2", keystore.StandardScryptN, keystore.StandardScryptP)
	// _ := accounts.NewManager(&accounts.Config{InsecureUnlockAllowed: false}, ks)
	// Create a new account with the specified encryption passphrase.
	newAcc, _ := ks.NewAccount("Creation password")
	fmt.Println(newAcc)

	// Export the newly created account with a different passphrase. The returned
	// data from this method invocation is a JSON encoded, encrypted key-file.
	jsonAcc, _ := ks.Export(newAcc, "Creation password", "Export password")

	// Update the passphrase on the account created above inside the local keystore.
	_ = ks.Update(newAcc, "Creation password", "Update password")
	fmt.Println(jsonAcc)
}

func main() {
	txPerSecond := 10
	client, err := ethclient.Dial("/home/tt/eth/net/node1/geth.ipc")
	if err != nil {
		log.Fatal(err)
	}
	// 加载发送方私钥
	senderPrivateKey, err := crypto.HexToECDSA("6026f0ab8b2cb9d3d70758d54cfe314eff05b2ad66f7b2c9241d6dfbe6d7b993")
	fmt.Println("senderPrivateKey = ", senderPrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	// 计算发送方公钥和地址
	senderPublicKey := senderPrivateKey.Public().(*ecdsa.PublicKey)
	senderAddress := crypto.PubkeyToAddress(*senderPublicKey)
	// senderAddress := common.HexToAddress("0x037d9076363c82c5d9a6b1f3df921e22220000e8")
	fmt.Println("senderAddress", senderAddress)
	balance, err := client.BalanceAt(context.Background(), senderAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("balance = ", balance)

	receiverAddressChan := make(chan common.Address, 1)
	go func() {
		for {
			// 新建一个私钥
			privateKey, err := crypto.GenerateKey()
			if err != nil {
				log.Fatal(err)
			}

			// 计算公钥和地址
			publicKey := privateKey.Public().(*ecdsa.PublicKey)
			address := crypto.PubkeyToAddress(*publicKey)
			receiverAddressChan <- address
			time.Sleep(1000 / time.Duration(txPerSecond) * time.Millisecond)
		}
	}()

	var totalGas *big.Int = big.NewInt(0)
	index := 0
	for {
		receiverAddress := <-receiverAddressChan
		// 获得账户的senderNonce
		senderNonce, err := client.PendingNonceAt(context.Background(), senderAddress)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(senderNonce)

		// 需要转移的ETH数量
		amount := big.NewInt(1)
		// fmt.Println(amount)

		// 标准ETH交易的gas限制

		data := senderAddress.Bytes()
		gasLimit, err := client.EstimateGas(
			context.Background(),
			ethereum.CallMsg{
				From: senderAddress,
				To:   &receiverAddress,
				Data: data,
			},
		)

		if err != nil {
			log.Fatal("error1 ", err)

		}
		fmt.Println("gasLimit = ", gasLimit)

		// 获取平均gas价格
		gasPrice, err := client.SuggestGasPrice(context.Background())
		totalGas = totalGas.Add(totalGas, gasPrice)
		if err != nil {
			log.Fatal("error2 ", err)
		}
		fmt.Println("gasPrice = ", gasPrice)

		// 生成交易
		transaction := types.NewTransaction(senderNonce, receiverAddress, amount, gasLimit, gasPrice, data)
		// fmt.Printf("%#v\n", transaction)
		// fmt.Println(transaction.Hash())

		// 使用私钥对交易签名
		chainID, err := client.NetworkID(context.Background())
		if err != nil {
			log.Fatal("NetworkID error ", err)
		}
		fmt.Println("chainID", chainID)
		signedTransaction, err := types.SignTx(transaction, types.NewEIP155Signer(chainID), senderPrivateKey)
		if err != nil {
			log.Fatal("SignTx error ", err)
		}
		// fmt.Printf("%#v\n", signedTransaction)
		// fmt.Println(signedTransaction.Hash())

		// 发送交易
		err = client.SendTransaction(context.Background(), signedTransaction)
		if err != nil {
			log.Fatal("SendTransaction error ", err)
		}

		// 确认交易，上链成功再发送下一条
		for {
			time.Sleep(1 * time.Second)
			_, isPending, err := client.TransactionByHash(context.Background(), signedTransaction.Hash())
			if err != nil {
				log.Fatal(err)
			}
			if !isPending {
				fmt.Println("Transaction ", index+1, " sent")
				break
			}
		}
	}
}
