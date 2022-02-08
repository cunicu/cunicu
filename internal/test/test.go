package test

import (
	"math/rand"
	"net"

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
		Type: pb.SignalingMessage_OFFER,
		Description: &pb.SessionDescription{
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
