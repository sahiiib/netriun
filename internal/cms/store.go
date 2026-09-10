package cms

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Store struct{ DB *pgxpool.Pool }

func Open(ctx context.Context, dsn string) (*Store, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	config.MaxConns = 10
	config.ConnConfig.ConnectTimeout = 5 * time.Second
	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	s := &Store{db}
	tx, err := db.Begin(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(2065)`); err != nil {
		tx.Rollback(context.Background())
		db.Close()
		return nil, err
	}
	_, err = tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS site_content (id integer PRIMARY KEY CHECK (id=1), revision bigint NOT NULL DEFAULT 1, body jsonb NOT NULL);
 CREATE TABLE IF NOT EXISTS admin_sessions (token text PRIMARY KEY, csrf text NOT NULL, expires timestamptz NOT NULL);
 CREATE TABLE IF NOT EXISTS request_limits (key text PRIMARY KEY, hits integer NOT NULL, resets timestamptz NOT NULL);
 CREATE TABLE IF NOT EXISTS contact_messages (id bigserial PRIMARY KEY, created timestamptz NOT NULL DEFAULT now(), name text NOT NULL, email text NOT NULL, company text NOT NULL, topic text NOT NULL, message text NOT NULL, recipient text NOT NULL, status text NOT NULL DEFAULT 'received');`)
	if err == nil {
		_, err = tx.Exec(ctx, `INSERT INTO site_content(id,body) VALUES(1,$1) ON CONFLICT DO NOTHING`, seed)
	}
	if err == nil {
		err = tx.Commit(ctx)
	}
	if err != nil {
		tx.Rollback(context.Background())
		db.Close()
		return nil, err
	}
	return s, nil
}
func (s *Store) Load(ctx context.Context) (Content, int64, error) {
	var raw []byte
	var rev int64
	err := s.DB.QueryRow(ctx, `SELECT body,revision FROM site_content WHERE id=1`).Scan(&raw, &rev)
	var c Content
	if err == nil {
		err = json.Unmarshal(raw, &c)
	}
	return c, rev, err
}

var ErrConflict = errors.New("Content was changed in another session. Reload before saving.")

func (s *Store) Save(ctx context.Context, c Content, rev int64) error {
	if err := c.Validate(); err != nil {
		return err
	}
	raw, err := json.Marshal(c)
	if err != nil {
		return err
	}
	result, err := s.DB.Exec(ctx, `UPDATE site_content SET body=$1,revision=revision+1 WHERE id=1 AND revision=$2`, raw, rev)
	if err == nil && result.RowsAffected() != 1 {
		return ErrConflict
	}
	return err
}
func (s *Store) Allow(ctx context.Context, key string, max int, window time.Duration) bool {
	var hits int
	err := s.DB.QueryRow(ctx, `INSERT INTO request_limits(key,hits,resets) VALUES($1,1,now()+$2::interval)
 ON CONFLICT(key) DO UPDATE SET hits=CASE WHEN request_limits.resets<now() THEN 1 ELSE request_limits.hits+1 END,
 resets=CASE WHEN request_limits.resets<now() THEN excluded.resets ELSE request_limits.resets END RETURNING hits`, key, window.String()).Scan(&hits)
	return err == nil && hits <= max
}
func (s *Store) Cleanup(ctx context.Context) {
	s.DB.Exec(ctx, `DELETE FROM admin_sessions WHERE expires<now()`)
	s.DB.Exec(ctx, `DELETE FROM request_limits WHERE resets<now()`)
}
