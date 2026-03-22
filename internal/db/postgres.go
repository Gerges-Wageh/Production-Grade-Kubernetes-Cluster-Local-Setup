package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}

	// ── App-side connection pool ──────────────────────────────────────────────
	//
	// DATABASE_URL points at PgBouncer (:6432), not Postgres directly.
	// Keep this pool small — PgBouncer multiplexes connections to Postgres.
	//
	// Flow: app pool (10 conns) → PgBouncer :6432 → Postgres :5432 (25 real conns)
	//
	// With 3 app replicas: 3 × 10 = 30 connections reach PgBouncer,
	// which funnels them into 25 real Postgres connections regardless
	// of how many pods or k6 VUs are running.

	// Maximum open connections to PgBouncer from this app instance.
	// Rule of thumb: (vCPU × 2) + 1, capped at 10 for Go's goroutine model.
	DB.SetMaxOpenConns(10)

	// Keep 2 connections warm to avoid reconnect latency on burst traffic.
	DB.SetMaxIdleConns(2)

	// Close idle connections after 5 minutes — prevents stale conn buildup
	// in PgBouncer's server pool when app traffic drops.
	DB.SetConnMaxIdleTime(5 * time.Minute)

	// Recycle connections every 30 minutes — prevents issues with
	// PgBouncer's server_lifetime and load balancer TCP timeouts.
	DB.SetConnMaxLifetime(30 * time.Minute)

	// ── IMPORTANT: PgBouncer transaction mode constraints ─────────────────────
	//
	// In transaction mode, a server connection is only held for the duration
	// of a single transaction. Between transactions, you get a different
	// backend connection. This means the following must NOT be used:
	//
	//   ✗  SET search_path = ...       (session state, lost between txns)
	//   ✗  SET LOCAL ...               (same)
	//   ✗  LISTEN / NOTIFY             (requires dedicated connection)
	//   ✗  pg_advisory_lock()          (session-scoped locks)
	//   ✗  PREPARE / EXECUTE           (prepared statements are session-scoped)
	//   ✗  temporary tables            (session-scoped)
	//
	//   ✓  Regular queries with $1 placeholders — fine
	//   ✓  BEGIN / COMMIT transactions  — fine (PgBouncer tracks these)
	//   ✓  COPY                        — fine in transaction mode

	if err := DB.Ping(); err != nil {
		log.Fatal("Cannot ping DB:", err)
	}

	// Auto-create table on startup.
	// In production, replace this with a proper migration tool (goose, atlas).
	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS posts (
		id      SERIAL PRIMARY KEY,
		title   TEXT NOT NULL,
		content TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

	fmt.Println("Connected to DB — pool: max_open=10 max_idle=2")
}

// ── Health check ──────────────────────────────────────────────────────────────
// Call this from your /healthz handler so Kubernetes knows the app
// is actually able to reach the database, not just that it started.

func Ping() error {
	return DB.Ping()
}

// ── Pool stats — log or expose via /metrics ───────────────────────────────────
// Useful for tuning SetMaxOpenConns and spotting pool exhaustion.

func PoolStats() sql.DBStats {
	return DB.Stats()
}

// ── Models & queries ──────────────────────────────────────────────────────────

type Post struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func InsertPost(post *Post) error {
	return DB.QueryRow(
		"INSERT INTO posts (title, content) VALUES ($1, $2) RETURNING id",
		post.Title, post.Content,
	).Scan(&post.ID)
}

func GetPostByID(id int) (*Post, error) {
	post := &Post{}
	err := DB.QueryRow(
		"SELECT id, title, content FROM posts WHERE id=$1",
		id,
	).Scan(&post.ID, &post.Title, &post.Content)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return post, err
}
