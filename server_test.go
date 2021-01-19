package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/jayt106/bitcoinAddressGenerator/cipher"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestHTTPServerGetServerPublicKeys(t *testing.T) {
	_, err := GetServerPublicKey()
	if err != nil {
		t.Error(err)
	}
}

func GetServerPublicKey() (*btcec.PublicKey, error) {
	resp, err := http.Get("http://localhost:8080/v1/serverPublicKeys")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rsp map[string]string
	err = json.Unmarshal(body, &rsp)
	if err != nil {
		return nil, err
	}

	channelPubKeyServerString := rsp["publicKey"]
	bs, err := hex.DecodeString(channelPubKeyServerString)
	if err != nil {
		return nil, err
	}

	// Verifying the receiving data is a ecdsa publicKey
	pubKey, err := btcec.ParsePubKey(bs, btcec.S256())
	if err != nil {
		return nil, err
	}

	return pubKey, err
}

func TestHTTPServerGenPublicKeyAndSegWitAddress(t *testing.T) {
	serverPubECKey, err := GetServerPublicKey()
	if err != nil {
		t.Error(err)
	}

	// read seed file
	workingDir, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	var filePath = workingDir + "/test.json"
	keyParam := ReadSeedFromJsonFile(&filePath)
	marshalledData, err := json.Marshal(keyParam)
	if err != nil {
		t.Error(err)
	}

	channelPrivKeyClient, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		t.Error(err)
	}

	var slice []byte
	slice = append(channelPrivKeyClient.PubKey().SerializeCompressed(), marshalledData...)

	ciphertext, err := cipher.MessageEncrypt(serverPubECKey, &slice)
	if err != nil {
		t.Error(err)
	}

	data := make(map[string]string)
	data["data"] = hex.EncodeToString(*ciphertext)
	bytesData, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST","http://localhost:8080/v1/genPublicKeyAndSegWitAddress", bytes.NewReader(bytesData))
	if err != nil {
		t.Error(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	plaintext, err := cipher.MessageDecrypt(channelPrivKeyClient, &body)
	if err != nil {
		t.Error(err)
	}

	var rsp map[string]string
	err = json.Unmarshal(*plaintext, &rsp)
	if err != nil {
		t.Error(err)
	}

	publicKey := rsp["publicKey"]
	segwitAddress := rsp["segwitAddress"]

	if publicKey != "02f9cef06660ba26dcb605c33e2ec7e389e71095142f9593c3fa34a1e5ed81b26e" {
		t.Error("Unmatched public key")
	}

	if segwitAddress != "bc1q8c87x4v0m3dfrxksv724rtwpxy5ghpw8gwf8da" {
		t.Error("Unmatched segwitAddress")
	}
}

func TestGenerateSegwitAddress(t *testing.T) {
	publickey := "0279BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798"

	keyBytes, err := hex.DecodeString(publickey)
	if err != nil {
		t.Error(err)
	}

	segwitAddress, err := GenerateSegwitAddress(&keyBytes)
	if err != nil {
		t.Error(err)
	}

	if *segwitAddress != "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4" {
		t.Error("Unmatched segwit address")
	}
}

// ReadSeedFromJsonFile a helper function to read the json file to a BIP32PARAM instance
func ReadSeedFromJsonFile(file *string) *BIP32PARAM  {
	data, err := ioutil.ReadFile(*file)
	if err != nil {
		fmt.Println("File reading error", err)
		return nil
	}

	obj := BIP32PARAM{}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		fmt.Println("Json object unmarshal error", err)
		return nil
	}

	return &obj
}
