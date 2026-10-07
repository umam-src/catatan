package db

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSearchMeasurement(t *testing.T) {
	if os.Getenv("CATATAN_BENCHMARK") != "1" {
		t.Skip("pengukuran hanya dijalankan saat CATATAN_BENCHMARK=1")
	}

	for _, jumlah := range []int{1000, 10000} {
		t.Run(fmt.Sprintf("%d", jumlah), func(t *testing.T) {
			d, path := bukaDBUji(t)

			tx, err := d.Begin()
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < jumlah; i++ {
				id := fmt.Sprintf("ukur-%d", i)
				_, err = tx.Exec(
					`INSERT INTO notes(id, notebook_id, title, content, created_at, updated_at)
					  VALUES (?, "default", ?, ?, datetime("now"), datetime("now"))`,
					id, "Catatan pengukuran lokal", "Isi pengukuran dengan kata kunci pencarian offline",
				)
				if err != nil {
					_ = tx.Rollback()
					t.Fatal(err)
				}
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}

			if _, err := d.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}

			const pengulangan = 100
			for i := 0; i < 10; i++ {
				if _, err := d.Search(context.Background(), "local", "", "pengukuran pencarian", 20); err != nil {
					t.Fatal(err)
				}
			}
			mulai := time.Now()
			for i := 0; i < pengulangan; i++ {
				if _, err := d.Search(context.Background(), "local", "", "pengukuran pencarian", 20); err != nil {
					t.Fatal(err)
				}
			}
			durasi := time.Since(mulai)
			rata := durasi / pengulangan

			t.Logf("jumlah=%d ukuran_db=%d byte rata_rata_pencarian=%s", jumlah, info.Size(), rata)
		})
	}
}
