package pb

import "github.com/pion/randutil"

const (
	runesAlpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	runesDigit = "0123456789"

	lenUFrag = 16
	lenPwd   = 32
)

func NewCredentials() Credentials {
	ufrag, err := randutil.GenerateCryptoRandomString(lenUFrag, runesAlpha)
	if err != nil {
		panic(err)
	}

	pwd, err := randutil.GenerateCryptoRandomString(lenPwd, runesAlpha)
	if err != nil {
		panic(err)
	}

	return Credentials{
		Ufrag: ufrag,
		Pwd:   pwd,
	}
}
