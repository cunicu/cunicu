package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ucarion/jcs"
)

type JWK struct {
	KeyType string `json:"kty"`
	Curve   string `json:"crv"`
	X       string `json:"x"`
}

type JWS struct {
	Algorithm string `json:"alg"`
	Key       JWK    `json:"jwk"`
}

func jsonCanonicalize(obj interface{}) (string, error) {
	var objIntf interface{}

	objJson, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(objJson, &objIntf); err != nil {
		return "", err
	}

	msg, err := jcs.Format(objIntf)
	if err != nil {
		return "", err
	}

	return msg, nil
}

func JWSCTSign(obj interface{}, sk Key) (string, error) {
	var pk Key

	hdr := JWS{
		Algorithm: "XEdDSA-25519",
		Key: JWK{
			KeyType: "OKP",
			Curve:   "X25519",
			X:       pk.String(),
		},
	}

	msg, err := jsonCanonicalize(obj)
	if err != nil {
		return "", err
	}

	rnd := make([]byte, 32)

	_, err = rand.Reader.Read(rnd[:])
	if err != nil {
		panic(err)
	}

	sig := sk.Sign([]byte(msg), rnd)

	hdrBytes, err := json.Marshal(&hdr)
	if err != nil {
		return "", err
	}

	hdrBase64 := base64.URLEncoding.EncodeToString(hdrBytes)
	sigBase64 := base64.URLEncoding.EncodeToString(sig[:])
	plBase64 := "" // payload is always empty for JWS-CT

	return fmt.Sprintf("%s.%s.%s", hdrBase64, plBase64, sigBase64), nil
}

func JWSCTVerify(obj interface{}, jwsStr string, pk Key) (bool, error) {
	jwsStrParts := strings.Split(jwsStr, ".")
	if len(jwsStrParts) != 3 {
		return false, errors.New("invalid JWS format")
	}

	hdrBase64 := jwsStrParts[0]
	plBase64 := jwsStrParts[1]
	sigBase64 := jwsStrParts[2]

	if plBase64 != "" {
		return false, errors.New("payload field in JWS is not empty")
	}

	hdrBytes, err := base64.URLEncoding.DecodeString(hdrBase64)
	if err != nil {
		return false, err
	}

	sig, err := base64.URLEncoding.DecodeString(sigBase64)
	if err != nil {
		return false, err
	}

	var hdr JWS
	if err := json.Unmarshal(hdrBytes, &hdr); err != nil {
		return false, err
	}

	if hdr.Key.KeyType != "OKP" {
		return false, fmt.Errorf("unsupported key type: %s", hdr.Key.KeyType)
	}

	if hdr.Key.Curve != "X25519" {
		return false, fmt.Errorf("unsupported curve type: %s", hdr.Key.Curve)
	}

	var ssig Signature
	copy(ssig[:], sig)

	msg, err := jsonCanonicalize(obj)
	if err != nil {
		return false, err
	}

	return pk.Verify([]byte(msg), ssig), nil
}
