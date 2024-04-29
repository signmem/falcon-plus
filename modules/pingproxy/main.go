package main

import (
	"flag"
	"fmt"
	"os"
	"github.com/signmem/falcon-plus/modules/pingproxy/g"
	"github.com/signmem/falcon-plus/modules/pingproxy/http"
)


func main() {

	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")

	flag.Parse()

	if *version {
		version := g.Version
		fmt.Printf("%s", version)
		os.Exit(0)
	}

	g.ParseConfig(*cfg)
	g.Logger = g.InitLog()

	http.Start()

	select {}

}