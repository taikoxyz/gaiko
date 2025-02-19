package transition

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/taikoxyz/gaiko/internal"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/pacaya"
)

type BlockMetadataFork interface {
	ABIEncoder
	Hash() common.Hash
}

type HeklaBlockMetadata struct {
	*ontake.TaikoDataBlockMetadata
}

func (m *HeklaBlockMetadata) Encode() ([]byte, error) {
	return blockMetadataComponentsArgs.Pack(m.TaikoDataBlockMetadata)
}

func (m *HeklaBlockMetadata) Hash() common.Hash {
	b, _ := m.Encode()
	return common.BytesToHash(internal.Keccak(b))
}

type OntakeBlockMetadata struct {
	*ontake.TaikoDataBlockMetadataV2
}

func (m *OntakeBlockMetadata) Encode() ([]byte, error) {
	return blockMetadataV2ComponentsArgs.Pack(m.TaikoDataBlockMetadataV2)
}

func (m *OntakeBlockMetadata) Hash() common.Hash {
	b, _ := m.Encode()
	return common.BytesToHash(internal.Keccak(b))
}

type PacayaBlockMetadata struct {
	*pacaya.ITaikoInboxBatchMetadata
}

func NewPacayaBlockMetadata(meta *pacaya.ITaikoInboxBatchMetadata) *PacayaBlockMetadata {
	return &PacayaBlockMetadata{meta}
}

func (m *PacayaBlockMetadata) Encode() ([]byte, error) {
	return blockMetadataComponentsArgs.Pack(m.ITaikoInboxBatchMetadata)
}

func (m *PacayaBlockMetadata) Hash() common.Hash {
	b, _ := m.Encode()
	return common.BytesToHash(internal.Keccak(b))
}
