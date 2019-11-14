package main

import (
	"fmt"
	"time"
)

func server1(ch chan string) {
	time.Sleep(6 * time.Second)
	ch <- "from server1"
}
func server2(ch chan string) {
	time.Sleep(3 * time.Second)
	ch <- "from server2"

}
func main() {
	output1 := make(chan string)
	for {
		fmt.Println("start")
		go server1(output1)
		select {
		case s1 := <-output1:
			fmt.Println(s1)
		}
	}
}
