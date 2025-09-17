package main

import (
	"context"
	"fmt"
	"itk/internal/utils"
	"itk/internal/wallet"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("itk start!")

	db, err := utils.NewPostgresPool(context.Background())
	if err != nil {
		log.Fatalln("DB connection failed:", err)
	}
	defer db.Close()
	log.Println("database connected")

	repo := wallet.NewDBRepo(db)
	svc := wallet.NewWalletService(repo)
	h := wallet.NewWalletHandler(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("listening %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-stop
	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
	log.Println("server exited gracefully")
}
