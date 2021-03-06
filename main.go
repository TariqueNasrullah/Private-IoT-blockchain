package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/TariqueNasrullah/iotchain/blockchain"
	"github.com/TariqueNasrullah/iotchain/cli"
	"golang.org/x/net/context"

	"google.golang.org/grpc"
)

var (
	keyPath = "key.data"
)

func generateToken(username, password string) ([]byte, error) {
	hash := []byte(username + password)

	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	r, s, err := ecdsa.Sign(rand.Reader, key, hash[:])
	if err != nil {
		return []byte{}, err
	}

	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}
func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func testBlockchain() {
	token, err := generateToken("admin", "pass")
	handle(err)
	key, err := blockchain.GenerateKey(keyPath)
	handle(err)
	key.Token = token

	trans := blockchain.Transaction{
		Data: []byte("Hello Transaction"),
	}
	block := blockchain.Block{
		Transactions: []*blockchain.Transaction{&trans},
		Token:        key.Token,
		PublicKey:    key.PublicKey,
	}

	err = block.Sign(key.PrivateKey)
	handle(err)

	pow := blockchain.NewProof(&block)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash

	fmt.Printf("%s\n", block)

	validSig := block.VerifySignature()
	fmt.Printf("Signature Verification: %v\n", validSig)
}

func sendAddrToGrpc() {
	conn, err := grpc.DialContext(context.Background(), "192.168.0.2:8000", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Duration(time.Second*10)))
	if err != nil {
		log.Fatalf("Conn established failed: %v\n", err)
	}
	defer conn.Close()

	clinet := blockchain.NewMinerClient(conn)

	response, err := clinet.SendAddress(context.Background(), &blockchain.SendAddressRequest{Addr: "192.168.0.2:8000"})
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		log.Printf("response: %v", response.ResponseText)
	}
}

func main() {
	cl := cli.CommandLine{}
	cl.Run()
}
