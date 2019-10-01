package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	UIPort := flag.String("UIPort", "8080", "port for the UI client")
	msg := flag.String("msg", "", "message to be sent")

	flag.Parse()
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	fmt.Println(*UIPort + " " + *msg)

}
