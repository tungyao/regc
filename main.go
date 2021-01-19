package regc

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

type KVStoreService struct {
	m      map[string]string
	filter map[string]func(key string)
	mu     sync.Mutex
}

func NewKVStoreService() *KVStoreService {
	return &KVStoreService{
		m:      make(map[string]string),
		filter: make(map[string]func(key string)),
	}
}
func (p *KVStoreService) Get(key string, value *string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if v, ok := p.m[key]; ok {
		*value = v
		return nil
	}

	return fmt.Errorf("not found")
}
func (p *KVStoreService) Set(kv [2]string, reply *struct{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	key, value := kv[0], kv[1]
	if oldValue := p.m[key]; oldValue != value {
		for _, fn := range p.filter {
			fn(key)
		}
	}
	p.m[key] = value
	return nil
}
func (p *KVStoreService) Watch(timeoutSecond int, keyChanged *string) error {
	id := uuid.New().String()
	ch := make(chan string, 10) // buffered
	// 用一个算法来进行 判断那边是否是已经断开了连接
	// 每次change都向后面延长 标注的时间
	var isStop = make(chan bool)
	go func() {
		for {
			select {
			case key := <-ch:
				*keyChanged = key
			case <-time.After(time.Duration(timeoutSecond) * time.Second):
				isStop <- true
			}
		}
	}()
	p.mu.Lock()
	p.filter[id] = func(key string) { ch <- key }
	p.mu.Unlock()

	// 用一个协程单独进行计时
	for {
		ok, v := <-isStop
		if ok && v {
			// 已经断开了连接
			break
		}
	}
	return nil
}
