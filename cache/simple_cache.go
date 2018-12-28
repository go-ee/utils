package cache

import (
	"sync"
	"time"
)

type SimpleCache struct {
	MaxSize int
	data    map[string]interface{}
	lock    sync.Mutex
}

func NewCache() *SimpleCache {
	return &SimpleCache{MaxSize: 1000, data: make(map[string]interface{}), lock: sync.Mutex{}}
}

func (o SimpleCache) Clear() {
	o.lock.Lock()
	o.data = make(map[string]interface{})
	o.lock.Unlock()
	return
}

func (o SimpleCache) Get(key string, builder func() interface{}) (value interface{}, ok bool) {
	o.lock.Lock()
	value, ok = o.data[key]
	o.lock.Unlock()
	return
}

func (o SimpleCache) GetOrBuild(key string, builder func() (interface{}, error)) (value interface{}, err error) {
	o.lock.Lock()
	value, ok := o.data[key]
	if !ok {
		value, err = builder()
		if err == nil {
			o.put(key, value)
		}
	}
	o.lock.Unlock()
	return
}

func (o SimpleCache) Put(key string, value interface{}) {
	o.lock.Lock()
	o.put(key, value)
	o.lock.Unlock()
}

func (o SimpleCache) put(key string, value interface{}) {
	//reset cache
	if len(o.data) >= o.MaxSize {
		o.data = make(map[string]interface{})
	}
	o.data[key] = value
}

type TimeoutObjectCache struct {
	Timeout time.Duration
	Builder func() (interface{}, error)

	buildAt time.Time
	obj     interface{}

	lock sync.Mutex
}

func (o *TimeoutObjectCache) Get() (ret interface{}, err error) {
	o.lock.Lock()
	if o.obj == nil || time.Since(o.buildAt) > o.Timeout {
		o.obj, err = o.Builder()
		o.buildAt = time.Now()
		if err != nil {
			o.obj = nil
		}
	}
	ret = o.obj
	o.lock.Unlock()
	return
}

func NewObjectCache(builder func() (interface{}, error)) *TimeoutObjectCache {
	return &TimeoutObjectCache{Timeout: 3 * time.Second, Builder: builder}
}
