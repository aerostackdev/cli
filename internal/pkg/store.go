package pkg

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS symbols (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	kind TEXT NOT NULL,
	file_path TEXT NOT NULL,
	line_start INTEGER NOT NULL,
	line_end INTEGER NOT NULL,
	content TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS relationships (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	source_id INTEGER NOT NULL,
	target_id INTEGER NOT NULL,
	type TEXT NOT NULL,
	FOREIGN KEY(source_id) REFERENCES symbols(id),
	FOREIGN KEY(target_id) REFERENCES symbols(id)
);

CREATE INDEX IF NOT EXISTS idx_symbols_name ON symbols(name);
CREATE INDEX IF NOT EXISTS idx_symbols_file_path ON symbols(file_path);
`

type Store struct {
	db *sql.DB
}

func NewStore(projectRoot string) (*Store, error) {
	aerostackDir := filepath.Join(projectRoot, ".aerostack")
	if err := os.MkdirAll(aerostackDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .aerostack directory: %w", err)
	}

	dbPath := filepath.Join(aerostackDir, "pkg.db")
	fmt.Println("DEBUG: Opening DB at", dbPath)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open pkg database: %w", err)
	}

	fmt.Println("DEBUG: Executing DB schema")
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	fmt.Println("DEBUG: Store initialized successfully")
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// AddSymbol inserts a new symbol into the store
func (s *Store) AddSymbol(name, kind, filePath, content string, lineStart, lineEnd int) (int64, error) {
	res, err := s.db.Exec(`
		INSERT INTO symbols (name, kind, file_path, content, line_start, line_end)
		VALUES (?, ?, ?, ?, ?, ?)
	`, name, kind, filePath, content, lineStart, lineEnd)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetSymbolsByKind returns all symbols matching the given kind
func (s *Store) GetSymbolsByKind(kind string) ([]Symbol, error) {
	rows, err := s.db.Query("SELECT id, name, kind, file_path, line_start, line_end FROM symbols WHERE kind = ?", kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var symbols []Symbol
	for rows.Next() {
		var sym Symbol
		if err := rows.Scan(&sym.ID, &sym.Name, &sym.Kind, &sym.FilePath, &sym.LineStart, &sym.LineEnd); err != nil {
			return nil, err
		}
		symbols = append(symbols, sym)
	}
	return symbols, nil
}

// SearchSymbols finds symbols by name (LIKE) or kind
func (s *Store) SearchSymbols(query string) ([]Symbol, error) {
	pattern := "%" + query + "%"
	rows, err := s.db.Query(`
		SELECT id, name, kind, file_path, line_start, line_end FROM symbols
		WHERE name LIKE ? OR kind LIKE ?
		ORDER BY name
	`, pattern, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var symbols []Symbol
	for rows.Next() {
		var sym Symbol
		if err := rows.Scan(&sym.ID, &sym.Name, &sym.Kind, &sym.FilePath, &sym.LineStart, &sym.LineEnd); err != nil {
			return nil, err
		}
		symbols = append(symbols, sym)
	}
	return symbols, nil
}

type Symbol struct {
	ID        int64
	Name      string
	Kind      string
	FilePath  string
	LineStart int
	LineEnd   int
	Content   string
}
