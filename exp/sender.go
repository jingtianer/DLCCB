package main

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	sendTime, isM := send()
	fmt.Println(sendTime, " ", sendTime.Milliseconds(), " ", isM)
}

var j int64 = 0

func send() (time.Duration, bool) {
	rand.Seed(time.Now().UnixNano())
	messageLength := 34
	message := ""
	for i := 0; i < messageLength; i++ {
		message = message + string(rand.Intn(26)+int('a'))
	}
	sk := []byte("this is a shared secret key here")
	fmt.Println("message:", message)
	start := time.Now() // 获取当前时间
	encryptedMessage, err := AESCTR([]byte(message), sk)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("encrypted message:", hex.EncodeToString(encryptedMessage))

	// 根据加密消息生成接收方地址
	fmt.Println("receiver address:")
	var receiverAddressList []common.Address
	pieceLength := 1
	for i := 0; i < len(encryptedMessage); i += pieceLength {
		piece := encryptedMessage[i : i+pieceLength]
		for {
			// 新建一个私钥
			privateKey, err := crypto.GenerateKey()
			if err != nil {
				log.Fatal(err)
			}

			// 计算公钥和地址
			publicKey := privateKey.Public().(*ecdsa.PublicKey)
			address := crypto.PubkeyToAddress(*publicKey)
			if bytes.Equal(address[20-pieceLength:20], piece) {
				receiverAddressList = append(receiverAddressList, address)
				fmt.Printf("%d %#v\n", i+1, address)
				break
			}
		}
	}

	// 连接一个客户端
	client, err := ethclient.Dial("/home/tt/eth/net/node2/geth.ipc")
	if err != nil {
		log.Fatal("Dial ", err)
	}
	// 加载发送方私钥
	senderPrivateKey, err := crypto.HexToECDSA("eafea0e2167d8876a6a6dd056113db4f8acdbf649063ad6cd41f047acef22635")
	if err != nil {
		log.Fatal("HexToECDSA ", err)
	}

	// 计算发送方公钥和地址
	senderPublicKey := senderPrivateKey.Public().(*ecdsa.PublicKey)
	senderAddress := crypto.PubkeyToAddress(*senderPublicKey)

	addrSk, next_j, M := createAddrSk(j, sk, client)
	addrSpecial := createAddrSpecial(addrSk)
	fmt.Println("add_Special:", addrSpecial)

	// var totalGas *big.Int = big.NewInt(0)
	for index, receiverAddress := range receiverAddressList {
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
		data := []byte(addrSpecial)
		gasLimit, err := client.EstimateGas(
			context.Background(),
			ethereum.CallMsg{
				To:   &receiverAddress,
				Data: data,
			},
		)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(gasLimit)

		// 获取平均gas价格
		gasPrice, err := client.SuggestGasPrice(context.Background())
		// totalGas = totalGas.Add(totalGas, gasPrice)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(gasPrice)

		// 生成交易
		transaction := types.NewTransaction(senderNonce, receiverAddress, amount, gasLimit, gasPrice, data)
		// fmt.Printf("%#v\n", transaction)
		// fmt.Println(transaction.Hash())

		// 使用私钥对交易签名
		chainID, err := client.NetworkID(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		signedTransaction, err := types.SignTx(transaction, types.NewEIP155Signer(chainID), senderPrivateKey)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("%#v\n", signedTransaction)
		// fmt.Println(signedTransaction.Hash())

		// 发送交易
		err = client.SendTransaction(context.Background(), signedTransaction)
		if err != nil {
			log.Fatal(err)
		}

		// 确认交易，上链成功再发送下一条
		for {
			time.Sleep(1 * time.Second)
			_, isPending, err := client.TransactionByHash(context.Background(), signedTransaction.Hash())
			if err != nil {
				log.Fatal(err)
			}
			if !isPending {
				fmt.Println("Transaction", index+1, "sent")
				break
			}
		}
	}
	elapsed := time.Since(start)
	fmt.Println("消息"+message+"发送完成耗时：", elapsed)
	covertness := j%M == 0
	j = next_j
	return elapsed, covertness
}

func createAddrSk(j int64, sk []byte, client *ethclient.Client) (string, int64, int64) {
	// 获得区块高度
	block, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	blockHeight := block.Number().Int64()
	print("blockHeight = ", blockHeight)
	fmt.Println(blockHeight)
	blockJ, err := client.BlockByNumber(context.Background(), big.NewInt(j))
	if err != nil {
		log.Fatal("BlockByNumberJ ", err)
	}
	blockHashJ := blockJ.Hash()
	fmt.Println(blockHashJ)
	// var M int64 = j%int64(10) + 1
	var M int64 = blockHashJ.Big().Int64()%int64(10) + j%int64(10) + 1
	i := Min(blockHeight-1, j+M-j%M)
	blockI, err := client.BlockByNumber(context.Background(), big.NewInt(i))
	if err != nil {
		log.Fatal("BlockByNumberI ", err)
	}
	blockHashI := blockI.Hash().Hex()
	h := hmac.New(sha256.New, sk)
	h.Write([]byte(blockHashI))
	bytes := h.Sum(nil)
	AddrSk := hex.EncodeToString(bytes)
	return AddrSk, i, M
}

func Min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func createAddrSpecial(AddrSk string) string {
	hash := sha256.New()
	hash.Write([]byte(AddrSk))
	bytes := hash.Sum(nil)
	AddrSpecial := hex.EncodeToString(bytes)
	return AddrSpecial
}

// AESCTR 使用AES中的计算器模式进行加密和解密
func AESCTR(data []byte, key []byte) ([]byte, error) {
	//1. 创建cipher.Block接口
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//2. 创建分组模式，在crypto/cipher包中
	iv := bytes.Repeat([]byte("1"), block.BlockSize())
	stream := cipher.NewCTR(block, iv)
	//3. 加密
	dst := make([]byte, len(data))
	stream.XORKeyStream(dst, data)

	return dst, nil
}
