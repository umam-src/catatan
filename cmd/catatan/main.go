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
		fmt.Printf("catatan %s\\n", version.Value)
		return
	}

	dir := *dataDir
	if dir == "" {
		var err error
		dir, err = os.UserConfigDir()
		if err != nil {
			log.Fatal(err)
		}
		dir = filepath.Join(dir, "BukuCatatan")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.Open(ctx, filepath.Join(dir, "buku-catatan.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	server := &http.Server{Addr: *addr, Handler: apphttp.New(database)}
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
