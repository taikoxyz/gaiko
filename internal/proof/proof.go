package proof

import (
	"encoding/binary"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type ProofResponse struct {
	Proof           hexutil.Bytes  `json:"proof"`
	Quote           hexutil.Bytes  `json:"quote"`
	PublicKey       hexutil.Bytes  `json:"public_key"`
	InstanceAddress common.Address `json:"instance_address"`
	Input           common.Hash    `json:"input"`
}

type Proof [89]byte

func (p *Proof) Hex() string {
	return hex.EncodeToString(p[:])
}

func NewProof(instanceID uint32, newInstance common.Address, sign []byte) *Proof {
	var proof Proof
	binary.BigEndian.PutUint32(proof[:4], instanceID)
	copy(proof[4:24], newInstance.Bytes())
	copy(proof[24:], sign)
	return &proof
}
