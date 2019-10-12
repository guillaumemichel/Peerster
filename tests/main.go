package main

import "fmt"

func main() {
	var test []string
	test[0] = "coucou"
	test[2] = "ça va?"
	fmt.Println(len(test))
}
