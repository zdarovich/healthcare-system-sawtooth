package crypto

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"healthcare-system-sawtooth/client/lib"
	tpCrypto "healthcare-system-sawtooth/crypto"
)

var key = tpCrypto.GenerateRandomAESKey(lib.AESKeySize)

func TestEncryptFile(t *testing.T) {
	in := []byte("test")
	hash, out, err := EncryptData(in, key)
	if err != nil {
		t.Error(err)
	}
	t.Log(hash)

	hash, out, err = DecryptData(out, key)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, string(in), string(out))
	t.Log(hash)
}
