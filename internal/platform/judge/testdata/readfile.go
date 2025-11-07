package main

import (
	"fmt"
	"os"
)

func main() {
	data, err := os.ReadFile("secret_text")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(data)
	}
}
