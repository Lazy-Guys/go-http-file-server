package main

import (
	"fmt"
	"time"

	"github.com/Lazy-Guys/myhttp"
)

var wk *myhttp.Worker

func main() {
	wk = myhttp.GetWorkerInstance("0.0.0.0:8888", "openfaas-fn", "/home/xtt/sf", "/home/xtt/cdx/sf/openfaas/sindemo04")
	go wk.RunServer()
	fmt.Println("Server start!")
	for {
		time.Sleep(time.Second * 5)
	}
}
