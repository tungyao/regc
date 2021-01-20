package regc

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type KVStoreService struct {
	m      map[string]string
	filter map[string]func(key string)
	mu     sync.Mutex
	host   string
}
type KVStoreClient struct {
	host string
}
type Data struct {
	Uuid   string
	TimeSe int
}

func NewKVStoreService(host string) *KVStoreService {
	return &KVStoreService{
		m:      make(map[string]string),
		filter: make(map[string]func(key string)),
		host:   host,
	}
}

// its client
func NewKVStoreClient(host string) *KVStoreClient {
	return &KVStoreClient{
		host: host,
	}
}
func ServiceStart(host string) {
	var kv = NewKVStoreService(host)
	_ = rpc.RegisterName("KVStoreService", kv)
	listener, err := net.Listen("tcp", kv.host)
	if err != nil {
		log.Fatal("ListenTCP error:", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept error:", err)
		}
		rpc.ServeConn(conn)
	}
}

func (ks *KVStoreClient) Exec() chan bool {
	client, err := rpc.Dial("tcp", ks.host)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	var keyChanged = uuid.New().String()
	go func() {
		for {
			err = client.Call(
				"KVStoreService.Set", [2]string{keyChanged, "abc-value"},
				new(struct{}),
			)
			log.Println("call once")
			<-time.After(time.Second * 5)
		}
	}()
	err = client.Call("KVStoreService.Watch", Data{
		Uuid:   keyChanged,
		TimeSe: 10,
	}, new(struct{}))
	if err != nil {
		log.Fatal(err)
	}
	return make(chan bool)
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
	if oldValue := p.m[key]; oldValue == value {
		for _, fn := range p.filter {
			fn(key)
		}
	}
	p.m[key] = value
	return nil
}
func (p *KVStoreService) Watch(d Data, keyChanged *struct{}) error {
	ch := make(chan string, 10) // buffered
	// 用一个算法来进行 判断那边是否是已经断开了连接
	// 每次change都向后面延长 标注的时间
	var isStop = make(chan bool)
	go func() {
		for {
			log.Println("loop")
			select {
			case key := <-ch:
				log.Println("get data", key)
			case <-time.After(time.Duration(d.TimeSe) * time.Second):
				isStop <- true
			}
		}
	}()
	p.mu.Lock()
	p.filter[d.Uuid] = func(key string) { ch <- key }
	p.mu.Unlock()

	// 用一个协程单独进行计时
	for {
		ok, v := <-isStop
		if ok && v {
			// 已经断开了连接
			log.Println("disconnect ", d.Uuid)
			delete(p.m, d.Uuid)
			break
		}
	}
	return nil
}
