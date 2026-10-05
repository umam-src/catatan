package db

import (
    "context"
    "path/filepath"
    "testing"
)

func TestOpenRunsMigrationAndCreatesDefaultNotebook(t *testing.T) {
    path := filepath.Join(t.TempDir(), "catatan.db")
    d, err := Open(context.Background(), path)
    if err != nil { t.Fatal(err) }
    defer d.Close()

    var count int
    if err := d.QueryRow(`SELECT COUNT(*) FROM notebooks WHERE id='default'`).Scan(&count); err != nil { t.Fatal(err) }
    if count != 1 { t.Fatalf("notebook awal = %d, ingin 1", count) }

    var version int
    if err := d.QueryRow(`SELECT MAX(version) FROM schema_migrations`).Scan(&version); err != nil { t.Fatal(err) }
    if version != 1 { t.Fatalf("versi migrasi = %d, ingin 1", version) }
}
