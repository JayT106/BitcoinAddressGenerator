package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcd/btcec"
	"github.com/jayt106/bitcoinAddressGenerator/cipher"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var privKey, _ = btcec.NewPrivateKey(btcec.S256())

func GetServerPublicKey() (*btcec.PublicKey, error) {
	pubkh := &PubKeyHandler{privKey.PubKey()}
	resp, err := http.NewRequest("GET", "v1/serverPublicKeys", nil)
	if err != nil {
		return nil, err
	}

	rr := httptest.NewRecorder()
	http.Handle("/v1/serverPublicKeys", pubkh)
	pubkh.ServeHTTP(rr, resp)

	body, err := ioutil.ReadAll(rr.Body)
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

	var filePath = workingDir + "/../test/test.json"
	keyParam, err := ReadSeedFromJsonFile(&filePath)
	if err != nil {
		t.Error(err)
	}

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

	req, err := http.NewRequest("POST","v1/genPublicKeyAndSegWitAddress", bytes.NewReader(bytesData))
	if err != nil {
		t.Error(err)
	}

	privkh := &PrivKeyHandler{privKey}
	http.Handle("v1/genPublicKeyAndSegWitAddress", privkh)
	rr := httptest.NewRecorder()
	privkh.ServeHTTP(rr, req)

	body, err := ioutil.ReadAll(rr.Body)
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

func TestHTTPServerGenMultiSigP2SHAddress(t *testing.T) {
	data := make(map[string]string)
	data["n"] = "2"
	data["m"] = "3"
	data["publicKeys"] = "04a882d414e478039cd5b52a92ffb13dd5e6bd4515497439dffd691a0f12af9575fa349b5694ed3155b136f09e63975a1700c9f4d4df849323dac06cf3bd6458cd,046ce31db9bdd543e72fe3039a1f1c047dab87037c36a669ff90e28da1848f640de68c2fe913d363a51154a0c62d7adea1b822d05035077418267b1a1379790187,0411ffd36c70776538d079fbae117dc38effafb33304af83ce4894589747aee1ef992f63280567f52f5ba870678b4ab4ff6c8ea600bd217870a8b4f1f09f3a8e83"
	bytesData, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
	}

	req, err := http.NewRequest("POST","/v1/genMultiSigP2SHAddress", bytes.NewReader(bytesData))
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	http.HandleFunc("v1/genMultiSigP2SHAddress", GenMultiSigP2SHAddress)
	GenMultiSigP2SHAddress(rr, req)

	body, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Error(err)
	}

	var rsp map[string]string
	err = json.Unmarshal(body, &rsp)
	if err != nil {
		t.Error(err)
	}

	P2SHAddress := rsp["ps2hAddress"]
	redeemScriptHex := rsp["redeemScriptHex"]

	testAddress := "347N1Thc213QqfYCz3PZkjoJpNv5b14kBd"
	testRedeemScriptHex := "524104a882d414e478039cd5b52a92ffb13dd5e6bd4515497439dffd691a0f12af9575fa349b5694ed3155b136f09e63975a1700c9f4d4df849323dac06cf3bd6458cd41046ce31db9bdd543e72fe3039a1f1c047dab87037c36a669ff90e28da1848f640de68c2fe913d363a51154a0c62d7adea1b822d05035077418267b1a1379790187410411ffd36c70776538d079fbae117dc38effafb33304af83ce4894589747aee1ef992f63280567f52f5ba870678b4ab4ff6c8ea600bd217870a8b4f1f09f3a8e8353ae"
	if testAddress != P2SHAddress {
		t.Error(t, "Generated P2SH address different from expected address.", testAddress, P2SHAddress)
	}
	if testRedeemScriptHex != redeemScriptHex {
		t.Error(t, "Generated P2SH address different from expected address.", testRedeemScriptHex, redeemScriptHex)
	}
}
