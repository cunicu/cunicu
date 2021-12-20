package crypto_test

import (
	"testing"

	"riasc.eu/wice/pkg/crypto"
)

type Person struct {
	Name      string
	Age       int
	Children  []Person
	Signature string
}

func TestJWSCT(t *testing.T) {
	einstein := Person{
		Name: "Albert Einstein",
		Age:  66,
		Children: []Person{
			{
				Name: "Yoda",
				Age:  9999,
			},
		},
	}

	sk, err := crypto.ParseKey("GMHOtIxfUrGmncORjYK/slCSK/8V2TF9MjzzoPDTkEc=")
	if err != nil {
		panic(err)
	}

	pk, err := crypto.ParseKey("Hxm0/KTFRGFirpOoTWO2iMde/gJX+oVswUXEzVN5En8=")
	if err != nil {
		panic(err)
	}

	einstein.Signature, err = crypto.JWSCTSign(&einstein, sk)
	if err != nil {
		t.Errorf("Failed to sign: %s", err)
	}

	sig := einstein.Signature
	einstein.Signature = ""

	t.Logf("Signature: %s", sig)

	match, err := crypto.JWSCTVerify(&einstein, sig, pk)
	if err != nil {
		t.Errorf("Failed to verify: %s", err)
	}

	if !match {
		t.Errorf("Signature mismatch")
	}

	einstein.Age = 67

	match, err = crypto.JWSCTVerify(&einstein, sig, pk)
	if err != nil {
		t.Errorf("Failed to verify: %s", err)
	}

	if match {
		t.Errorf("Signature false positive")
	}
}
