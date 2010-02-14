package main

import (
	"fmt"
	"teratogen"
)

func main() {
	context := teratogen.NewContext()
	context.InitGame()
	fmt.Println("Command-line teratogen.")
	fmt.Printf("%#v\n", context)
}
