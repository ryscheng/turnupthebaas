package server

import (
	"fmt"
)

type Frontend struct {
	temp int
}

func MakeFrontend(port int) *FrontEnd {
	fe := new(Frontend)
	fe.temp = 100
	fmt.Println(port)
	fmt.Println(fe.temp)
	return fe
}
