package links

import "math/rand/v2"

const (
	alphabet   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	codeLength = 6
)

func NewCode() string {
	b := make([]byte, codeLength)

	for i := range codeLength {
		b[i] = alphabet[rand.IntN(len(alphabet))]
	}

	return string(b)
}
