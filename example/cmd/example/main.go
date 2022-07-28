package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mokiat/gocrane/example/internal"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./public")))
	mux.Handle("/greet", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, internal.Greet())
	}))
	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}

	log.Println("listening...")
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server error: %v\n", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigChan)
	<-sigChan

	log.Println("shutting down...")
	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("failed to shutdown server: %v\n", err)
	}
	log.Println("good bye")
}
