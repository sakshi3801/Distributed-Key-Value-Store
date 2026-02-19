package hash

import (
	"hash/crc32"
	"sort"
	"sync"
)

// Ring implements consistent hashing for distributing keys across nodes.
// Supports dynamic node addition and removal with minimal key movement.
type Ring struct {
	mu         sync.RWMutex
	replicas   int
	sortedKeys []uint32
	circle     map[uint32]string
	nodes      map[string]struct{}
}

// NewRing creates a consistent hash ring with the given number of virtual nodes per physical node.
func NewRing(replicas int) *Ring {
	if replicas < 1 {
		replicas = 64
	}
	return &Ring{
		replicas: replicas,
		circle:   make(map[uint32]string),
		nodes:    make(map[string]struct{}),
	}
}

// hashKey returns a uint32 hash for the given key.
func hashKey(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

// Add adds a node to the ring with multiple virtual replicas.
func (r *Ring) Add(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.nodes[node]; ok {
		return
	}
	r.nodes[node] = struct{}{}
	for i := 0; i < r.replicas; i++ {
		h := hashKey(node + "-" + string(rune(i)))
		r.circle[h] = node
		r.sortedKeys = append(r.sortedKeys, h)
	}
	sort.Slice(r.sortedKeys, func(i, j int) bool { return r.sortedKeys[i] < r.sortedKeys[j] })
}

// Remove removes a node and all its virtual replicas from the ring.
func (r *Ring) Remove(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.nodes[node]; !ok {
		return
	}
	delete(r.nodes, node)
	var newKeys []uint32
	for _, k := range r.sortedKeys {
		if r.circle[k] == node {
			delete(r.circle, k)
		} else {
			newKeys = append(newKeys, k)
		}
	}
	r.sortedKeys = newKeys
}

// Get returns the node responsible for the given key.
// Returns empty string if the ring is empty.
func (r *Ring) Get(key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.sortedKeys) == 0 {
		return ""
	}
	h := hashKey(key)
	idx := sort.Search(len(r.sortedKeys), func(i int) bool {
		return r.sortedKeys[i] >= h
	})
	if idx == len(r.sortedKeys) {
		idx = 0
	}
	return r.circle[r.sortedKeys[idx]]
}

// GetN returns the N nodes responsible for the key (for replication).
// Fewer than N may be returned if the ring has fewer nodes.
func (r *Ring) GetN(key string, n int) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.sortedKeys) == 0 || n <= 0 {
		return nil
	}
	h := hashKey(key)
	idx := sort.Search(len(r.sortedKeys), func(i int) bool {
		return r.sortedKeys[i] >= h
	})
	if idx == len(r.sortedKeys) {
		idx = 0
	}
	seen := make(map[string]struct{})
	var result []string
	for i := 0; i < len(r.sortedKeys) && len(result) < n; i++ {
		pos := (idx + i) % len(r.sortedKeys)
		node := r.circle[r.sortedKeys[pos]]
		if _, ok := seen[node]; !ok {
			seen[node] = struct{}{}
			result = append(result, node)
		}
	}
	return result
}

// Nodes returns all physical nodes in the ring.
func (r *Ring) Nodes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.nodes))
	for n := range r.nodes {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Contains returns whether the node is in the ring.
func (r *Ring) Contains(node string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.nodes[node]
	return ok
}
