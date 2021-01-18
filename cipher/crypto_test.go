package cipher

import (
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"reflect"
	"testing"
)

func TestMessageEncryptDecrypt(t *testing.T) {

	privateKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		t.Error("Generate key error:", err)
	}

	msg := []byte("test message")
	ciphertext, err := MessageEncrypt(privateKey.PubKey(), &msg)
	if err != nil {
		t.Error("MessageEncrypt error:", err)
	}

	plaintext, err := MessageDecrypt(privateKey, ciphertext)
	if err != nil {
		t.Error("MessageDecrypt error:", err)
	}

	if !reflect.DeepEqual(msg, *plaintext) {
		t.Error("MessageDecrypt failed:", "\n", "original: ", fmt.Sprint(msg), "\n", "decrypted:", fmt.Sprint(*plaintext))
	}
}