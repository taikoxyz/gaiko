package t8n

import (
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
)

// hasher is used to compute the sha256 hash of the provided data.
type hasher struct{ sha crypto.KeccakState }

var hasherPool = sync.Pool{
	New: func() interface{} { return &hasher{sha: crypto.NewKeccakState()} },
}

func newHasher() *hasher {
	return hasherPool.Get().(*hasher)
}

func (h *hasher) hash(data []byte) []byte {
	b := make([]byte, 32)
	h.sha.Reset()
	h.sha.Write(data)
	h.sha.Read(b)
	return b
}

func (h *hasher) release() {
	hasherPool.Put(h)
}

func keccak(data []byte) []byte {
	h := newHasher()
	defer h.release()
	return h.hash(data)
}
