package main

import (
	"fmt"
	"server"
)

func main() {
	fmt.Println("---")
	fe := server.MakeFrontend(9999)
}
