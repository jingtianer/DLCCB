package main

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

var heightCmp int64
var j int64 = 0

func main() {
	sk := []byte("this is a shared secret key here")
	client, err := ethclient.Dial("/home/tt/eth/net/node2/geth.ipc")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("we have a connection")
	block, _ := client.BlockByNumber(context.Background(), nil)
	heightCmp = block.Number().Int64()
	// addr_Sk := create_addr_sk(j, sk, client)
	// add_Special := create_addr_special(addr_Sk)
	// fmt.Println("add_Special:", add_Special)

	fmt.Println("筛选开始")
	judge(j, sk, client)
	// cha(client)
}

func cha(client *ethclient.Client) {
	block, _ := client.BlockByNumber(context.Background(), big.NewInt(11670074))
	transactions := block.Transactions()
	for _, tx := range transactions {
		newBlockAddrSpecial := tx.Data()
		// fmt.Println("special:", []byte(add_Special))
		// fmt.Println("add:", tx.Data())
		fmt.Printf("newadd:%x", newBlockAddrSpecial)
		fmt.Println()
		// if add_Special == newBlockAddrSpecial {
		// 	fmt.Println("find trans :", add_Special)
		// 	break
		// }

	}
}

func create_addr_sk(j int, sk []byte, client *ethclient.Client) string {
	// 获得区块高度
	block, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	blockHeight := block.Number().Int64()

	fmt.Println(blockHeight)
	rand.Seed(time.Now().Unix())
	i := rand.Int63n(blockHeight-int64(j)) + int64(j)
	fmt.Println(i)
	block_i, err_i := client.BlockByNumber(context.Background(), big.NewInt(i))
	if err_i != nil {
		log.Fatal(err)
	}
	blockHash_i := block_i.Hash().Hex()
	// fmt.Println(reflect.TypeOf(blockHash_i))

	h := hmac.New(sha256.New, sk)
	h.Write([]byte(blockHash_i))
	bytes := h.Sum(nil)
	Addr_sk := hex.EncodeToString(bytes)
	// fmt.Println(Addr_sk)
	// fmt.Println(reflect.TypeOf(Addr_sk))
	return Addr_sk
}

func create_addr_special(Addr_sk string) string {
	hash := sha256.New()
	hash.Write([]byte(Addr_sk))
	bytes := hash.Sum(nil)
	Addr_special := hex.EncodeToString(bytes)
	// fmt.Println(Addr_special)
	return Addr_special
}

func judge(j int64, sk []byte, client *ethclient.Client) {
	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(j+100)))
	if err != nil {
		log.Fatal(err)
	}
	blockHeight := block.Number().Int64()
	var resaddr []string

	for k := int64(j); k < blockHeight; k++ {
		block_k, err_k := client.BlockByNumber(context.Background(), big.NewInt(k))
		if err_k != nil {
			log.Fatal(err_k)
		}
		blockHash_k := block_k.Hash().Hex()
		h := hmac.New(sha256.New, sk)
		h.Write([]byte(blockHash_k))
		bytes := h.Sum(nil)
		judgeAddrSk := hex.EncodeToString(bytes)
		judgeAddrSpecial := create_addr_special(judgeAddrSk)
		resaddr = append(resaddr, judgeAddrSpecial)
	}

	getNewBlockTransactions(resaddr, client, int64(j), sk)
}

func getNewBlockTransactions(resaddr []string, client *ethclient.Client, j int64, sk []byte) {
	// fmt.Println("getNewBlockTransactions")
	// fmt.Println("resaddrsize:", len(resaddr))
	res := []byte{}
	f := 1
	for f <= 34 {
		block, err := client.BlockByNumber(context.Background(), big.NewInt(j))

		if err != nil {
			log.Fatal(err)
		}

		transactions := block.Transactions()
		for _, tx := range transactions {
			newBlockAddrSpecial := tx.Data()
			for _, addr := range resaddr {
				tempaddr := []byte(addr)
				if bytes.Equal(tempaddr, newBlockAddrSpecial) {
					fmt.Println("num:", f)
					fmt.Println("height is :", j)
					fmt.Printf("find trans AddrSpecial: %x", newBlockAddrSpecial)
					fmt.Println()
					fmt.Println("find trans output address:", tx.To().Hex())
					fmt.Println()
					s := tx.To().Bytes()
					res = append(res, s[len(s)-1])
					f++
					break
				}
			}

		}
		j++

	}
	decryptedMessage, err := AESCTR(res, sk)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("decrypted message:", string(decryptedMessage))

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
