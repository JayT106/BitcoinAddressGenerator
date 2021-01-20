package main

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"reflect"
)

type KEYPATH struct {
	ACCOUNT uint32
	CHAIN uint32
	ADDRESS uint32
}

type BIP32PARAM struct {
	SEED string
	PATH KEYPATH
}

// Clear clear the data of a instance especially the importance data like a seed, reduce the possibilities of the malware attack
func Clear(v interface{}) {
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
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