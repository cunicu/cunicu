package crypto_test

import (
	"encoding/json"
	"testing"

	"riasc.eu/wice/pkg/crypto"
)

func TestKeyString(t *testing.T) {
	key1, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.Fail()
	}

	keyString := key1.String()

	key2, err := crypto.ParseKey(keyString)
	if err != nil {
		t.Fail()
	}

	if key1 != key2 {
		t.Fail()
	}
}

func TestGeneratePrivateKey(t *testing.T) {
	key1, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.Fail()
	}

	var zeroKey crypto.Key
	if key1 == zeroKey {
		t.Fail()
	}

	key2, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.Fail()
	}

	if key1 == key2 {
		t.Fail()
	}
}

type testObj struct {
	Key crypto.Key
}

func TestMarshal(t *testing.T) {
	key, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.Fail()
	}

	var obj1, obj2 testObj

	obj1 = testObj{
		Key: key,
	}

	objJson, err := json.Marshal(&obj1)
	if err != nil {
		t.Fail()
	}

	err = json.Unmarshal(objJson, &obj2)
	if err != nil {
		t.Fail()
	}

	if obj1 != obj2 {
		t.Fail()
	}
}

func TestPublicKey(t *testing.T) {
	sk, err := crypto.ParseKey("GMHOtIxfUrGmncORjYK/slCSK/8V2TF9MjzzoPDTkEc=")
	if err != nil {
		t.Fail()
	}

	pk, err := crypto.ParseKey("Hxm0/KTFRGFirpOoTWO2iMde/gJX+oVswUXEzVN5En8=")
	if err != nil {
		t.Fail()
	}

	if sk.PublicKey() != pk {
		t.Fail()
	}
}

func TestIsSet(t *testing.T) {
	key, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.Fail()
	}

	if !key.IsSet() {
		t.Fail()
	}

	key = crypto.Key{}

	if key.IsSet() {
		t.Fail()
	}
}
