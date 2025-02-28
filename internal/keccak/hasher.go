package keccak

import (
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
)

// keccakHasher is used to compute the keccak256 hash of the provided data.
type keccakHasher struct{ crypto.KeccakState }

var keccakHasherPool = sync.Pool{
	New: func() any { return &keccakHasher{crypto.NewKeccakState()} },
}

// newKeccakHasher retrieves a keccakHasher from the pool or creates a new one
func newKeccakHasher() *keccakHasher {
	return keccakHasherPool.Get().(*keccakHasher)
}

// hash computes the keccak256 hash of the input data
// Note: This could return errors from Write/Read operations but follows
// the go-ethereum pattern where these operations are assumed to succeed
func (h *keccakHasher) hash(data []byte) []byte {
	b := make([]byte, 32)
	h.Reset()
	h.Write(data)
	h.Read(b)
	return b
}

// release returns the hasher to the pool for reuse
func (h *keccakHasher) release() {
	keccakHasherPool.Put(h)
}

// Keccak computes the keccak256 hash of the input data
// using a pooled hasher for better performance.
func Keccak(data []byte) []byte {
	h := newKeccakHasher()
	defer h.release()
	return h.hash(data)
}
