package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/internal/configuration"
	"github.com/Guillem96/optimized-m3u-iptv-list-server/internal/server"
	"github.com/Guillem96/optimized-m3u-iptv-list-server/internal/siptv"
	"github.com/Guillem96/optimized-m3u-iptv-list-server/pkg/utils"
)

const DEFAULT_PORT = 9090
const DEFAULT_CONF_PATH = "config.yaml"

const DEFAULT_USER = "user"
const DEFAULT_PASSWORD = "pass"

var conf = siptv.DigestYAMLConfiguration(loadConfFromEnv())

func main() {
	l := log.New(os.Stdout, "SERVER: ", log.Ldate|log.Ltime)
	a := siptv.UserCredentials{
		Username: getEnvOrDefault("USERNAME", DEFAULT_USER),
		Password: getEnvOrDefault("PASSWORD", DEFAULT_PASSWORD),
	}
	h := siptv.NewBasicHTTPHandler(conf, &a, l)

	if utils.IsRunningInLambdaEnv() {
		s := server.NewLambdaServer(server.LambdaServerConfig{Handler: h, Logger: l})
		s.Run()
	} else {
		port, err := portFromEnv()
		if err != nil {
			l.Fatal(err)
		}
		sc := server.HttpServerConfig{Host: "0.0.0.0", Port: port, Handler: h, Logger: l}
		s := server.NewHttpServer(sc, h)
		s.Run()
	}
}

func portFromEnv() (int, error) {
	ps := getEnvOrDefault("PORT", fmt.Sprintf("%d", DEFAULT_PORT))
	port, err := strconv.Atoi(ps)
	if err != nil {
		return -1, fmt.Errorf("error parsing port from $PORT env. %v", err)
	}
	return port, nil
}

func loadConfFromEnv() configuration.Configuration {
	return configuration.LoadConfiguration(getEnvOrDefault("CONFIG", DEFAULT_CONF_PATH))
}

func getEnvOrDefault(key, defaultValue string) string {
	ev, isPresent := os.LookupEnv(key)
	if !isPresent {
		return defaultValue
	}
	return ev
}
