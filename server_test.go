package main

import (
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcd/btcec"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestHTTPServerGetServerPublicKeys(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/v1/serverPublicKeys")
	if err != nil {
		t.Error(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	var rsp map[string]string
	err = json.Unmarshal(body, &rsp)
	if err != nil {
		t.Error(err)
	}

	channelPubKeyServerString := rsp["publicKey"]
	bs, err := hex.DecodeString(channelPubKeyServerString)
	if err != nil {
		t.Error(err)
	}

	// Verifying the receiving data is a ecdsa publicKey
	_, err = btcec.ParsePubKey(bs, btcec.S256())
	if err != nil {
		t.Error(err)
	}
}
