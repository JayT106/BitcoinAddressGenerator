package cipher

import (
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"log"
	"strings"
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

// Duplicate the multisig functions due to the project package issues
// Refrence: github.com/soroushjp/go-bitcoin-multisig/multisig
// OutputAddress formats and prints relevant outputs to the user.
func OutputAddress(flagM int, flagN int, flagPublicKeys string) (string, string, error) {
	P2SHAddress, redeemScriptHex, err := generateAddress(flagM, flagN, flagPublicKeys)
	if err != nil {
		return "", "", err
	}

	if flagM*73+flagN*66 > 496 {
		err := errors.New(fmt.Sprintf(`
-----------------------------------------------------------------------------------------------------------------------------------
WARNING: 
%d-of-%d multisig transaction is valid but *non-standard* for Bitcoin v0.9.x and earlier.
It may take a very long time (possibly never) for transaction spending multisig funds to be included in a block.
To remain valid, choose smaller m and n values such that m*73+n*66 <= 496, as per standardness rules.
See http://bitcoin.stackexchange.com/questions/23893/what-are-the-limits-of-m-and-n-in-m-of-n-multisig-addresses for more details.
------------------------------------------------------------------------------------------------------------------------------------
`,
			flagM,
			flagN,
		))

		return P2SHAddress, redeemScriptHex, err
	}

	return P2SHAddress, redeemScriptHex, nil
}

// Refrence: github.com/soroushjp/go-bitcoin-multisig/multisig
// GenerateAddress is the high-level logic for creating P2SH multisig addresses with the 'go-bitcoin-multisig address' subcommand.
// Takes flagM (number of keys required to spend), flagN (total number of keys)
// and flagPublicKeys (comma separated list of N public keys) as arguments.
func generateAddress(flagM int, flagN int, flagPublicKeys string) (string, string, error) {
	//Convert public keys argument into slice of public key bytes with necessary tidying
	flagPublicKeys = strings.Replace(flagPublicKeys, "'", "\"", -1) //Replace single quotes with double since csv package only recognizes double quotes
	publicKeyStrings, err := csv.NewReader(strings.NewReader(flagPublicKeys)).Read()
	if err != nil {
		log.Fatal(err)
		return "", "", err
	}
	publicKeys := make([][]byte, len(publicKeyStrings))
	for i, publicKeyString := range publicKeyStrings {
		publicKeyString = strings.TrimSpace(publicKeyString)   //Trim whitespace
		publicKeys[i], err = hex.DecodeString(publicKeyString) //Get private keys as slice of raw bytes
		if err != nil {
			log.Fatal(err, "\n", "Offending publicKey: \n", publicKeyString)
			return "", "", err
		}
	}
	//Create redeemScript from public keys
	redeemScript, err := newMOfNRedeemScript(flagM, flagN, publicKeys)
	if err != nil {
		log.Fatal(err)
		return "", "", err
	}
	redeemScriptHash := btcutil.Hash160(redeemScript)

	//Get P2SH address by base58 encoding with P2SH prefix 0x05
	P2SHAddress := base58.CheckEncode(redeemScriptHash, 5)

	//Get redeemScript in Hex
	redeemScriptHex := hex.EncodeToString(redeemScript)

	return P2SHAddress, redeemScriptHex, nil
}

// Refrence: github.com/soroushjp/go-bitcoin-multisig/btcutils
// newMOfNRedeemScript creates a M-of-N Multisig redeem script given m, n and n public keys
func newMOfNRedeemScript(m int, n int, publicKeys [][]byte) ([]byte, error) {
	//Check we have valid numbers for M and N
	if n < 1 || n > 7 {
		return nil, errors.New("N must be between 1 and 7 (inclusive) for valid, standard P2SH multisig transaction as per Bitcoin protocol.")
	}
	if m < 1 || m > n {
		return nil, errors.New("M must be between 1 and N (inclusive).")
	}
	//Check we have N public keys as necessary.
	if len(publicKeys) != n {
		return nil, errors.New(fmt.Sprintf("Need exactly %d public keys to create P2SH address for %d-of-%d multisig transaction. Only %d keys provided.", n, m, n, len(publicKeys)))
	}
	//Get OP Code for m and n.
	//81 is OP_1, 82 is OP_2 etc.
	//80 is not a valid OP_Code, so we floor at 81
	mOPCode := 81 + (m - 1)
	nOPCode := 81 + (n - 1)
	//Multisig redeemScript format:
	//<OP_m> <A pubkey> <B pubkey> <C pubkey>... <OP_n> OP_CHECKMULTISIG
	var redeemScript bytes.Buffer
	redeemScript.WriteByte(byte(mOPCode)) //m
	for _, publicKey := range publicKeys {
		err := checkPublicKeyIsValid(publicKey)
		if err != nil {
			return nil, err
		}
		redeemScript.WriteByte(byte(len(publicKey))) //PUSH
		redeemScript.Write(publicKey)                //<pubkey>
	}
	redeemScript.WriteByte(byte(nOPCode)) //n
	redeemScript.WriteByte(byte(174))
	return redeemScript.Bytes(), nil
}

// Refrence: github.com/soroushjp/go-bitcoin-multisig/btcutils
// checkPublicKeyIsValid runs a couple of checks to make sure a public key looks valid.
// Returns an error with a helpful message or nil if key is valid.
func checkPublicKeyIsValid(publicKey []byte) error {
	errMessage := ""
	if publicKey == nil {
		errMessage += "Public key cannot be empty.\n"
	} else if len(publicKey) != 65 {
		errMessage += fmt.Sprintf("Public key should be 65 bytes long. Provided public key is %d bytes long.", len(publicKey))
	} else if publicKey[0] != byte(4) {
		errMessage += fmt.Sprintf("Public key first byte should be 0x04. Provided public key first byte is 0x%v.", hex.EncodeToString([]byte{publicKey[0]}))
	}
	if errMessage != "" {
		errMessage += "Invalid public key:\n"
		errMessage += hex.EncodeToString(publicKey)
		return errors.New(errMessage)
	}
	return nil
}
