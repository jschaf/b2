package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jschaf/b2/pkg/texts"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type Store interface {
	AllRawFetches() ([]RawFetch, error)
}

type RawFetch struct {
	URL       string
	FetchTime time.Time
	Content   string
	Assets    []string
}

type SQLiteStore struct {
	db     *sql.DB
	logger *zap.Logger
}

const (
	sqlitePath = "pkg/db/b2.db"
)

func NewSQLiteStore(l *zap.Logger) *SQLiteStore {
	return &SQLiteStore{logger: l.Named("sqlite")}
}

func (s *SQLiteStore) Open() error {
	s.logger.Info("opening", zap.String("db", sqlitePath))
	db, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		return fmt.Errorf("open sqlite db: %w", err)
	}
	s.db = db
	return nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) AllRawFetches() ([]RawFetch, error) {
	query := texts.Dedent(`
		SELECT rf.url, rf.fetch_time, rf.content, group_concat(ra.local_path, '!SPLIT!')
		FROM raw_fetch rf
					 LEFT JOIN raw_asset ra ON rf.url = ra.url;
  `)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("AllRawFetches query: %w", err)
	}
	fetches := make([]RawFetch, 0)
	for rows.Next() {
		var url string
		var fetchTime time.Time
		var content string
		var localPaths string
		if err := rows.Scan(&url, &fetchTime, &content, &localPaths); err != nil {
			return nil, fmt.Errorf("AllRawFetches rows scan: %w", err)
		}
		fetches = append(fetches, RawFetch{
			URL:       url,
			FetchTime: fetchTime,
			Content:   content,
			Assets:    strings.Split(localPaths, "!SPLIT!"),
		})
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("AllRawFetches rows iteration: %w", err)
	}

	if rows.Close() != nil {
		return nil, fmt.Errorf("AllRawFetches rows close: %w", err)
	}

	return fetches, nil
}
