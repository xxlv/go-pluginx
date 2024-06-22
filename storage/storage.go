package storage

import "fmt"

type Kv struct {
	Key   string
	Value string
}

func (kv Kv) String() string {
	return fmt.Sprintf("key=%s,value=%s", kv.Key, kv.Value)
}

type Storage interface {
	Store(kv Kv) error
	Delete(key string) error
	Get(key string) (string, error)
}
