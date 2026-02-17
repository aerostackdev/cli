package pkg

type Mapper struct {
	store *Store
}

func NewMapper(store *Store) *Mapper {
	return &Mapper{store: store}
}

// MapRelationships scans existing symbols and creates links based on heuristics
func (m *Mapper) MapRelationships() error {
	// 1. Link API Handlers to DB Models
	// Heuristic: If a handler function uses a model struct variable or calls a DB method
	// For MVP, we'll try to match names. e.g. "GetProduct" handler relates to "Product" model

	// Get all API handlers (Go functions starting with Handle or Get/Post...)
	// This requires better symbol classification in Indexer first.
	// For now, let's just look for name overlap.

	// Example: Find a "Product" struct and specific "GetProduct" function

	// This is a placeholder for the logic.
	// Real logic needs AST analysis of function bodies (which definitions are used).

	return nil
}

// FindRelatedSymbols returns symbols related to the given symbol ID
func (m *Mapper) FindRelatedSymbols(id int64) ([]Symbol, error) {
	// Query relationships table
	// SELECT * FROM relationships WHERE source_id = ?
	// JOIN symbols ON target_id = id

	query := `
    SELECT s.id, s.name, s.kind, s.file_path, s.line_start, s.line_end
    FROM relationships r
    JOIN symbols s ON r.target_id = s.id
    WHERE r.source_id = ?
    `
	rows, err := m.store.db.Query(query, id)
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

func (m *Mapper) createLink(sourceID, targetID int64, relType string) error {
	_, err := m.store.db.Exec(`
		INSERT INTO relationships (source_id, target_id, type)
		VALUES (?, ?, ?)
	`, sourceID, targetID, relType)
	return err
}
