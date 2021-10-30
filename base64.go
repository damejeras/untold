package untold

import (
	"encoding/base64"
	"errors"
)

var (
	base64Encoding = base64.RawStdEncoding
)

func Base64Decode(input []byte) ([]byte, error) {
	result := make([]byte, base64Encoding.DecodedLen(len(input)))
	_, err := base64Encoding.Decode(result, input)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func Base64Encode(input []byte) []byte {
	result := make([]byte, base64Encoding.EncodedLen(len(input)))
	base64Encoding.Encode(result, input)

	return result
}

func DecodeBase64Key(encodedKey []byte) (decodedKey [32]byte, err error) {
	var publicKey []byte

	publicKey, err = Base64Decode(encodedKey)
	if err != nil {
		return
	}

	if len(publicKey) != 32 {
		err = errors.New("corrupted key")

		return
	}

	copy(decodedKey[:], publicKey)

	return
}
