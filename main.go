package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const (
	agentName    = "sirus-print-agent"
	agentVersion = "1.0.0"
)

func main() {
	// Lokasi config relatif terhadap binary (penting untuk Windows Service —
	// working dir bisa berbeda dengan letak .exe).
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("cannot get exe path: %v", err)
	}
	configPath := filepath.Join(filepath.Dir(exePath), "config.json")

	// Saat dev, fallback ke ./config.json
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "config.json"
	}

	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("load config gagal: %v", err)
	}
	log.Printf("Config loaded from %s — %d printer(s) terdaftar", configPath, len(config.Printers))

	// Routing
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex(config))
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/printers", handlePrinters(config))
	mux.HandleFunc("/system-printers", handleSystemPrinters)
	mux.HandleFunc("/cetak-pdf", handleCetakPDF(config))

	handler := logMiddleware(corsMiddleware(config.AllowedOrigins, mux))

	addr := fmt.Sprintf("127.0.0.1:%d", config.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second, // print bisa makan waktu
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("%s v%s listening at http://%s", agentName, agentVersion, addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// Graceful shutdown via signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("Bye.")
}
