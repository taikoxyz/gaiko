package prover

import (
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSign(t *testing.T) {
	proof := "0x01000000c13bd882edb37ffbabc9f9e34a0d9789633b850fe55e625b768cc8e5feed7d9f7ab536cbc210c2fcc1385aaf88d8a91d8adc2740245f9deee5fd3d61dd2a71662fb6639515f1e2f3354361a82d86c1952352c1a81b"
	msg := "0x216ac5cd5a5e13b0c9a81efb1ad04526b9f4ddd2fe6ebc02819c5097dfb0958c"
	proofBytes := hexutil.MustDecode(proof)
	msgBytes := hexutil.MustDecode(msg)
	pubKey, err := SigToPub(msgBytes, proofBytes[24:])
	require.NoError(t, err)

	proofAddr := crypto.PubkeyToAddress(*pubKey)
	privKeyStr := "324b5d1744ec27d6ac458350ce6a6248680bb0209521b2c730c1fe82a433eb54"
	privKey, err := crypto.HexToECDSA(privKeyStr)
	require.NoError(t, err)
	pubAddr := crypto.PubkeyToAddress(privKey.PublicKey)
	assert.Equal(t, pubAddr, proofAddr)
}
