package untold

import "encoding/base64"

var (
	Base64Encoding = base64.RawStdEncoding
)

func Base64Decode(input []byte) ([]byte, error) {
	result := make([]byte, Base64Encoding.DecodedLen(len(input)))
	_, err := Base64Encoding.Decode(result, input)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func Base64Encode(input []byte) []byte {
	result := make([]byte, Base64Encoding.EncodedLen(len(input)))
	Base64Encoding.Encode(result, input)

	return result
}
