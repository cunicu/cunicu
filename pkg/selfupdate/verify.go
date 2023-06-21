// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package selfupdate

// derived from http://github.com/restic/restic

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"

	pgp "golang.org/x/crypto/openpgp" //nolint:staticcheck
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/stv0g/cunicu/pkg/proto"
)

//go:embed keys/*.gpg
var keys embed.FS

var errVersionMismatch = errors.New("version mismatch")

func loadKeyRing() (pgp.EntityList, error) {
	el := pgp.EntityList{}

	des, err := keys.ReadDir("keys")
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	for _, de := range des {
		fn := filepath.Join("keys", de.Name())
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

func VersionVerify(binaryFile, expectedVersion string) error {
	cmd := exec.Command(binaryFile, "version", "--format=json")

	out, err := cmd.Output()
	if err != nil {
		return err
	}

	bi := &proto.BuildInfos{}
	mo := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: true,
	}

	if err := mo.Unmarshal(out, bi); err != nil {
		return err
	}

	if "v"+expectedVersion != bi.Client.Version {
		return fmt.Errorf("%w: dowloaded %s != expected v%s", errVersionMismatch, bi.Client.Version, expectedVersion)
	}

	return nil
}
