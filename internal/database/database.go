package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool
}

type DomainRecord struct {
	ID          int64     `json:"id"`
	Domain      string    `json:"domain"`
	Available   bool      `json:"available"`
	Signatures  []string  `json:"signatures"`
	CheckedAt   time.Time `json:"checked_at"`
	Pattern     string    `json:"pattern"`
	Length      int       `json:"length"`
	Suffix      string    `json:"suffix"`
}

type ScanSession struct {
	ID              int64     `json:"id"`
	Pattern         string    `json:"pattern"`
	Length          int       `json:"length"`
	Suffix          string    `json:"suffix"`
	TotalDomains    int       `json:"total_domains"`
	AvailableCount  int       `json:"available_count"`
	RegisteredCount int       `json:"registered_count"`
	StartedAt       time.Time `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	Status          string    `json:"status"` // running, completed, failed
}

func New(connString string) (*Database, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	db := &Database{pool: pool}

	// Initialize database schema
	if err := db.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

func (db *Database) initSchema() error {
	ctx := context.Background()

	// Create scan_sessions table
	_, err := db.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS scan_sessions (
			id SERIAL PRIMARY KEY,
			pattern VARCHAR(10) NOT NULL,
			length INTEGER NOT NULL,
			suffix VARCHAR(50) NOT NULL,
			total_domains INTEGER DEFAULT 0,
			available_count INTEGER DEFAULT 0,
			registered_count INTEGER DEFAULT 0,
			started_at TIMESTAMP NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMP,
			status VARCHAR(20) NOT NULL DEFAULT 'running'
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create scan_sessions table: %w", err)
	}

	// Create domain_records table
	_, err = db.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS domain_records (
			id SERIAL PRIMARY KEY,
			session_id INTEGER REFERENCES scan_sessions(id) ON DELETE CASCADE,
			domain VARCHAR(255) NOT NULL UNIQUE,
			available BOOLEAN NOT NULL,
			signatures TEXT[],
			checked_at TIMESTAMP NOT NULL DEFAULT NOW(),
			pattern VARCHAR(10),
			length INTEGER,
			suffix VARCHAR(50)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create domain_records table: %w", err)
	}

	// Create indexes
	_, err = db.pool.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_domain_records_available ON domain_records(available);
		CREATE INDEX IF NOT EXISTS idx_domain_records_session ON domain_records(session_id);
		CREATE INDEX IF NOT EXISTS idx_domain_records_domain ON domain_records(domain);
		CREATE INDEX IF NOT EXISTS idx_scan_sessions_status ON scan_sessions(status);
	`)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

func (db *Database) CreateScanSession(pattern string, length int, suffix string, totalDomains int) (int64, error) {
	ctx := context.Background()
	var sessionID int64

	err := db.pool.QueryRow(ctx, `
		INSERT INTO scan_sessions (pattern, length, suffix, total_domains, status)
		VALUES ($1, $2, $3, $4, 'running')
		RETURNING id
	`, pattern, length, suffix, totalDomains).Scan(&sessionID)

	if err != nil {
		return 0, fmt.Errorf("failed to create scan session: %w", err)
	}

	return sessionID, nil
}

func (db *Database) SaveDomainResult(sessionID int64, domain string, available bool, signatures []string, pattern string, length int, suffix string) error {
	ctx := context.Background()

	_, err := db.pool.Exec(ctx, `
		INSERT INTO domain_records (session_id, domain, available, signatures, pattern, length, suffix)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (domain) DO UPDATE
		SET available = EXCLUDED.available,
		    signatures = EXCLUDED.signatures,
		    checked_at = NOW(),
		    session_id = EXCLUDED.session_id
	`, sessionID, domain, available, signatures, pattern, length, suffix)

	if err != nil {
		return fmt.Errorf("failed to save domain result: %w", err)
	}

	// Update session counters
	if available {
		_, err = db.pool.Exec(ctx, `
			UPDATE scan_sessions
			SET available_count = available_count + 1
			WHERE id = $1
		`, sessionID)
	} else {
		_, err = db.pool.Exec(ctx, `
			UPDATE scan_sessions
			SET registered_count = registered_count + 1
			WHERE id = $1
		`, sessionID)
	}

	return err
}

func (db *Database) CompleteScanSession(sessionID int64) error {
	ctx := context.Background()

	_, err := db.pool.Exec(ctx, `
		UPDATE scan_sessions
		SET completed_at = NOW(), status = 'completed'
		WHERE id = $1
	`, sessionID)

	return err
}

func (db *Database) GetDomains(available *bool, limit, offset int) ([]DomainRecord, error) {
	ctx := context.Background()

	query := `
		SELECT id, domain, available, signatures, checked_at, pattern, length, suffix
		FROM domain_records
	`

	args := []interface{}{}
	if available != nil {
		query += ` WHERE available = $1`
		args = append(args, *available)
		query += ` ORDER BY checked_at DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	} else {
		query += ` ORDER BY checked_at DESC LIMIT $1 OFFSET $2`
		args = append(args, limit, offset)
	}

	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []DomainRecord
	for rows.Next() {
		var d DomainRecord
		err := rows.Scan(&d.ID, &d.Domain, &d.Available, &d.Signatures, &d.CheckedAt, &d.Pattern, &d.Length, &d.Suffix)
		if err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}

	return domains, nil
}

func (db *Database) GetScanSessions(limit, offset int) ([]ScanSession, error) {
	ctx := context.Background()

	rows, err := db.pool.Query(ctx, `
		SELECT id, pattern, length, suffix, total_domains, available_count,
		       registered_count, started_at, completed_at, status
		FROM scan_sessions
		ORDER BY started_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []ScanSession
	for rows.Next() {
		var s ScanSession
		err := rows.Scan(&s.ID, &s.Pattern, &s.Length, &s.Suffix, &s.TotalDomains,
			&s.AvailableCount, &s.RegisteredCount, &s.StartedAt, &s.CompletedAt, &s.Status)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}

func (db *Database) GetStats() (map[string]interface{}, error) {
	ctx := context.Background()

	var totalDomains, availableDomains, registeredDomains int64

	err := db.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE available = true) as available,
			COUNT(*) FILTER (WHERE available = false) as registered
		FROM domain_records
	`).Scan(&totalDomains, &availableDomains, &registeredDomains)

	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_domains":      totalDomains,
		"available_domains":  availableDomains,
		"registered_domains": registeredDomains,
	}

	return stats, nil
}

func (db *Database) SearchDomains(query string, available *bool, limit, offset int) ([]DomainRecord, error) {
	ctx := context.Background()

	sqlQuery := `
		SELECT id, domain, available, signatures, checked_at, pattern, length, suffix
		FROM domain_records
		WHERE domain LIKE $1
	`

	args := []interface{}{"%" + query + "%"}

	if available != nil {
		sqlQuery += ` AND available = $2`
		args = append(args, *available)
		sqlQuery += ` ORDER BY checked_at DESC LIMIT $3 OFFSET $4`
		args = append(args, limit, offset)
	} else {
		sqlQuery += ` ORDER BY checked_at DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	}

	rows, err := db.pool.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []DomainRecord
	for rows.Next() {
		var d DomainRecord
		err := rows.Scan(&d.ID, &d.Domain, &d.Available, &d.Signatures, &d.CheckedAt, &d.Pattern, &d.Length, &d.Suffix)
		if err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}

	return domains, nil
}

func (db *Database) Close() {
	db.pool.Close()
}

