package crypto

import "crypto/rand"

func GetNonce(len int) (Nonce, error) {
	var nonce = make(Nonce, len)

	_, err := rand.Read(nonce)
	if err != nil {
		return nonce, err
	}

	return nonce, nil
}
