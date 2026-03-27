package crypto_test

import (
	"testing"

	"github.com/aoaYaoa/go-gin-starter/pkg/utils/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testKeys = []struct {
	name string
	key  []byte
}{
	{"AES-128", []byte("0123456789abcdef")},
	{"AES-192", []byte("0123456789abcdef01234567")},
	{"AES-256", []byte("0123456789abcdef0123456789abcdef")},
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	for _, tc := range testKeys {
		t.Run(tc.name, func(t *testing.T) {
			plaintext := "hello, world!"
			encoded, err := crypto.EncryptString(tc.key, plaintext)
			require.NoError(t, err)
			assert.NotEqual(t, plaintext, encoded)

			decoded, err := crypto.DecryptString(tc.key, encoded)
			require.NoError(t, err)
			assert.Equal(t, plaintext, decoded)
		})
	}
}

func TestEncryptProducesUniqueCiphertexts(t *testing.T) {
	key := []byte("0123456789abcdef")
	a, err := crypto.EncryptString(key, "same")
	require.NoError(t, err)
	b, err := crypto.EncryptString(key, "same")
	require.NoError(t, err)
	// GCM uses a random nonce each time — ciphertexts must differ
	assert.NotEqual(t, a, b)
}

func TestDecryptInvalidData(t *testing.T) {
	key := []byte("0123456789abcdef")
	_, err := crypto.DecryptString(key, "not-valid-base64!!")
	assert.Error(t, err)

	_, err = crypto.DecryptString(key, "dG9vc2hvcnQ=") // "tooshort" in base64
	assert.Error(t, err)
}

func TestInvalidKeyLength(t *testing.T) {
	_, err := crypto.EncryptString([]byte("shortkey"), "data")
	assert.Error(t, err)
}
