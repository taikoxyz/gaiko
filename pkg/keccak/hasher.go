// Package keccak provides utilities for computing Keccak256 hashes
package keccak

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
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
func (h *keccakHasher) hash(data []byte) common.Hash {
	return crypto.HashData(h.KeccakState, data)
}

// release returns the hasher to the pool for reuse
func (h *keccakHasher) release() {
	keccakHasherPool.Put(h)
}

// Keccak computes the keccak256 hash of the input data
// using a pooled hasher for better performance.
func Keccak(data []byte) common.Hash {
	h := newKeccakHasher()
	defer h.release()
	return h.hash(data)
}
