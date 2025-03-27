package prover

import (
	"crypto/ecdsa"
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
)

func Sign(digestHash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	sig, err = crypto.Sign(digestHash, prv)
	if err != nil {
		return nil, err
	}
	recid := 27
	if sig[64] != 0 {
		recid = 28
	}
	sig[64] = byte(recid)
	return sig, nil
}

func SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	if len(sig) != 65 {
		return nil, errors.New("signature's format not correct")
	}
	if sig[64] >= 27 {
		sig[64] -= 27
	}
	return crypto.SigToPub(hash, sig)
}
