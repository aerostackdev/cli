package pkg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

type Indexer struct {
	store *Store
}

func NewIndexer(store *Store) *Indexer {
	return &Indexer{store: store}
}

// IndexProject walks the project directory and indexes all supported files
func (i *Indexer) IndexProject(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == "node_modules" || info.Name() == ".git" || info.Name() == ".aerostack" {
				return filepath.SkipDir
			}
			return nil
		}
		return i.IndexFile(path)
	})
}

// IndexFile parses a single file and stores its symbols
func (i *Indexer) IndexFile(path string) error {
	lang := getLanguage(path)
	if lang == nil {
		return nil // Unsupported file type
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", path, err)
	}

	// Simple extraction logic - extend this with queries later
	rootNode := tree.RootNode()
	i.extractSymbols(rootNode, path, content)

	return nil
}

func getLanguage(path string) *sitter.Language {
	ext := filepath.Ext(path)
	switch ext {
	case ".go":
		return golang.GetLanguage()
	case ".ts":
		if strings.HasSuffix(path, ".d.ts") {
			return nil // Skip definition files for now
		}
		return typescript.GetLanguage()
	case ".tsx":
		return tsx.GetLanguage()
	default:
		return nil
	}
}

func (i *Indexer) extractSymbols(node *sitter.Node, filePath string, content []byte) {
	count := node.ChildCount()
	for j := 0; j < int(count); j++ {
		child := node.Child(j)
		nodeType := child.Type()

		if nodeType == "function_declaration" || nodeType == "class_declaration" || nodeType == "type_declaration" || nodeType == "method_declaration" || nodeType == "interface_declaration" || nodeType == "type_alias_declaration" {
			name := extractName(child, content)
			if name == "" {
				name = "anonymous"
			}
			i.store.AddSymbol(name, nodeType, filePath, "", int(child.StartPoint().Row), int(child.EndPoint().Row))
		}

		// Recurse into export statements and block bodies for nested declarations
		if nodeType == "export_statement" || nodeType == "export_default_declaration" {
			for k := 0; k < int(child.ChildCount()); k++ {
				sub := child.Child(k)
				if sub.Type() == "function_declaration" || sub.Type() == "class_declaration" {
					name := extractName(sub, content)
					if name == "" {
						name = "default"
					}
					i.store.AddSymbol(name, sub.Type(), filePath, "", int(sub.StartPoint().Row), int(sub.EndPoint().Row))
				}
			}
		}
	}
}

// extractName gets the identifier for a declaration node using tree-sitter field names
func extractName(node *sitter.Node, content []byte) string {
	// Try common field names: name, value (for some nodes)
	for _, field := range []string{"name", "value"} {
		if n := node.ChildByFieldName(field); n != nil {
			return string(content[n.StartByte():n.EndByte()])
		}
	}
	// Fallback: first identifier-like child (identifier, type_identifier, field_identifier)
	for i := 0; i < int(node.ChildCount()); i++ {
		c := node.Child(i)
		t := c.Type()
		if t == "identifier" || t == "type_identifier" || t == "field_identifier" || t == "property_identifier" {
			return string(content[c.StartByte():c.EndByte()])
		}
	}
	return ""
}
