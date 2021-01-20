package test

import (
	"log"
	"regc"
	"testing"
)

func TestMains(t *testing.T) {
	regc.ServiceStart(":1234")

}
func TestWatch(t *testing.T) {
	client := regc.NewKVStoreClient("localhost:1234")
	log.Println(<-client.Exec())
}
