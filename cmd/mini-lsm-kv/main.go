package main

import (
	"fmt"
	"log"

	"github.com/dashjay/mini-lsm-go/pkg/lsm"
)

func main() {
	lsmKV := lsm.NewStorage("/workspaces/mini-lsm-go/test")

	const count = 1000
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := fmt.Sprintf("value-%d", i)
		log.Printf("key: %s, value: %s", key, value)
		lsmKV.Put([]byte(key), []byte(value))
	}
	err := lsmKV.Sync()
	if err != nil {
		panic(err)
	}
	scanner := lsmKV.Scan([]byte("key-500"), []byte("key-509"))
	for i := 500; i < 510; i++ {
		log.Printf("%s should equal %s, %s should equal %s", scanner.Key(), fmt.Sprintf("key-%d", i), scanner.Value(), fmt.Sprintf("value-%d", i))
		scanner.Next()
	}
	for i := 0; i < count; i++ {
		lsmKV.Put([]byte(fmt.Sprintf("keyn-%d", i)), []byte(fmt.Sprintf("valuen-%d", i)))
	}
	err = lsmKV.Sync()
	if err != nil {
		panic(err)
	}
	log.Printf("%t should be %t", scanner.IsValid(), false)

	// for i := 0; i < count/2; i++ {
	// 	lsmKV.Delete([]byte(fmt.Sprintf("key-%d", i)))
	// }
	// err = lsmKV.Sync()
	// if err != nil {
	// 	panic(err)
	// }
	// scanner = lsmKV.Scan(key, append(key, 'n'))
	// for scanner.IsValid() {
	// 	fmt.Printf("%s: %s\n", scanner.Key(), scanner.Value())
	// 	scanner.Next()
	// }
	// lsmKV.Compact()
	// scanner = lsmKV.Scan(key, append(key, 'n'))
	// for scanner.IsValid() {
	// 	fmt.Printf("%s: %s\n", scanner.Key(), scanner.Value())
	// 	scanner.Next()
	// }
	// time.Sleep(1 * time.Second)
}
