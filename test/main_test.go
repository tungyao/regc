package test

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"regc"
	"testing"
	"time"
)

func TestMains(t *testing.T) {
	var kv = regc.NewKVStoreService()
	_ = rpc.RegisterName("KVStoreService", kv)
	listener, err := net.Listen("tcp", ":1234")
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
func TestWatch(t *testing.T) {
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	go func() {
		var keyChanged string
		err := client.Call("KVStoreService.Watch", 10, &keyChanged)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("watch:", keyChanged)
	}()
	for {
		err = client.Call(
			"KVStoreService.Set", [2]string{"abc", "abc-value"},
			new(struct{}),
		)
		if err != nil {
			log.Println(err)
		}

		time.Sleep(time.Second * 3)
	}

}
