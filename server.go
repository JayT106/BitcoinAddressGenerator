package main

import (
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/jayt106/bitcoinAddressGenerator/cipher"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	//Handling the /v1/genPublicKeyAndSegWitAddress.
	privkh := &PrivKeyHandler{privKey}
	mux.Handle("/v1/genPublicKeyAndSegWitAddress", privkh)

	//Handling the /v1/genMultiSigP2SH address
	mux.HandleFunc("/v1/genMultiSigP2SHAddress", GenMultiSigP2SHAddress)

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

type PrivKeyHandler struct {
	privKey *btcec.PrivateKey
}

// ServeHTTP handle the V1/genPublicKeyAndSegWitAddress API request.
// The http client send the seed, path and the public key(for the return message encryption) encrypted by the server's public key
// (See V1/serverPublicKeys API). This function decrypted the message by the server's private key, generate the HD key base on the
// seed and the path and return the public key and the SegWit address encrypted by the client's public key.
func (ph *PrivKeyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ServerErrorHandle(w, err, "Read body error:")
		return
	}

	msgParam := make(map[string]string)
	err = json.Unmarshal(body, &msgParam)
	if err != nil {
		ServerErrorHandle(w, err, "Json unmarshal error:")
		return
	}

	cipherBytes, err := hex.DecodeString(msgParam["data"])
	if err != nil {
		ServerErrorHandle(w, err, "Hex decode string error:")
		return
	}

	plainBytes, err := cipher.MessageDecrypt(ph.privKey, &cipherBytes)
	if err != nil {
		ServerErrorHandle(w, err, "Decrypt data error:")
		return
	}

	slice := *plainBytes
	clientCipherPublicKey := slice[:btcec.PubKeyBytesLenCompressed]
	keyPath := slice[btcec.PubKeyBytesLenCompressed:]
	Clear(plainBytes)
	Clear(&slice)

	var keyParam BIP32PARAM
	err = json.Unmarshal(keyPath, &keyParam)
	Clear(&keyPath)
	if err != nil {
		ServerErrorHandle(w, err, "Unmarshal data error:")
		return
	}

	// Generate a HD key chain (Bitcoin mainnet) using the seed.
	clientHDPubKey, err := GenerateHDPublicKey(&keyParam)
	Clear(&keyParam)
	if err != nil {
		ServerErrorHandle(w, err, "Generate HD public key failed:")
		return
	}

	compressedPubKey, err := ConvertPublicKey(clientHDPubKey)
	if err != nil {
		ServerErrorHandle(w, err, "Convert HD public key failed:")
		return
	}

	segwitAddress, err := GenerateSegwitAddress(compressedPubKey)
	if err != nil {
		ServerErrorHandle(w, err, "Generate segwit address failed:")
		return
	}

	resp := make(map[string]string)
	resp["publicKey"] = hex.EncodeToString(*compressedPubKey)
	resp["segwitAddress"] = *segwitAddress

	marshalledData, err := json.Marshal(resp)
	if err != nil {
		ServerErrorHandle(w, err, "Json Marshal error:")
		return
	}

	pubKey, err := btcec.ParsePubKey(clientCipherPublicKey, btcec.S256())
	if err != nil {
		ServerErrorHandle(w, err, "ParsePubKey error:")
		return
	}

	cipherText, err := cipher.MessageEncrypt(pubKey, &marshalledData)
	if err != nil {
		ServerErrorHandle(w, err, "MessageEncrypt error:")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_, err = w.Write(*cipherText)
	if err != nil {
		log.Println("ServeHTTP write error:", err)
	}
}

// GenerateSegwitAddress Generate segwit address using for the bitcoin mainnet by the public key
func GenerateSegwitAddress(key *[]byte) (*string, error) {
	witnessProg := btcutil.Hash160(*key)
	addressWitnessPubKeyHash, err := btcutil.NewAddressWitnessPubKeyHash(witnessProg, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}

	segwitAddress := addressWitnessPubKeyHash.EncodeAddress()
	return &segwitAddress, nil
}

// ConvertPublicKey Serialize the HD key struct to a compressed public key data represent by a byte array
func ConvertPublicKey(key *hdkeychain.ExtendedKey) (*[]byte, error) {
	ecPubKey, err := key.ECPubKey()
	if err != nil {
		return nil, err
	}

	compressed := ecPubKey.SerializeCompressed()
	return &compressed, nil
}

// ServerErrorHandle Handle the response message when the error happens during the HTTP request processing
func ServerErrorHandle(w http.ResponseWriter, e error, s string) {
	log.Println(s, e)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)
}

// GenerateHDPublicKey Generate a bitcoin mainnet HD public key given the seed and path following by BIP032
func GenerateHDPublicKey(p *BIP32PARAM) (*hdkeychain.ExtendedKey, error){
	seed, err := hex.DecodeString(p.SEED)
	if err != nil {
		return nil, err
	}

	clientMasterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	Clear(&seed)
	if err != nil {
		return nil, err
	}

	accKey, err := clientMasterKey.Derive(hdkeychain.HardenedKeyStart + p.PATH.ACCOUNT)
	Clear(&clientMasterKey)
	if err != nil {
		return nil, err
	}

	chainKey, err := accKey.Derive(p.PATH.CHAIN)
	if err != nil {
		return nil, err
	}

	addressKey, err := chainKey.Derive(p.PATH.ADDRESS)
	if err != nil {
		return nil, err
	}

	clientHDPubKey, err := addressKey.Neuter()
	if err != nil {
		return nil, err
	}

	return clientHDPubKey, nil
}

// HandleMultiSigP2SHAddress a handle function to genarate the n-out-of-m MultiSig P2SH bitcoin Address
func GenMultiSigP2SHAddress(w http.ResponseWriter, r *http.Request)  {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ServerErrorHandle(w, err, "Read body error:")
		return
	}

	var msgParam map[string]string
	err = json.Unmarshal(body, &msgParam)
	if err != nil {
		ServerErrorHandle(w, err, "Json unmarshal error:")
		return
	}

	n, err := strconv.ParseInt(msgParam["n"], 10, 32)
	if err != nil {
		ServerErrorHandle(w, err, "The argument n parsing error:")
		return
	}
	m, err := strconv.ParseInt(msgParam["m"], 10, 32)
	if err != nil {
		ServerErrorHandle(w, err, "The argument m parsing error:")
		return
	}
	publicKeys := msgParam["publicKeys"]

	// the client input requirement is n-of-m multisig. Therefore, the order of the param for calling the following function
	// need to be careful
	P2SHAddress, redeemScriptHex, err := cipher.OutputAddress(int(n), int(m), publicKeys)

	resp := make(map[string]string)
	if err != nil {
		errString := err.Error()
		if !strings.Contains(errString, "WARNING:") {
			ServerErrorHandle(w, err, "P2SH Address generating error:")
			return
		}
		resp["warning"] = errString
	}

	resp["ps2hAddress"] = P2SHAddress
	resp["redeemScriptHex"] = redeemScriptHex

	marshalledData, err := json.Marshal(resp)
	if err != nil {
		log.Println("Json Marshal error:", err)
		w.WriteHeader(500)
	} else {
		w.WriteHeader(200)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(marshalledData)
	if err != nil {
		log.Println("ServeHTTP write error:", err)
	}
}