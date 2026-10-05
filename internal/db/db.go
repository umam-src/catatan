package db

import (
    "context"
    "database/sql"
    "embed"
    "fmt"
    "sort"

    _ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// DB membungkus koneksi basis data lokal dan migrasinya.
type DB struct { *sql.DB }

func Open(ctx context.Context, path string) (*DB, error) {
    dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)", path)
    raw, err := sql.Open("sqlite", dsn)
    if err != nil { return nil, err }
    raw.SetMaxOpenConns(1)
    if err := raw.PingContext(ctx); err != nil { raw.Close(); return nil, err }
    d := &DB{DB: raw}
    if err := d.migrate(ctx); err != nil { raw.Close(); return nil, err }
    return d, nil
}

func (d *DB) migrate(ctx context.Context) error {
    if _, err := d.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, name TEXT NOT NULL)`); err != nil { return err }
    entries, err := migrationFiles.ReadDir("migrations")
    if err != nil { return err }
    sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
    for _, entry := range entries {
        var version int
        if _, err := fmt.Sscanf(entry.Name(), "%d_", &version); err != nil { continue }
        var applied int
        if err := d.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version).Scan(&applied); err != nil { return err }
        if applied != 0 { continue }
        sqlText, err := migrationFiles.ReadFile("migrations/" + entry.Name())
        if err != nil { return err }
        tx, err := d.BeginTx(ctx, nil)
        if err != nil { return err }
        if _, err = tx.ExecContext(ctx, string(sqlText)); err != nil { tx.Rollback(); return fmt.Errorf("migrasi %s: %w", entry.Name(), err) }
        if _, err = tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, name) VALUES (?, ?)`, version, entry.Name()); err != nil { tx.Rollback(); return err }
        if err = tx.Commit(); err != nil { return err }
    }
    return nil
}

func (d *DB) Close() error { return d.DB.Close() }
