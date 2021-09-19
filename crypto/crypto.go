package crypto

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	ellcurv "github.com/btcsuite/btcd/btcec"
)

// SHA256
func SHA256BytesFromBytes(data []byte) []byte {
	hashHandler := sha256.New()
	hashHandler.Write(data)
	return hashHandler.Sum(nil)
}

func SHA256HexFromBytes(data []byte) string {
	return BytesToHex(SHA256BytesFromBytes(data))
}

// SHA512
func SHA512BytesFromBytes(data []byte) []byte {
	hashHandler := sha512.New()
	hashHandler.Write(data)
	return hashHandler.Sum(nil)
}

func SHA512BytesFromHex(data string) []byte {
	return SHA512BytesFromBytes(HexToBytes(data))
}

func SHA512HexFromBytes(data []byte) string {
	return BytesToHex(SHA512BytesFromBytes(data))
}

func SHA512HexFromHex(data string) string {
	return BytesToHex(SHA512BytesFromHex(data))
}

// Ellcurv
func Encryption(publicKey, data string) ([]byte, error) {
	pub, err := ellcurv.ParsePubKey(HexToBytes(publicKey), ellcurv.S256())
	if err != nil {
		return nil, err
	}
	result, err := ellcurv.Encrypt(pub, HexToBytes(data))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func Decryption(privateKey, data string) ([]byte, error) {
	priv, _ := ellcurv.PrivKeyFromBytes(ellcurv.S256(), HexToBytes(privateKey))
	result, err := ellcurv.Decrypt(priv, HexToBytes(data))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// AES
func GenerateRandomAESKey(len int) []byte {
	if len != 128 && len != 192 && len != 256 {
		panic(aes.KeySizeError(len))
	}
	key := make([]byte, len/8)
	_, err := rand.Read(key)
	if err != nil {
		panic(err.Error())
	}
	return key
}

// Convert between Hex and Bytes
func HexToBytes(str string) []byte {
	data, _ := hex.DecodeString(str)
	return data
}

func BytesToHex(data []byte) string {
	return hex.EncodeToString(data)
}
