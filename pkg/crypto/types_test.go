package crypto_test

import (
	"encoding/json"
	"net"
	"testing"

	"riasc.eu/wice/pkg/crypto"
)

func TestGenerateKeyFromPassword(t *testing.T) {
	key1 := crypto.GenerateKeyFromPassword("test")
	key2 := crypto.GenerateKeyFromPassword("test2")

	expectedKey1, err := crypto.ParseKey("SAyMLIWTO+DSnTx/JDak+lRR5huci8m4JsEabkkIxFY=")
	if err != nil {
		t.FailNow()
	}

	if key1 != expectedKey1 {
		t.Fail()
	}

	if key1 == key2 {
		t.Fail()
	}

	if len(key1) != crypto.KeyLength {
		t.Fail()
	}

	if !key1.IsSet() {
		t.Fail()
	}

	if _, err := crypto.ParseKeyBytes(key1[:]); err != nil {
		t.Fail()
	}
}

func TestGenerateKey(t *testing.T) {
	key1, err := crypto.GenerateKey()
	if err != nil {
		t.Fail()
	}

	key2, err := crypto.GenerateKey()
	if err != nil {
		t.Fail()
	}

	if !key1.IsSet() || !key2.IsSet() {
		t.Fail()
	}

	if key1 == key2 {
		t.Fail()
	}
}
func TestGeneratePrivateKey(t *testing.T) {
	key1, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.Fail()
	}

	key2, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.Fail()
	}

	if !key1.IsSet() || !key2.IsSet() {
		t.Fail()
	}

	if key1 == key2 {
		t.Fail()
	}
}

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

func TestKeyBytes(t *testing.T) {
	key1, err := crypto.GenerateKey()
	if err != nil {
		t.Fail()
	}

	if key2, err := crypto.ParseKeyBytes(key1.Bytes()); err != nil || key2 != key1 {
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

	objJSON, err := json.Marshal(&obj1)
	if err != nil {
		t.Fail()
	}

	if err := json.Unmarshal(objJSON, &obj2); err != nil {
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

func TestKeyIsSet(t *testing.T) {
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

func TestShared(t *testing.T) {
	key1, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.Fail()
	}

	key2, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.Fail()
	}

	kp1 := crypto.KeyPair{
		Ours:   key1,
		Theirs: key2.PublicKey(),
	}

	kp2 := crypto.KeyPair{
		Ours:   key2,
		Theirs: key1.PublicKey(),
	}

	if kp1.Shared() != kp2.Shared() {
		t.Fail()
	}
}

func TestIPv6Address(t *testing.T) {
	key, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.FailNow()
	}

	addr := key.PublicKey().IPv6Address()

	_, ll, err := net.ParseCIDR("fe80::/10")
	if err != nil {
		t.FailNow()
	}

	if ones, bits := addr.Mask.Size(); ones != 64 || bits != net.IPv6len*8 {
		t.Fail()
	}

	if !ll.Contains(addr.IP) {
		t.Fail()
	}
}
