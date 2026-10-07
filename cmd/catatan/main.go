package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/umam-src/catatan/internal/ai"
	"github.com/umam-src/catatan/internal/db"
	"github.com/umam-src/catatan/internal/version"
	apphttp "github.com/umam-src/catatan/internal/http"
)

func main() {
	versi := flag.Bool("version", false, "tampilkan versi aplikasi")
	dataDir := flag.String("data", "", "direktori data lokal")
	addr := flag.String("alamat", "127.0.0.1:8787", "alamat layanan lokal")
	flag.Parse()
	if *versi {
		fmt.Printf("catatan %s\n", version.Value)
		return
	}

	dir := *dataDir
	if dir == "" {
		var err error
		dir, err = os.UserConfigDir()
		if err != nil {
			log.Fatal(err)
		}
		dir = filepath.Join(dir, "catatan")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.Open(ctx, filepath.Join(dir, "catatan.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	configPath := filepath.Join(dir, "config.json")
	aiConfig, aiModel, aiConfigured, err := ai.ConfigFromFile(configPath)
	if err != nil {
		log.Fatal(err)
	}
	if envURL := os.Getenv("CATATAN_AI_URL"); envURL != "" {
		envConfig, envConfigured, envErr := ai.ConfigFromEnv()
		if envErr != nil {
			log.Fatal(envErr)
		}
		if envConfigured {
			aiConfig = envConfig
			aiConfigured = true
		}
	}
	if envKey := os.Getenv("CATATAN_AI_API_KEY"); envKey != "" && aiConfig != nil {
		aiConfig.APIKey = envKey
	}
	if envModel := ai.ModelFromEnv(); envModel != "" {
		aiModel = envModel
	}
	var aiProvider *ai.Client
	if aiConfigured {
		aiProvider, err = ai.NewClient(*aiConfig)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("penyedia model lokal aktif: %s", aiConfig.BaseURL)
	}

	server := &http.Server{Addr: *addr, Handler: apphttp.NewWithAIModel(database, aiProvider, aiModel)}
	go func() {
		log.Printf("Buku Catatan: http://%s", *addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	_ = server.Shutdown(context.Background())
	fmt.Println("Buku Catatan ditutup.")
}
