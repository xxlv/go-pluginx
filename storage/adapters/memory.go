package storages

import (
	"errors"

	"github.com/xxlv/go-pluginx/storage"
)

type MemoryStorage struct {
	Kv map[string]string
}

func (ms *MemoryStorage) Store(kv storage.Kv) error {
	if ms.Kv == nil {
		ms.Kv = make(map[string]string)
	}
	ms.Kv[kv.Key] = kv.Value
	return nil
}

func (ms *MemoryStorage) Delete(key string) error {
	if ms.Kv != nil {
		delete(ms.Kv, key)
	}
	return nil
}

func (ms *MemoryStorage) Get(key string) (string, error) {
	if ms.Kv != nil {
		return ms.Kv[key], nil
	}
	return "", errors.New("not found")
}
