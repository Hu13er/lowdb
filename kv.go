package lowdb

import (
	"errors"
	"sync"
	"time"
)

var (
	InvalidRevsion = errors.New("Invalid revision")
)

type KVStore interface {
	Keys() []string
	Get(key string) KeyValueMetadata
	Set(data KeyValueMetadata) error
	Delete(data KeyValueMetadata) error
}

type KeyValueMetadata struct {
	Key       string              `json:"key"`
	Value     []byte              `json:"value"`
	Revision  int                 `json:"revision"`
	CreatedAt time.Time           `json:"created_at"`
	Headers   map[string][]string `json:"headers"`
}

func (kvm KeyValueMetadata) Empty() bool {
	return kvm.Key == ""
}

// Simple key value store
type memoryKVStore struct {
	data map[string]KeyValueMetadata
	mu   sync.RWMutex
}

var _ KVStore = (*memoryKVStore)(nil)

func NewMemoryKVStore() KVStore {
	return &memoryKVStore{
		data: make(map[string]KeyValueMetadata),
	}
}

func (kv *memoryKVStore) Keys() []string {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	result := make([]string, 0, len(kv.data))
	for k := range kv.data {
		result = append(result, k)
	}
	return result
}

func (kv *memoryKVStore) Get(key string) KeyValueMetadata {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	return kv.data[key]
}

func (kv *memoryKVStore) Set(data KeyValueMetadata) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	v := kv.data[data.Key]
	if data.Revision >= 0 && data.Revision != v.Revision {
		return InvalidRevsion
	}
	data.Revision += 1
	data.CreatedAt = time.Now()
	kv.data[data.Key] = data
	return nil
}

func (kv *memoryKVStore) Delete(data KeyValueMetadata) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	v := kv.data[data.Key]
	if data.Revision >= 0 && data.Revision != v.Revision {
		return InvalidRevsion
	}
	delete(kv.data, data.Key)
	return nil
}
