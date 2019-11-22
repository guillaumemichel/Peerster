package main

import (
	"fmt"
	"time"
)

func server1(ch chan string) {
	for {
		time.Sleep(3 * time.Second)
		ch <- "ack"
	}
}
func server2(ch chan string) {
	for {
		time.Sleep(10 * time.Second)
		ch <- "end"
	}

}
func main() {
	output1 := make(chan string)
	output2 := make(chan string)
	go server1(output1)
	go server2(output2)

	for {
		fmt.Println("start again")
		for {
			select {
			case s1 := <-output1:
				fmt.Println(s1)
				continue

			case s2 := <-output2:
				fmt.Println(s2)
				break
			}
		}
	}
}
