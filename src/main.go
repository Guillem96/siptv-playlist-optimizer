package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/src/configuration"
	"github.com/Guillem96/optimized-m3u-iptv-list-server/src/server"
	"github.com/Guillem96/optimized-m3u-iptv-list-server/src/siptv"
	"github.com/Guillem96/optimized-m3u-iptv-list-server/src/utils"
)

const DEFAULT_PORT = 9090
const DEFAULT_CONF_PATH = "config.yaml"

var conf = siptv.DigestYAMLConfiguration(loadConfFromEnv())

func main() {
	fmt.Println(conf)
	l := log.New(os.Stdout, "SERVER: ", log.Ldate|log.Ltime)
	h := server.NewHandler(conf, l)

	if utils.IsRunningInLambdaEnv() {
		s := server.NewLambdaServer(h, l)
		s.Run()
	} else {
		port, err := portFromEnv()
		if err != nil {
			l.Fatal(err)
		}
		s := server.NewHttpServer("0.0.0.0", port, h, l)
		s.Run()
	}
}

func portFromEnv() (int, error) {
	ps, present := os.LookupEnv("PORT")
	if !present {
		return DEFAULT_PORT, nil
	}

	port, err := strconv.Atoi(ps)
	if err != nil {
		return -1, fmt.Errorf("error parsing port from $PORT env. %v", err)
	}
	return port, nil
}

func loadConfFromEnv() configuration.Configuration {
	fpath, isPresent := os.LookupEnv("CONFIG")
	if !isPresent {
		fpath = DEFAULT_CONF_PATH
	}
	return configuration.LoadConfiguration(fpath)
}
