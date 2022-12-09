package main

import (
	"encoding/json"
	"io"
	"sync"
)

type CacheManager struct {
	data map[string]string
	sync.RWMutex
}

func NewCacheManager() *CacheManager {
	cm := &CacheManager{
		data: make(map[string]string),
	}
	return cm
}

func (cm *CacheManager) Get(key string) string {
	cm.RLock()
	ret := cm.data[key]
	cm.RUnlock()
	return ret
}

func (cm *CacheManager) Set(key string, value string) error {
	cm.Lock()
	defer cm.Unlock()
	cm.data[key] = value
	return nil
}

// Marshal serializes cache data
func (cm *CacheManager) Marshal() ([]byte, error) {
	cm.RLock()
	defer cm.RUnlock()
	dataBytes, err := json.Marshal(cm.data)
	return dataBytes, err
}

// UnMarshal deserializes cache data
func (cm *CacheManager) UnMarshal(serialized io.ReadCloser) error {
	var newData map[string]string
	if err := json.NewDecoder(serialized).Decode(&newData); err != nil {
		return err
	}

	cm.Lock()
	defer cm.Unlock()
	cm.data = newData

	return nil
}
