package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("Hello world")
	a, _ := os.OpenFile("../", os.O_RDONLY, os.ModeDir)
	b, _ := a.Readdir(-1)
	fmt.Printf("%p", b[0])

}
