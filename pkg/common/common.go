package common

import (
	"sync"
)

type ConcMap struct {
	l     sync.RWMutex
	KVMap map[interface{}]interface{}
}

func NewConcMap() *ConcMap {
	return &ConcMap{
		KVMap: make(map[interface{}]interface{}),
	}
}

func (c *ConcMap) Add(key interface{}, value interface{}) {
	c.l.Lock()
	defer c.l.Unlock()
	c.KVMap[key] = value
}

func (c *ConcMap) GetKeyValMap() map[interface{}]interface{} {
	return c.KVMap
}

type ConcStrToStrMap struct {
	l     sync.RWMutex
	KVMap map[string]string
}

func NewConcStrToStrMap() *ConcStrToStrMap {
	return &ConcStrToStrMap{
		KVMap: make(map[string]string),
	}
}

func (c *ConcStrToStrMap) Add(key string, value string) {
	c.l.Lock()
	defer c.l.Unlock()
	c.KVMap[key] = value
}

func (c *ConcStrToStrMap) GetKeyValMap() map[string]string {
	return c.KVMap
}

type ConcStringErrChanMap struct {
	l             sync.RWMutex
	StrErrChanMap map[string]chan (error)
}

func NewConcStringErrChanMap() *ConcStringErrChanMap {
	return &ConcStringErrChanMap{
		StrErrChanMap: make(map[string]chan (error)),
	}
}

func (c *ConcStringErrChanMap) Add(key string, value chan (error)) {
	c.l.Lock()
	defer c.l.Unlock()
	c.StrErrChanMap[key] = value
}

func (c *ConcStringErrChanMap) GetKeyValMap() map[string]chan (error) {
	c.l.Lock()
	defer c.l.Unlock()
	return c.StrErrChanMap
}
