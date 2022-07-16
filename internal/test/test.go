package test

import (
	"bytes"
	"encoding/hex"
	"math"
	"math/rand"
	"net"
	"os"
	"path/filepath"

	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
)

func GenerateKeyPairs() (*crypto.KeyPair, *crypto.KeyPair, error) {
	ourKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, nil, err
	}

	theirKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, nil, err
	}

	return &crypto.KeyPair{
			Ours:   ourKey,
			Theirs: theirKey.PublicKey(),
		}, &crypto.KeyPair{
			Ours:   theirKey,
			Theirs: ourKey.PublicKey(),
		}, nil
}

func GenerateSignalingMessage() *pb.SignalingMessage {
	return &pb.SignalingMessage{
		Session: &pb.SessionDescription{
			//#nosec G404 -- This is just test data
			Epoch: rand.Int63(),
		},
	}
}

func ParseIP(s string) (net.IPNet, error) {
	ip, netw, err := net.ParseCIDR(s)
	if err != nil {
		return net.IPNet{}, err
	}

	return net.IPNet{
		IP:   ip,
		Mask: netw.Mask,
	}, nil
}

func Entropy(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}

	var length = float64(len(data))
	var entropy = 0.0

	for i := 0; i < 256; i++ {
		if p := float64(bytes.Count(data, []byte{byte(i)})) / length; p > 0 {
			entropy += -p * math.Log2(p)
		}
	}

	return entropy
}

// TempFileName generates a temporary filename for use in testing or whatever
func TempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}
