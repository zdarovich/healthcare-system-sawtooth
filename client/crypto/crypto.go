package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"healthcare-system-sawtooth/client/db/models"
	"healthcare-system-sawtooth/client/lib"
	"healthcare-system-sawtooth/crypto"
	tpStorage "healthcare-system-sawtooth/tp/storage"
)

// GenerateDataInfo generate the information of data for storage system.
func GenerateDataInfo(name, target, publicKey, username, keyAes string, accessType uint, expiration int64) (info tpStorage.DataInfo, err error) {

	keyEncrypt, err := crypto.Encryption(publicKey, keyAes)
	if err != nil {
		return
	}
	hash, out, err := EncryptData([]byte(target), crypto.HexToBytes(keyAes))
	if err != nil {
		return
	}
	data := &models.Data{
		Name:       name,
		Hash:       hash,
		Payload:    crypto.BytesToHex(out),
		Expiration: expiration,
	}
	_, err = data.Save()
	if err != nil {
		return
	}
	info = tpStorage.DataInfo{
		Name:       name,
		Size:       int64(len(target)),
		Hash:       hash,
		Addr:       username,
		Key:        crypto.BytesToHex(keyEncrypt),
		AccessType: accessType,
	}
	return
}

// GenerateSharedDataInfo generates the information of shared data for storage system.
func GenerateSharedDataInfo(name, publicKey, username, keyAes, hash string, size int64) (info tpStorage.DataInfo, err error) {

	keyEncrypt, err := crypto.Encryption(publicKey, keyAes)
	if err != nil {
		return
	}

	info = tpStorage.DataInfo{
		Name: name,
		Size: size,
		Hash: hash,
		Addr: username,
		Key:  crypto.BytesToHex(keyEncrypt),
	}
	return
}

// EncryptData encrypt the data using AES-CTR. After encryption, calculate the hash of data.
func EncryptData(in, keyAes []byte) (hash string, out []byte, err error) {

	iv := make([]byte, lib.IvSize)
	_, err = rand.Read(iv)
	if err != nil {
		return
	}
	block, err := aes.NewCipher(keyAes)
	if err != nil {
		return
	}
	ctr := cipher.NewCTR(block, iv)

	hashes := make([][]byte, 0)
	outBuf := make([]byte, len(in))

	ctr.XORKeyStream(outBuf, in)
	hashes = append(hashes, crypto.SHA512BytesFromBytes(outBuf))

	out = append(iv, outBuf...)
	hash = crypto.SHA512HexFromBytes(bytes.Join(hashes, []byte{}))
	return
}

// DecryptData decrypt the data using AES-CTR. After decryption, calculate the hash of data.
func DecryptData(in, key []byte) (hash string, out []byte, err error) {

	iv := in[:lib.IvSize]
	enc := in[lib.IvSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", nil, err
	}
	ctr := cipher.NewCTR(block, iv)

	hashes := make([][]byte, 0)
	outBuf := make([]byte, len(enc))

	ctr.XORKeyStream(outBuf, enc)
	hashes = append(hashes, crypto.SHA512BytesFromBytes(outBuf))

	out = outBuf
	hash = crypto.SHA512HexFromBytes(bytes.Join(hashes, []byte{}))
	return
}

// CalDataHash calculate the hash of data.
func CalDataHash(data string) (hash string, err error) {
	hash = crypto.SHA512HexFromBytes([]byte(data))
	return
}
