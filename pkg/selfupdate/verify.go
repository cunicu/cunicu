package selfupdate

// derived from http://github.com/restic/restic

import (
	"bytes"
	"embed"
	"fmt"
	"path"

	//lint:ignore SA1019 We still need to find an alternative

	pgp "golang.org/x/crypto/openpgp"
)

//go:embed keys/*.gpg
var keys embed.FS

func loadKeyRing() (pgp.EntityList, error) {
	el := pgp.EntityList{}

	des, err := keys.ReadDir("keys")
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	for _, de := range des {
		fn := path.Join("keys", de.Name())
		f, err := keys.Open(fn)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}

		e, err := pgp.ReadKeyRing(f)
		if err != nil {
			fmt.Printf("reading keyring failed")
			return nil, err
		}

		el = append(el, e...)
	}

	return el, nil
}

// GPGVerify checks the authenticity of data by verifying the signature sig,
// which must be ASCII armored (base64). When the signature matches, GPGVerify
// returns true and a nil error.
func GPGVerify(data, sig []byte) (ok bool, err error) {
	keyring, err := loadKeyRing()
	if err != nil {
		return false, fmt.Errorf("failed to load keyring: %w", err)
	}

	_, err = pgp.CheckArmoredDetachedSignature(keyring, bytes.NewReader(data), bytes.NewReader(sig))
	if err != nil {
		return false, err
	}

	return true, nil
}
