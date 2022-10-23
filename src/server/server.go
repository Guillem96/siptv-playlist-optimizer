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

type LambdaServer struct {
	r *mux.Router
	l *log.Logger
}

func NewLambdaServer(handler *Handler, logger *log.Logger) *LambdaServer {
	ls := &LambdaServer{l: logger}
	ls.r = setupRouter(handler)
	return ls
}

func (ls *LambdaServer) Run() {
	ls.l.Println("Starting up in Lambda Runtime")
	adapter := gorillamux.NewV2(ls.r)
	lambda.Start(adapter.ProxyWithContext)
}

type HttpServer struct {
	Port int
	Host string
	r    *mux.Router
	l    *log.Logger
}

func NewHttpServer(host string, port int, handler *Handler, logger *log.Logger) *HttpServer {
	https := &HttpServer{l: logger, Host: host, Port: port}
	https.r = setupRouter(handler)
	return https
}

func (https *HttpServer) Run() {
	server := &http.Server{
		Addr:         fmt.Sprintf("%v:%v", https.Host, https.Port),
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

func setupRouter(h *Handler) *mux.Router {
	r := mux.NewRouter()
	r.Use(h.loggingMiddleware)
	r.Path("/{tv}").HandlerFunc(h.fetchTVM3UPlaylist).Methods("GET")
	r.NotFoundHandler = http.HandlerFunc(h.notFoundHandler)
	return r
}
