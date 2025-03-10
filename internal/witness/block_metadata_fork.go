package witness

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/taikoxyz/gaiko/pkg/keccak"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/ontake"
	"github.com/taikoxyz/taiko-mono/packages/taiko-client/bindings/pacaya"
)

type BlockMetadataFork interface {
	ABIEncoder
	Hash() common.Hash
}

type NothingBlockMetadata struct{}

func (m *NothingBlockMetadata) ABIEncode() ([]byte, error) {
	return nil, nil
}

func (m *NothingBlockMetadata) Hash() common.Hash {
	return keccak.Keccak(nil)
}

type HeklaBlockMetadata struct {
	*ontake.TaikoDataBlockMetadata
}

func NewHeklaBlockMetadata(meta *ontake.TaikoDataBlockMetadata) *HeklaBlockMetadata {
	return &HeklaBlockMetadata{meta}
}

func (m *HeklaBlockMetadata) ABIEncode() ([]byte, error) {
	return blockMetadataComponentsArgs.Pack(m.TaikoDataBlockMetadata)
}

func (m *HeklaBlockMetadata) Hash() common.Hash {
	b, _ := m.ABIEncode()
	return keccak.Keccak(b)
}

type OntakeBlockMetadata struct {
	*ontake.TaikoDataBlockMetadataV2
}

func NewOntakeBlockMetadata(meta *ontake.TaikoDataBlockMetadataV2) *OntakeBlockMetadata {
	return &OntakeBlockMetadata{meta}
}

func (m *OntakeBlockMetadata) ABIEncode() ([]byte, error) {
	return blockMetadataV2ComponentsArgs.Pack(m.TaikoDataBlockMetadataV2)
}

func (m *OntakeBlockMetadata) Hash() common.Hash {
	b, _ := m.ABIEncode()
	return keccak.Keccak(b)
}

type PacayaBlockMetadata struct {
	*pacaya.ITaikoInboxBatchMetadata
}

func NewPacayaBlockMetadata(meta *pacaya.ITaikoInboxBatchMetadata) *PacayaBlockMetadata {
	return &PacayaBlockMetadata{meta}
}

func (m *PacayaBlockMetadata) ABIEncode() ([]byte, error) {
	return blockMetadataComponentsArgs.Pack(m.ITaikoInboxBatchMetadata)
}

func (m *PacayaBlockMetadata) Hash() common.Hash {
	b, _ := m.ABIEncode()
	return keccak.Keccak(b)
}
