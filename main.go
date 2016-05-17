package main

import (
	"fmt"
	"github.com/jlgrady1/moby/domain"
	"github.com/jlgrady1/moby/infrastructure"
)

func main() {
	fmt.Println("test")
	fmt.Println(domain.STOPPED)
	log := infrastructure.Logger()
	log.log("hello world")
}
