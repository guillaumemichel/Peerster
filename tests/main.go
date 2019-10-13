package main

import (
	"fmt"
	"sync"

	u "github.com/guillaumemichel/Peerster/utils"
)

func main() {
	var sm sync.Map
	sm.Store("coucou", 12)
	sm.Store("hello", 1)
	sm.Store("hell22o", 1)
	sm.Store("hellofeef", 1)

	f := func(s, i interface{}) bool {
		fmt.Println(s, i)
		return true
	}

	sm.Range(f)
	fmt.Println(u.SyncMapCount(&sm))
}
