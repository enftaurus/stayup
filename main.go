package main

import (
	"fmt"
)

func print() {
	//for i := 0; i < 10; i++ {
	//	fmt.Println("iteration", i)
	//}
	fmt.Println("Hello World")
}
func main() {
	go print()
	fmt.Println("hello world")

}
