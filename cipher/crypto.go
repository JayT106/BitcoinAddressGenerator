package cipher

import (
	"fmt"
	"github.com/btcsuite/btcd/btcec"
)

// MessageEncrypt encrypts data for the target public key using AES-256-CBC.
func MessageEncrypt(pubKey *btcec.PublicKey, plainText *[]byte) (*[]byte, error) {
	ciphertext, err := btcec.Encrypt(pubKey, *plainText)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &ciphertext, nil
}

// MessageDecrypt decrypts data that was encrypted using the Encrypt function.
func MessageDecrypt(privKey *btcec.PrivateKey, ciphertext *[]byte) (*[]byte, error) {
	plainText, err := btcec.Decrypt(privKey, *ciphertext)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &plainText, nil
}
