package main

import (
	"flag"
	"fmt"
	"github.com/jianzhiyao/sgo"
	"log"
)

var (
	port       int
	backend    string
	waitSecond int
)

func main() {
	flag.IntVar(&port, `p`, 0, `set backend server`)
	flag.StringVar(&backend, `b`, "", `set port to listen`)
	flag.IntVar(&waitSecond, `w`, 3, `set wait seconds to render`)

	flag.Parse()

	if backend == `` {
		log.Fatal("Please set backend server(-b http://google.com)")
		return
	}
	if port == 0 {
		log.Fatal("Please set listen port(-p 8989)")
		return
	}

	if err := sgo.NewDefaultServer(backend, port, waitSecond).ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}
