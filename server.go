package main

import (
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcd/btcec"
	"log"
	"net/http"
)
func main() {

	// Generate the key for the data encrypt/decrypt during the message passing
	privKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}

	//Create the default mux
	mux := http.NewServeMux()

	//Handling the /v1/serverPublicKeys.
	pubkh := &PubKeyHandler{privKey.PubKey()}
	mux.Handle("/v1/serverPublicKeys", pubkh)

	//Create the http server.
	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start the server
	err = s.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

// PubKeyHandler the handler uses for passing this struct into the ServerHTTP function
type PubKeyHandler struct {
	pubKey *btcec.PublicKey
}

// ServeHTTP handle the V1/serverPublicKeys API request. Return the server public key for encrypting the client data
func (ph *PubKeyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["publicKey"] = hex.EncodeToString(ph.pubKey.SerializeCompressed())
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Println("Json Marshal error:", err)
		w.WriteHeader(500)
	} else {
		w.WriteHeader(200)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonResp)
	if err != nil {
		log.Println("ServeHTTP write error:", err)
	}
}

