package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"golang.org/x/crypto/ripemd160"

	bt "bytes"

	"github.com/akamensky/base58"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func main() {

	//fmt.Printf("% x\n", createPrivateKey())
	privateKey := createPrivateKey()
	publicKey := getPublicKey(privateKey)

	fmt.Println(hex.EncodeToString(privateKey[:]))
	fmt.Println(hex.EncodeToString(publicKey[:]))
	vsAdress := getVsAddress(publicKey)
	fmt.Println(vsAdress)

	//gammel3 := wifToPrivateKey("5JiQp3uu1nxYzvn1Agr7z7bhphjCcpK8qRfCM7p9oJpz9tMcAQh")
	//fmt.Println(hex.EncodeToString(gammel3[:]))
}

func getVsAddress(publicKey [65]byte) string {
	firstHash := sha256.Sum256(publicKey[:])
	h := ripemd160.New()
	h.Write(firstHash[:])
	secondHash := h.Sum(nil)
	//secondHash := ripemd160.New().Sum(firstHash[:])

	return bytesToBase58Check(secondHash[:], 0x00)
}

func getPublicKey(privateKey [32]byte) [65]byte {
	x, y := secp256k1.S256().ScalarBaseMult(privateKey[:])
	var publicKey [65]byte
	//Versionspräfix VS-Adresse
	publicKey[0] = 0x04
	copy(publicKey[1:33], x.Bytes())
	copy(publicKey[33:65], y.Bytes())
	return publicKey
}

// n = 1,158.. *10^77
var nMinusOneBytes = [32]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE, 0xBA, 0xAE, 0xDC, 0xE6, 0xAF, 0x48, 0xA0, 0x3B, 0xBF, 0xD2, 0x5E, 0x8C, 0xD0, 0x36, 0x41, 0x40}
var nMinusOne = new(big.Int).SetBytes(nMinusOneBytes[:])

func createPrivateKey() [32]byte {
	for {
		//1. create random 256 bits
		var randomBytes [32]byte
		rand.Read(randomBytes[:])

		//2. create hash of random bits
		key := sha256.Sum256(randomBytes[:])

		//3. key < n-1, otherwise create e new key
		keyNum := new(big.Int).SetBytes(key[:])
		if keyNum.Cmp(nMinusOne) == -1 {
			return key
		}
	}
}

func privateKeyToWif(privateKey [32]byte) string {
	return bytesToBase58Check(privateKey[:], 0x80)
}

func wifToPrivateKey(wif string) [32]byte {
	bytes, version := base58CheckToBytes(wif)

	if version != 0x80 || len(bytes) != 32 {
		//TODO fehlerbehandlung
		fmt.Println("Ist kein WIF!")
	}

	return [32]byte(bytes)
}

func bytesToBase58Check(bytes []byte, version byte) string {

	together := make([]byte, 0, 1+len(bytes)+4)

	//1. part: version
	together = append(together, version)

	//2. part: payload
	together = append([]byte{version}, bytes...)

	//3. part: checksum
	together = append(together, getFirstFourChecksumBytes([]byte{version}, bytes)...)

	return base58.Encode(together)
}

func base58CheckToBytes(input string) ([]byte, byte) {
	bytes, err := base58.Decode(input)
	if err != nil {
		//TODO Fehlerbehandlung
		fmt.Println("Error decoding base58")
	}
	version := bytes[0]
	payload := bytes[1 : len(bytes)-4]
	checksumBytes := bytes[len(bytes)-4:]

	if !bt.Equal(checksumBytes, getFirstFourChecksumBytes([]byte{version}, payload)) {
		//TODO Fehlerbehandlung
		fmt.Println("base58 check failed")
	}
	return payload, version
}

func getFirstFourChecksumBytes(bytes ...[]byte) []byte {
	h := sha256.New()
	for _, bytes := range bytes {
		h.Write(bytes)
	}
	firstHash := h.Sum(nil)
	secondHash := sha256.Sum256(firstHash)

	return secondHash[:4]
}
