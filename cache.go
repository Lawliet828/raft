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
	cm := &CacheManager{}
	cm.data = make(map[string]string)
	return cm
}

func (c *CacheManager) Get(key string) string {
	c.RLock()
	ret := c.data[key]
	c.RUnlock()
	return ret
}

func (c *CacheManager) Set(key string, value string) error {
	c.Lock()
	defer c.Unlock()
	c.data[key] = value
	return nil
}

// Marshal serializes cache data
func (c *CacheManager) Marshal() ([]byte, error) {
	c.RLock()
	defer c.RUnlock()
	dataBytes, err := json.Marshal(c.data)
	return dataBytes, err
}

// UnMarshal deserializes cache data
func (c *CacheManager) UnMarshal(serialized io.ReadCloser) error {
	var newData map[string]string
	if err := json.NewDecoder(serialized).Decode(&newData); err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()
	c.data = newData

	return nil
}
