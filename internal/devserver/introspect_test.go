package devserver

import (
	"strings"
	"testing"

	"github.com/aerostackdev/cli/internal/api"
)

// ─── mapSQLiteType ──────────────────────────────────────────────

func TestMapSQLiteType_Integer(t *testing.T) {
	cases := []string{"INTEGER", "INT", "BIGINT", "SMALLINT", "TINYINT", "MEDIUMINT", "int"}
	for _, c := range cases {
		if got := mapSQLiteType(c); got != "number" {
			t.Errorf("mapSQLiteType(%q) = %q, want %q", c, got, "number")
		}
	}
}

func TestMapSQLiteType_Text(t *testing.T) {
	cases := []string{"TEXT", "VARCHAR", "CHAR", "CHARACTER", "text", "varchar(255)"}
	for _, c := range cases {
		if got := mapSQLiteType(c); got != "string" {
			t.Errorf("mapSQLiteType(%q) = %q, want %q", c, got, "string")
		}
	}
}

func TestMapSQLiteType_UUID(t *testing.T) {
	if got := mapSQLiteType("UUID"); got != "string" {
		t.Errorf("mapSQLiteType(UUID) = %q, want string", got)
	}
}

func TestMapSQLiteType_Real(t *testing.T) {
	cases := []string{"REAL", "FLOAT", "DOUBLE", "DOUBLE PRECISION"}
	for _, c := range cases {
		if got := mapSQLiteType(c); got != "number" {
			t.Errorf("mapSQLiteType(%q) = %q, want %q", c, got, "number")
		}
	}
}

func TestMapSQLiteType_Blob(t *testing.T) {
	if got := mapSQLiteType("BLOB"); got != "Uint8Array" {
		t.Errorf("mapSQLiteType(BLOB) = %q, want Uint8Array", got)
	}
}

func TestMapSQLiteType_Boolean(t *testing.T) {
	if got := mapSQLiteType("BOOLEAN"); got != "boolean" {
		t.Errorf("mapSQLiteType(BOOLEAN) = %q, want boolean", got)
	}
	if got := mapSQLiteType("BOOL"); got != "boolean" {
		t.Errorf("mapSQLiteType(BOOL) = %q, want boolean", got)
	}
}

func TestMapSQLiteType_Unknown(t *testing.T) {
	cases := []string{"", "GEOMETRY", "CUSTOM_TYPE", "JSON"}
	for _, c := range cases {
		if got := mapSQLiteType(c); got != "any" {
			t.Errorf("mapSQLiteType(%q) = %q, want any", c, got)
		}
	}
}

func TestMapSQLiteType_CaseInsensitive(t *testing.T) {
	if got := mapSQLiteType("integer"); got != "number" {
		t.Errorf("mapSQLiteType(integer) = %q, want number", got)
	}
	if got := mapSQLiteType("Text"); got != "string" {
		t.Errorf("mapSQLiteType(Text) = %q, want string", got)
	}
}

// ─── mapPostgresType ────────────────────────────────────────────

func TestMapPostgresType_Integer(t *testing.T) {
	cases := []string{"integer", "int", "bigint", "smallint", "numeric", "real", "double precision"}
	for _, c := range cases {
		if got := mapPostgresType(c); got != "number" {
			t.Errorf("mapPostgresType(%q) = %q, want number", c, got)
		}
	}
}

func TestMapPostgresType_Text(t *testing.T) {
	cases := []string{"text", "varchar", "character varying", "char", "character", "uuid"}
	for _, c := range cases {
		if got := mapPostgresType(c); got != "string" {
			t.Errorf("mapPostgresType(%q) = %q, want string", c, got)
		}
	}
}

func TestMapPostgresType_Boolean(t *testing.T) {
	if got := mapPostgresType("boolean"); got != "boolean" {
		t.Errorf("mapPostgresType(boolean) = %q, want boolean", got)
	}
}

func TestMapPostgresType_Timestamp(t *testing.T) {
	if got := mapPostgresType("timestamp"); got != "string" {
		t.Errorf("mapPostgresType(timestamp) = %q, want string", got)
	}
	if got := mapPostgresType("date"); got != "string" {
		t.Errorf("mapPostgresType(date) = %q, want string", got)
	}
}

func TestMapPostgresType_JSON(t *testing.T) {
	if got := mapPostgresType("json"); got != "any" {
		t.Errorf("mapPostgresType(json) = %q, want any", got)
	}
	if got := mapPostgresType("jsonb"); got != "any" {
		t.Errorf("mapPostgresType(jsonb) = %q, want any", got)
	}
}

