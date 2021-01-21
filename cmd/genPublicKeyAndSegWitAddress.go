package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/jayt106/bitcoinAddressGenerator/cipher"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {

	l := len(os.Args)

	var ip string
	var port string
	var serverPublicKey string
	var relativePath string
	if l >= 2 && strings.ToLower(os.Args[1]) == "help" {
		help()
		return
	} else if l == 3 {
		serverPublicKey = os.Args[1]
		relativePath = os.Args[2]
		ip = "localhost"
		port = "8080"
	} else if l == 5 {
		ip = os.Args[1]
		port = os.Args[2]
		serverPublicKey = os.Args[3]
		relativePath = os.Args[4]
	} else {
		fmt.Println("Invalid arguments, please check your input")
		help()
		return
	}

	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
		return
	}

	var filePath = workingDir + "/" + relativePath
	keyParam, err := ReadSeedFromJsonFile(&filePath)
	if err != nil {
		log.Fatalln(err)
		return
	}

	marshalledData, err := json.Marshal(keyParam)
	if err != nil {
		log.Fatalln(err)
		return
	}

	channelPrivKeyClient, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		log.Fatalln(err)
		return
	}

	var slice []byte
	slice = append(channelPrivKeyClient.PubKey().SerializeCompressed(), marshalledData...)

	bs, err := hex.DecodeString(serverPublicKey)
	if err != nil {
		log.Fatalln(err)
		return
	}

	// Verifying the receiving data is a ecdsa publicKey
	pubKey, err := btcec.ParsePubKey(bs, btcec.S256())
	if err != nil {
		log.Fatalln(err)
		return
	}

	ciphertext, err := cipher.MessageEncrypt(pubKey, &slice)
	if err != nil {
		log.Fatalln(err)
		return
	}

	data := make(map[string]string)
	data["data"] = hex.EncodeToString(*ciphertext)
	bytesData, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
		return
	}

	api := "http://" + ip + ":" + port + "/v1/genPublicKeyAndSegWitAddress"

	req, err := http.NewRequest("POST", api, bytes.NewReader(bytesData))
	if err != nil {
		log.Fatalln(err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		return
	}

	plaintext, err := cipher.MessageDecrypt(channelPrivKeyClient, &body)
	if err != nil {
		log.Fatalln(err)
		return
	}

	var rsp map[string]string
	err = json.Unmarshal(*plaintext, &rsp)
	if err != nil {
		log.Fatalln(err)
		return
	}
	publicKey := rsp["publicKey"]
	segwitAddress := rsp["segwitAddress"]

	fmt.Println("publicKey:", publicKey)
	fmt.Println("segwitAddress:", segwitAddress)
}

func help() {
	fmt.Println("usage: ./genPublicKeyAndSegWitAddress [ip] [port] [server public key] [seed file path]")
	fmt.Println()
	fmt.Println("For connecting with the default server: localhost:8080")
	fmt.Println("usage: ./genPublicKeyAndSegWitAddress [server public key] [seed file path]")
}