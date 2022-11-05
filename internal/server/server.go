package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/pkg/utils"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/gorilla/mux"
)

type Handler interface {
	LogRequest(r *http.Request)
	CheckBasicAuth(username, password string) bool
	FetchTVM3UPlaylist(w http.ResponseWriter, r *http.Request)
	PlayerApiHandler(w http.ResponseWriter, r *http.Request)
	RedirectToStreamHandler(w http.ResponseWriter, r *http.Request)
}

type Server interface {
	Run()
}

type LambdaServerConfig struct {
	Handler Handler
	Logger  *log.Logger
}

type LambdaServer struct {
	c LambdaServerConfig
	r *mux.Router
	l *log.Logger
}

// NewLambdaServer creates a pointer to a server ready to run within an
// AWS lambda environment
func NewLambdaServer(config LambdaServerConfig) *LambdaServer {
	ls := &LambdaServer{l: config.Logger, c: config}
	ls.r = setupRouter(config.Handler)
	return ls
}

// Run starts the LambdaServer
func (ls *LambdaServer) Run() {
	ls.l.Println("Starting up in Lambda Runtime")
	adapter := gorillamux.NewV2(ls.r)
	lambda.Start(adapter.ProxyWithContext)
}

type HttpServerConfig struct {
	Port    int
	Host    string
	Handler Handler
	Logger  *log.Logger
}

type HttpServer struct {
	c HttpServerConfig
	r *mux.Router
	l *log.Logger
}

// NewHttpServer creates an HTTP server that can run everywhere
func NewHttpServer(config HttpServerConfig, handler Handler) *HttpServer {
	https := &HttpServer{l: config.Logger, c: config}
	https.r = setupRouter(config.Handler)
	return https
}

// Run starts the HTTP server
func (https *HttpServer) Run() {
	server := &http.Server{
		Addr:         fmt.Sprintf("%v:%v", https.c.Host, https.c.Port),
		Handler:      utils.NewCORS(https.r),
		IdleTimeout:  120 * time.Second,
		WriteTimeout: 20 * time.Second,
		ReadTimeout:  20 * time.Second,
	}
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			https.l.Fatal(err)
		}
	}()

	https.l.Printf("Serving at %v...\n", server.Addr)

	// Graceful stop
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChannel
	https.l.Println("Received terminate, graceful shutdown", sig)

	tc, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	cancelFn()
	server.Shutdown(tc)
}