func TestMapPostgresType_Unknown(t *testing.T) {
	cases := []string{"", "geometry", "cidr", "inet", "bytea"}
	for _, c := range cases {
		if got := mapPostgresType(c); got != "any" {
			t.Errorf("mapPostgresType(%q) = %q, want any", c, got)
		}
	}
}

func TestMapPostgresType_CaseInsensitive(t *testing.T) {
	// Function lowercases input
	if got := mapPostgresType("INTEGER"); got != "number" {
		t.Errorf("mapPostgresType(INTEGER) = %q, want number", got)
	}
}

// ─── toPascalCase ───────────────────────────────────────────────

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"users", "Users"},
		{"user_profiles", "UserProfiles"},
		{"order-items", "OrderItems"},
		{"a_b_c", "ABC"},
		{"hello_world", "HelloWorld"},
		{"UPPER_CASE", "UpperCase"},
		{"single", "Single"},
		{"", ""},
		{"already", "Already"},
		{"a", "A"},
		{"multi_word_name_here", "MultiWordNameHere"},
		{"with-hyphens-and_underscores", "WithHyphensAndUnderscores"},
	}
	for _, tt := range tests {
		if got := toPascalCase(tt.input); got != tt.want {
			t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ─── parseTableNames ────────────────────────────────────────────

func TestParseTableNames_WithTableOutput(t *testing.T) {
	output := `┌──────────────┐
│     name     │
├──────────────┤
│ users        │
│ posts        │
│ comments     │
└──────────────┘`
	names := parseTableNames(output)
	if len(names) != 3 {
		t.Fatalf("parseTableNames got %d names, want 3", len(names))
	}
	expected := []string{"users", "posts", "comments"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("parseTableNames[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestParseTableNames_EmptyOutput(t *testing.T) {
	names := parseTableNames("")
	if len(names) != 0 {
		t.Fatalf("parseTableNames empty = %d names, want 0", len(names))
	}
}

func TestParseTableNames_HeaderOnly(t *testing.T) {
	output := `┌──────────────┐
│     name     │
├──────────────┤
└──────────────┘`
	names := parseTableNames(output)
	if len(names) != 0 {
		t.Fatalf("parseTableNames header-only = %d names, want 0", len(names))
	}
}

func TestParseTableNames_SkipsNonTableLines(t *testing.T) {
	output := "Some random output\nAnother line\n"
	names := parseTableNames(output)
	if len(names) != 0 {
		t.Fatalf("parseTableNames non-table = %d names, want 0", len(names))
	}
}

// ─── GenerateTypeScript ─────────────────────────────────────────

func TestGenerateTypeScript_SingleTable(t *testing.T) {
	schemas := []TableSchema{
		{
			Name: "users",
			Columns: []ColumnSchema{
				{Name: "id", Type: "number", IsNullable: false, IsPrimary: true},
				{Name: "name", Type: "string", IsNullable: false, IsPrimary: false},
				{Name: "email", Type: "string", IsNullable: true, IsPrimary: false},
			},
		},
	}

	result := GenerateTypeScript(schemas, nil)

	if !strings.Contains(result, "export interface Users {") {
		t.Error("missing Users interface")
	}
	if !strings.Contains(result, "id: number;") {
		t.Error("missing id field")
	}
	if !strings.Contains(result, "name: string;") {
		t.Error("missing name field")
	}
	if !strings.Contains(result, "email?: string;") {
		t.Error("missing nullable email field (should have ?)")
	}
}

func TestGenerateTypeScript_WithSourceBinding(t *testing.T) {
	schemas := []TableSchema{
		{
			Name:          "users",
			Columns:       []ColumnSchema{{Name: "id", Type: "number"}},
			SourceBinding: "DB",
		},
	}

	result := GenerateTypeScript(schemas, nil)

	if !strings.Contains(result, "export interface DbUsers {") {
		t.Errorf("expected DbUsers interface, got:\n%s", result)
	}
}

func TestGenerateTypeScript_MultipleTables(t *testing.T) {
	schemas := []TableSchema{
		{Name: "posts", Columns: []ColumnSchema{{Name: "id", Type: "number"}}},
		{Name: "users", Columns: []ColumnSchema{{Name: "id", Type: "number"}}},
	}

	result := GenerateTypeScript(schemas, nil)

	if !strings.Contains(result, "export interface Posts {") {
		t.Error("missing Posts interface")
	}
	if !strings.Contains(result, "export interface Users {") {
		t.Error("missing Users interface")
	}
	// Should be sorted
	postsIdx := strings.Index(result, "export interface Posts")
	usersIdx := strings.Index(result, "export interface Users")
	if postsIdx > usersIdx {
		t.Error("tables not sorted alphabetically")
	}
}

func TestGenerateTypeScript_SortsByBindingFirst(t *testing.T) {
	schemas := []TableSchema{
		{Name: "users", SourceBinding: "PG", Columns: []ColumnSchema{{Name: "id", Type: "number"}}},
		{Name: "users", SourceBinding: "DB", Columns: []ColumnSchema{{Name: "id", Type: "number"}}},
	}

	result := GenerateTypeScript(schemas, nil)

	dbIdx := strings.Index(result, "DbUsers")
	pgIdx := strings.Index(result, "PgUsers")
	if dbIdx > pgIdx {
		t.Error("DB binding should come before PG binding")
	}
}

func TestGenerateTypeScript_WithCollections(t *testing.T) {
	meta := &api.ProjectMetadata{
		Collections: []struct {
			ID                string  `json:"id"`
			Name              string  `json:"name"`
			Slug              string  `json:"slug"`
			SchemaComponentID string  `json:"schema_component_id"`
			Schema            *string `json:"schema"`
		}{
			{Slug: "blog-posts"},
			{Slug: "faq"},
		},
	}

	result := GenerateTypeScript(nil, meta)

	if !strings.Contains(result, "export interface BlogPostsItem {") {
		t.Error("missing BlogPostsItem interface")
	}
	if !strings.Contains(result, "export interface FaqItem {") {
		t.Error("missing FaqItem interface")
	}
	if !strings.Contains(result, `"blog-posts": BlogPostsItem;`) {
		t.Error("missing blog-posts collection mapping")
	}
}

func TestGenerateTypeScript_WithHooks(t *testing.T) {
	meta := &api.ProjectMetadata{
		Hooks: []struct {
			ID             string   `json:"id"`
			Name           string   `json:"name"`
			Slug           string   `json:"slug"`
			EventType      string   `json:"event_type"`
			Type           string   `json:"type"`
			IsPublic       int      `json:"is_public"`
			AllowedMethods []string `json:"allowed_methods"`
		}{
			{Slug: "send-email"},
			{Slug: "process-payment"},
		},
	}

	result := GenerateTypeScript(nil, meta)

	if !strings.Contains(result, "export interface CustomApiSchema {") {
		t.Error("missing CustomApiSchema interface")
	}
	if !strings.Contains(result, `"send-email"`) {
		t.Error("missing send-email hook")
	}
	if !strings.Contains(result, `"process-payment"`) {
		t.Error("missing process-payment hook")
	}
	if !strings.Contains(result, "customApis: CustomApiSchema;") {
		t.Error("missing customApis: CustomApiSchema in ProjectSchema")
	}
}

func TestGenerateTypeScript_WithoutHooks(t *testing.T) {
	result := GenerateTypeScript(nil, nil)

	if !strings.Contains(result, "customApis: Record<string, { params: any; response: any }>;") {
		t.Error("missing default customApis type")
	}
}

func TestGenerateTypeScript_ProjectSchemaHasAllSections(t *testing.T) {
	result := GenerateTypeScript(nil, nil)

	if !strings.Contains(result, "export interface ProjectSchema {") {
		t.Error("missing ProjectSchema interface")
	}
	if !strings.Contains(result, "collections: {") {
		t.Error("missing collections section")
	}
	if !strings.Contains(result, "db: {") {
		t.Error("missing db section")
	}
	if !strings.Contains(result, "queues: Record<string, any>;") {
		t.Error("missing queues section")
	}
	if !strings.Contains(result, "cache: Record<string, any>;") {
		t.Error("missing cache section")
	}
}

func TestGenerateTypeScript_DbSectionWithBinding(t *testing.T) {
	schemas := []TableSchema{
		{Name: "users", SourceBinding: "DB", Columns: []ColumnSchema{{Name: "id", Type: "number"}}},
	}

	result := GenerateTypeScript(schemas, nil)

	if !strings.Contains(result, `"DB.users": DbUsers;`) {
		t.Errorf("missing DB.users key in db section, got:\n%s", result)
	}
}

func TestGenerateTypeScript_HeaderComment(t *testing.T) {
	result := GenerateTypeScript(nil, nil)
	if !strings.HasPrefix(result, "// Generated by Aerostack CLI") {
		t.Error("missing header comment")
	}
}
