package intf

import (
	"encoding/base64"

	"riasc.eu/wice/pkg/crypto"
)

// LocalCredentials returns local ufrag, pwd of a peer
// ufrag  is base64 encoded public key of the peer which wants to connect
// pwd    is the base64 encoded and encrypted public key of our interface
//        for encryption the public key of the peer is used
func (p *Peer) LocalCreds() (string, string, error) {
	pl := p.Interface.PublicKey()
	enc, err := crypto.Curve25519Crypt(p.Interface.PrivateKey(), p.PublicKey(), pl[:])
	if err != nil {
		return "", "", err
	}

	return base64.StdEncoding.EncodeToString(p.Peer.PublicKey[:]),
		base64.StdEncoding.EncodeToString(enc), nil
}

// RemoteCredentials returns remote ufrag, pwd of a peer
func (p *Peer) RemoteCredentials() (string, string, error) {
	pl := p.PublicKey()
	enc, err := crypto.Curve25519Crypt(p.Interface.PrivateKey(), p.PublicKey(), pl[:])
	if err != nil {
		return "", "", err
	}

	return p.Interface.PublicKey().String(),
		base64.StdEncoding.EncodeToString(enc), nil
}
