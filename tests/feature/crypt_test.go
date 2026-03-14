package feature

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
)

type CryptTestSuite struct {
	suite.Suite
}

func TestCryptTestSuite(t *testing.T) {
	suite.Run(t, new(CryptTestSuite))
}

func (s *CryptTestSuite) TestEncryptAndDecryptString() {
	values := []string{
		"goravel",
		"",
		"你好，Goravel 🚀",
	}

	for _, value := range values {
		s.Run(value, func() {
			encrypted, err := facades.Crypt().EncryptString(value)
			s.NoError(err)
			s.NotEmpty(encrypted)
			s.NotEqual(value, encrypted)

			decrypted, err := facades.Crypt().DecryptString(encrypted)
			s.NoError(err)
			s.Equal(value, decrypted)
		})
	}
}

func (s *CryptTestSuite) TestEncryptStringIsRandomized() {
	first, err := facades.Crypt().EncryptString("goravel")
	s.NoError(err)
	s.NotEmpty(first)

	second, err := facades.Crypt().EncryptString("goravel")
	s.NoError(err)
	s.NotEmpty(second)

	s.NotEqual(first, second)
}

func (s *CryptTestSuite) TestDecryptStringErrors() {
	_, err := facades.Crypt().DecryptString("invalid-payload")
	s.Error(err)

	invalidWithoutIV, err := json.Marshal(map[string][]byte{
		"value": []byte("ciphertext"),
	})
	s.NoError(err)

	_, err = facades.Crypt().DecryptString(base64.StdEncoding.EncodeToString(invalidWithoutIV))
	s.Error(err)
	s.ErrorContains(err, "iv")

	invalidWithoutValue, err := json.Marshal(map[string][]byte{
		"iv": []byte("nonce"),
	})
	s.NoError(err)

	_, err = facades.Crypt().DecryptString(base64.StdEncoding.EncodeToString(invalidWithoutValue))
	s.Error(err)
	s.ErrorContains(err, "value")

	encrypted, err := facades.Crypt().EncryptString("goravel")
	s.NoError(err)

	decoded, err := base64.StdEncoding.DecodeString(encrypted)
	s.NoError(err)

	var payload map[string][]byte
	s.NoError(json.Unmarshal(decoded, &payload))
	s.NotEmpty(payload["value"])
	payload["value"][0] ^= 0xFF

	tamperedPayload, err := json.Marshal(payload)
	s.NoError(err)

	_, err = facades.Crypt().DecryptString(base64.StdEncoding.EncodeToString(tamperedPayload))
	s.Error(err)
}
