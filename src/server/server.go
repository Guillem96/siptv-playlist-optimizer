package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/src/utils"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/gorilla/mux"
)

type Server interface {
	Run()
}

type UserCredentials struct {
	Username string
	Password string
}

type LambdaServerConfig struct {
	Auth   *UserCredentials
	Logger *log.Logger
}

type LambdaServer struct {
	c LambdaServerConfig
	r *mux.Router
	l *log.Logger
}

func NewLambdaServer(config LambdaServerConfig, handler *Handler) *LambdaServer {
	ls := &LambdaServer{l: config.Logger, c: config}
	ls.r = setupRouter(handler, config.Auth)
	return ls
}

func (ls *LambdaServer) Run() {
	ls.l.Println("Starting up in Lambda Runtime")
	adapter := gorillamux.NewV2(ls.r)
	lambda.Start(adapter.ProxyWithContext)
}

type HttpServerConfig struct {
	Port   int
	Host   string
	Auth   *UserCredentials
	Logger *log.Logger
}

type HttpServer struct {
	c HttpServerConfig
	r *mux.Router
	l *log.Logger
}

func NewHttpServer(config HttpServerConfig, handler *Handler) *HttpServer {
	https := &HttpServer{l: config.Logger, c: config}
	https.r = setupRouter(handler, config.Auth)
	return https
}

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
	sigChannel := make(chan os.Signal)
	signal.Notify(sigChannel, os.Interrupt)
	signal.Notify(sigChannel, os.Kill)

	sig := <-sigChannel
	https.l.Println("Received terminate, graceful shutdown", sig)

	tc, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	cancelFn()
	server.Shutdown(tc)
}

func setupRouter(h *Handler, a *UserCredentials) *mux.Router {
	r := mux.NewRouter()
	r.Use(h.loggingMiddleware)
	if a != nil {
		h.l.Println("Setting up basic authentication.")
		r.Use(h.basicAuthMiddleware(a.Username, a.Password, "SIPTV-Optim"))
	}

	r.Path("/{tv}").HandlerFunc(h.fetchTVM3UPlaylist).Methods("GET")
	r.NotFoundHandler = http.HandlerFunc(h.notFoundHandler)
	return r
}
