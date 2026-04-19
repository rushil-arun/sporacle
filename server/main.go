package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	gameinit "server/game-init"
	rediscoord "server/redis"
	state "server/state"
	trivia "server/trivia"
)

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Echo requested headers when present to satisfy preflight requests.
		if requested := r.Header.Get("Access-Control-Request-Headers"); requested != "" {
			w.Header().Set("Access-Control-Allow-Headers", requested)
		} else {
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		// Ensure caches don't mix responses across different preflight requests.
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		w.Header().Add("Vary", "Access-Control-Request-Headers")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Println("Welcome to Sporcle!")

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	listen := os.Getenv("SERVER_BASE_URL")
	serverAddr := os.Getenv("SERVER_ADDR")
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb, err := rediscoord.NewClient(redisAddr)
	if err != nil {
		log.Fatalf("redis connect: %v", err)
	}

	ctx := context.Background()
	if err := rediscoord.RegisterServer(ctx, rdb, serverAddr); err != nil {
		log.Fatalf("register server: %v", err)
	}
	log.Printf("Registered as %s", serverAddr)

	globalState := state.NewGlobalState()
	mux := http.NewServeMux()
	gameinit.RegisterRoutes(mux, globalState)
	trivia.RegisterRoutes(mux)

	srv := &http.Server{Addr: listen, Handler: cors(mux)}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	log.Printf("Listening on %s", listen)

	<-sigCh
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)

	if err := rediscoord.DeregisterServer(ctx, rdb, serverAddr); err != nil {
		log.Printf("deregister: %v", err)
	}
	rdb.Close()
	log.Println("Goodbye.")
}
