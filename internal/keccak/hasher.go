package keccak

import (
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
)

// keccakHasher is used to compute the sha256 hash of the provided data.
type keccakHasher struct{ sha crypto.KeccakState }

var keccakHasherPool = sync.Pool{
	New: func() any { return &keccakHasher{sha: crypto.NewKeccakState()} },
}

func newKeccakHasher() *keccakHasher {
	return keccakHasherPool.Get().(*keccakHasher)
}

func (h *keccakHasher) hash(data []byte) []byte {
	b := make([]byte, 32)
	h.sha.Reset()
	h.sha.Write(data)
	h.sha.Read(b)
	return b
}

func (h *keccakHasher) release() {
	keccakHasherPool.Put(h)
}

func Keccak(data []byte) []byte {
	h := newKeccakHasher()
	defer h.release()
	return h.hash(data)
}
