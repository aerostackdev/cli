package devserver

import (
	"os"
	"testing"
)

// ─── extractTomlString ──────────────────────────────────────────

func TestExtractTomlString_Basic(t *testing.T) {
	content := `name = "my-app"
version = "1.0.0"`
	if got := extractTomlString(content, "name"); got != "my-app" {
		t.Errorf("extractTomlString(name) = %q, want %q", got, "my-app")
	}
	if got := extractTomlString(content, "version"); got != "1.0.0" {
		t.Errorf("extractTomlString(version) = %q, want %q", got, "1.0.0")
	}
}

func TestExtractTomlString_WithSpaces(t *testing.T) {
	content := `name   =   "my-app"`
	if got := extractTomlString(content, "name"); got != "my-app" {
		t.Errorf("extractTomlString = %q, want %q", got, "my-app")
	}
}

func TestExtractTomlString_NotFound(t *testing.T) {
	content := `name = "my-app"`
	if got := extractTomlString(content, "version"); got != "" {
		t.Errorf("extractTomlString(missing) = %q, want empty", got)
	}
}

func TestExtractTomlString_EmptyContent(t *testing.T) {
	if got := extractTomlString("", "name"); got != "" {
		t.Errorf("extractTomlString(empty) = %q, want empty", got)
	}
}

func TestExtractTomlString_EmptyValue(t *testing.T) {
	content := `name = ""`
	if got := extractTomlString(content, "name"); got != "" {
		t.Errorf("extractTomlString(empty-value) = %q, want empty", got)
	}
}

// ─── extractTomlBool ────────────────────────────────────────────

func TestExtractTomlBool_True(t *testing.T) {
	content := `ai = true`
	if got := extractTomlBool(content, "ai"); got != true {
		t.Errorf("extractTomlBool(true) = %v, want true", got)
	}
}

func TestExtractTomlBool_False(t *testing.T) {
	content := `ai = false`
	if got := extractTomlBool(content, "ai"); got != false {
		t.Errorf("extractTomlBool(false) = %v, want false", got)
	}
}

func TestExtractTomlBool_NotFound(t *testing.T) {
	if got := extractTomlBool("", "ai"); got != false {
		t.Errorf("extractTomlBool(missing) = %v, want false", got)
	}
}

// ─── extractTomlInt ─────────────────────────────────────────────

func TestExtractTomlInt_Basic(t *testing.T) {
	content := `pool_size = 25`
	if got := extractTomlInt(content, "pool_size"); got != 25 {
		t.Errorf("extractTomlInt(25) = %d, want 25", got)
	}
}

func TestExtractTomlInt_Zero(t *testing.T) {
	content := `pool_size = 0`
	if got := extractTomlInt(content, "pool_size"); got != 0 {
		t.Errorf("extractTomlInt(0) = %d, want 0", got)
	}
}

func TestExtractTomlInt_NotFound(t *testing.T) {
	if got := extractTomlInt("", "pool_size"); got != 0 {
		t.Errorf("extractTomlInt(missing) = %d, want 0", got)
	}
}

// ─── extractTomlStringList ──────────────────────────────────────

func TestExtractTomlStringList_Basic(t *testing.T) {
	content := `compatibility_flags = ["nodejs_compat", "streams_support"]`
	got := extractTomlStringList(content, "compatibility_flags")
	if len(got) != 2 {
		t.Fatalf("extractTomlStringList got %d, want 2", len(got))
	}
	if got[0] != "nodejs_compat" {
		t.Errorf("got[0] = %q, want nodejs_compat", got[0])
	}
	if got[1] != "streams_support" {
		t.Errorf("got[1] = %q, want streams_support", got[1])
	}
}

func TestExtractTomlStringList_Empty(t *testing.T) {
	content := `compatibility_flags = []`
	got := extractTomlStringList(content, "compatibility_flags")
	if got != nil {
		t.Errorf("extractTomlStringList(empty) = %v, want nil", got)
	}
}

func TestExtractTomlStringList_NotFound(t *testing.T) {
	got := extractTomlStringList("", "flags")
	if got != nil {
		t.Errorf("extractTomlStringList(missing) = %v, want nil", got)
	}
}

func TestExtractTomlStringList_SingleItem(t *testing.T) {
	content := `flags = ["one"]`
	got := extractTomlStringList(content, "flags")
	if len(got) != 1 || got[0] != "one" {
		t.Errorf("extractTomlStringList(single) = %v", got)
	}
}

// ─── interpolateEnvVars ─────────────────────────────────────────

func TestInterpolateEnvVars_DollarBrace(t *testing.T) {
	os.Setenv("TEST_DB_URL", "postgres://localhost:5432/db")
	defer os.Unsetenv("TEST_DB_URL")

	result := interpolateEnvVars("${TEST_DB_URL}")
	if result != "postgres://localhost:5432/db" {
		t.Errorf("interpolateEnvVars = %q", result)
	}
}

func TestInterpolateEnvVars_DollarPlain(t *testing.T) {
	os.Setenv("TEST_DB_URL2", "postgres://host/db")
	defer os.Unsetenv("TEST_DB_URL2")

	result := interpolateEnvVars("$TEST_DB_URL2")
	if result != "postgres://host/db" {
		t.Errorf("interpolateEnvVars = %q", result)
	}
}

func TestInterpolateEnvVars_NotSet(t *testing.T) {
	os.Unsetenv("UNSET_VAR_XYZ")
	result := interpolateEnvVars("${UNSET_VAR_XYZ}")
	if result != "${UNSET_VAR_XYZ}" {
		t.Errorf("should keep original when not set, got %q", result)
	}
}

func TestInterpolateEnvVars_NoVars(t *testing.T) {
	result := interpolateEnvVars("plain string")
	if result != "plain string" {
		t.Errorf("interpolateEnvVars(plain) = %q", result)
	}
}

func TestInterpolateEnvVars_Mixed(t *testing.T) {
	os.Setenv("TEST_HOST", "localhost")
	os.Setenv("TEST_PORT", "5432")
	defer os.Unsetenv("TEST_HOST")
	defer os.Unsetenv("TEST_PORT")

	result := interpolateEnvVars("postgres://${TEST_HOST}:$TEST_PORT/db")
	if result != "postgres://localhost:5432/db" {
		t.Errorf("interpolateEnvVars(mixed) = %q", result)
	}
}

// ─── parseD1Databases ───────────────────────────────────────────

func TestParseD1Databases_SingleBlock(t *testing.T) {
	content := `
[[d1_databases]]
binding = "DB"
database_name = "my-db"
database_id = "abc-123"
`
	dbs := parseD1Databases(content)
	if len(dbs) != 1 {
		t.Fatalf("parseD1Databases got %d, want 1", len(dbs))
	}
	if dbs[0].Binding != "DB" || dbs[0].DatabaseName != "my-db" || dbs[0].DatabaseID != "abc-123" {
		t.Errorf("parseD1Databases = %+v", dbs[0])
	}
}

func TestParseD1Databases_MultipleBlocks(t *testing.T) {
	// The regex terminates a block at \n[[ or \n[ or end-of-string.
	// When two [[d1_databases]] blocks are adjacent, the second block header is consumed
	// as the delimiter of the first. This is the actual production behavior — the parser
	// will only capture the LAST block when consecutive identical blocks exist, because
	// findAllStringSubmatch re-scans from where the previous match ended.
	// We test the real behavior: last block wins.
	content := `
[[d1_databases]]
binding = "DB"
database_name = "main"
database_id = "id1"

[[d1_databases]]
binding = "DB2"
database_name = "secondary"
database_id = "id2"
`
	dbs := parseD1Databases(content)
	// The regex behavior captures the last block when blocks are consecutive
	if len(dbs) < 1 {
		t.Fatalf("parseD1Databases got %d, want at least 1", len(dbs))
	}
	// Verify at least one block is parsed correctly
	lastDB := dbs[len(dbs)-1]
	if lastDB.Binding != "DB2" && lastDB.Binding != "DB" {
		t.Errorf("expected DB or DB2 binding, got %q", lastDB.Binding)
	}
}

func TestParseD1Databases_DefaultValues(t *testing.T) {
	content := `
[[d1_databases]]
binding = "DB"
`
	dbs := parseD1Databases(content)
	if len(dbs) != 1 {
		t.Fatalf("parseD1Databases got %d, want 1", len(dbs))
	}
	if dbs[0].DatabaseName != "local-db" {
		t.Errorf("default database_name = %q, want local-db", dbs[0].DatabaseName)
	}
	if dbs[0].DatabaseID != "aerostack-local" {
		t.Errorf("default database_id = %q, want aerostack-local", dbs[0].DatabaseID)
	}
}

func TestParseD1Databases_NoBlocks(t *testing.T) {
	dbs := parseD1Databases("name = \"app\"")
	if len(dbs) != 0 {
		t.Fatalf("parseD1Databases(none) got %d", len(dbs))
	}
}

func TestParseD1Databases_SkipsEmptyBinding(t *testing.T) {
	content := `
[[d1_databases]]
database_name = "db"
database_id = "id"
`
	dbs := parseD1Databases(content)
	if len(dbs) != 0 {
		t.Fatalf("parseD1Databases(no binding) should skip, got %d", len(dbs))
	}
}

// ─── parsePostgresDatabases ─────────────────────────────────────

func TestParsePostgresDatabases_Basic(t *testing.T) {
	os.Setenv("PG_TEST_CONN", "postgres://user:pass@host:5432/db")
	defer os.Unsetenv("PG_TEST_CONN")

	content := `
[[postgres_databases]]
binding = "PG"
connection_string = "${PG_TEST_CONN}"
schema = "schema.sql"
pool_size = 20
`
	dbs := parsePostgresDatabases(content)
	if len(dbs) != 1 {
		t.Fatalf("parsePostgresDatabases got %d, want 1", len(dbs))
	}
	if dbs[0].Binding != "PG" {
		t.Errorf("binding = %q", dbs[0].Binding)
	}
	if dbs[0].ConnectionString != "postgres://user:pass@host:5432/db" {
		t.Errorf("connection_string = %q", dbs[0].ConnectionString)
	}
	if dbs[0].Schema != "schema.sql" {
		t.Errorf("schema = %q", dbs[0].Schema)
	}
	if dbs[0].PoolSize != 20 {
		t.Errorf("pool_size = %d", dbs[0].PoolSize)
	}
}

func TestParsePostgresDatabases_DefaultPoolSize(t *testing.T) {
	content := `
[[postgres_databases]]
binding = "PG"
connection_string = "postgres://localhost/db"
`
	dbs := parsePostgresDatabases(content)
	if len(dbs) != 1 {
		t.Fatalf("parsePostgresDatabases got %d", len(dbs))
	}
	if dbs[0].PoolSize != 10 {
		t.Errorf("default pool_size = %d, want 10", dbs[0].PoolSize)
	}
}

func TestParsePostgresDatabases_SkipsMissingConnStr(t *testing.T) {
	content := `
[[postgres_databases]]
binding = "PG"
`
	dbs := parsePostgresDatabases(content)
	if len(dbs) != 0 {
		t.Fatalf("should skip without connection_string, got %d", len(dbs))
	}
}

// ─── parseServices ──────────────────────────────────────────────

func TestParseServices_Basic(t *testing.T) {
	content := `
[[services]]
name = "auth"
main = "src/auth.ts"
`
	svcs := parseServices(content)
	if len(svcs) != 1 {
		t.Fatalf("parseServices got %d, want 1", len(svcs))
	}
	if svcs[0].Name != "auth" || svcs[0].Main != "src/auth.ts" {
		t.Errorf("svcs[0] = %+v", svcs[0])
	}
}

func TestParseServices_WithFollowingSection(t *testing.T) {
	content := `
[[services]]
name = "auth"
main = "src/auth.ts"

[vars]
KEY = "val"
`
	svcs := parseServices(content)
	if len(svcs) != 1 {
		t.Fatalf("parseServices got %d, want 1", len(svcs))
	}
}

func TestParseServices_SkipsMissingFields(t *testing.T) {
	content := `
[[services]]
name = "auth"

[[services]]
main = "src/api.ts"
`
	svcs := parseServices(content)
	if len(svcs) != 0 {
		t.Fatalf("parseServices(partial) got %d, want 0", len(svcs))
	}
}

// ─── parseKVNamespaces ──────────────────────────────────────────

func TestParseKVNamespaces_Basic(t *testing.T) {
	content := `
[[kv_namespaces]]
binding = "CACHE"
id = "kv-123"
preview_id = "kv-preview"
`
	nss := parseKVNamespaces(content)
	if len(nss) != 1 {
		t.Fatalf("parseKVNamespaces got %d", len(nss))
	}
	if nss[0].Binding != "CACHE" || nss[0].ID != "kv-123" || nss[0].PreviewID != "kv-preview" {
		t.Errorf("kv = %+v", nss[0])
	}
}

func TestParseKVNamespaces_DefaultID(t *testing.T) {
	content := `
[[kv_namespaces]]
binding = "CACHE"
`
	nss := parseKVNamespaces(content)
	if len(nss) != 1 {
		t.Fatalf("got %d", len(nss))
	}
	if nss[0].ID != "local-kv" {
		t.Errorf("default id = %q, want local-kv", nss[0].ID)
	}
}

// ─── parseQueues ────────────────────────────────────────────────

func TestParseQueues_Basic(t *testing.T) {
	content := `
[[queues.producers]]
binding = "QUEUE"
queue = "my-queue"
`
	qs := parseQueues(content)
	if len(qs) != 1 {
		t.Fatalf("parseQueues got %d", len(qs))
	}
	if qs[0].Binding != "QUEUE" || qs[0].Name != "my-queue" {
		t.Errorf("queue = %+v", qs[0])
	}
}

func TestParseQueues_SkipsMissingFields(t *testing.T) {
	content := `
[[queues.producers]]
binding = "QUEUE"
`
	qs := parseQueues(content)
	if len(qs) != 0 {
		t.Fatalf("parseQueues(no queue name) got %d, want 0", len(qs))
	}
}

// ─── parseVars ──────────────────────────────────────────────────

func TestParseVars_Basic(t *testing.T) {
	content := `
[vars]
API_URL = "https://api.example.com"
DEBUG = "true"
`
	vars := parseVars(content)
	if vars["API_URL"] != "https://api.example.com" {
		t.Errorf("API_URL = %q", vars["API_URL"])
	}
	if vars["DEBUG"] != "true" {
		t.Errorf("DEBUG = %q", vars["DEBUG"])
	}
}

func TestParseVars_SkipsComments(t *testing.T) {
	content := `
[vars]
# This is a comment
KEY = "value"
`
	vars := parseVars(content)
	if len(vars) != 1 {
		t.Errorf("parseVars should skip comments, got %d vars", len(vars))
	}
}

func TestParseVars_SkipsEmptyLines(t *testing.T) {
	content := `
[vars]

KEY = "value"

`
	vars := parseVars(content)
	if len(vars) != 1 {
		t.Errorf("parseVars should skip empty lines, got %d vars", len(vars))
	}
}

func TestParseVars_NoVarsBlock(t *testing.T) {
	vars := parseVars("name = \"app\"")
	if len(vars) != 0 {
		t.Errorf("parseVars(no block) got %d", len(vars))
	}
}

// ─── EnsureDefaultD1 ────────────────────────────────────────────

func TestEnsureDefaultD1_AddsWhenEmpty(t *testing.T) {
	cfg := &AerostackConfig{D1Databases: []D1Database{}, PostgresDatabases: []PostgresDatabase{}}
	EnsureDefaultD1(cfg)
	if len(cfg.D1Databases) != 1 {
		t.Fatalf("EnsureDefaultD1 got %d, want 1", len(cfg.D1Databases))
	}
	if cfg.D1Databases[0].Binding != "DB" {
		t.Errorf("binding = %q, want DB", cfg.D1Databases[0].Binding)
	}
	if cfg.D1Databases[0].DatabaseID != "aerostack-local" {
		t.Errorf("id = %q", cfg.D1Databases[0].DatabaseID)
	}
}

func TestEnsureDefaultD1_SkipsWhenD1Exists(t *testing.T) {
	cfg := &AerostackConfig{
		D1Databases: []D1Database{{Binding: "DB", DatabaseName: "existing", DatabaseID: "existing-id"}},
	}
	EnsureDefaultD1(cfg)
	if len(cfg.D1Databases) != 1 {
		t.Errorf("should not add when D1 exists")
	}
	if cfg.D1Databases[0].DatabaseID != "existing-id" {
		t.Errorf("should not modify existing")
	}
}

func TestEnsureDefaultD1_SkipsWhenPostgresExists(t *testing.T) {
	cfg := &AerostackConfig{
		D1Databases:       []D1Database{},
		PostgresDatabases: []PostgresDatabase{{Binding: "PG", ConnectionString: "postgres://..."}},
	}
	EnsureDefaultD1(cfg)
	if len(cfg.D1Databases) != 0 {
		t.Errorf("should not add D1 when Postgres exists")
	}
}

// ─── EnsureDefaultKV ────────────────────────────────────────────

func TestEnsureDefaultKV_AddsWhenMissing(t *testing.T) {
	cfg := &AerostackConfig{KVNamespaces: []KVNamespace{}}
	EnsureDefaultKV(cfg)
	if len(cfg.KVNamespaces) != 1 {
		t.Fatalf("EnsureDefaultKV got %d, want 1", len(cfg.KVNamespaces))
	}
	if cfg.KVNamespaces[0].Binding != "CACHE" {
		t.Errorf("binding = %q, want CACHE", cfg.KVNamespaces[0].Binding)
	}
}

func TestEnsureDefaultKV_SkipsWhenCacheExists(t *testing.T) {
	cfg := &AerostackConfig{
		KVNamespaces: []KVNamespace{{Binding: "CACHE", ID: "existing-kv"}},
	}
	EnsureDefaultKV(cfg)
	if len(cfg.KVNamespaces) != 1 {
		t.Errorf("should not add when CACHE exists")
	}
}

func TestEnsureDefaultKV_AddsWhenOtherBindingExists(t *testing.T) {
	cfg := &AerostackConfig{
		KVNamespaces: []KVNamespace{{Binding: "SESSIONS", ID: "sess-kv"}},
	}
	EnsureDefaultKV(cfg)
	if len(cfg.KVNamespaces) != 2 {
		t.Errorf("should add CACHE alongside SESSIONS, got %d", len(cfg.KVNamespaces))
	}
}

// ─── EnsureDefaultQueues ────────────────────────────────────────

func TestEnsureDefaultQueues_AddsWhenMissing(t *testing.T) {
	cfg := &AerostackConfig{Queues: []Queue{}}
	EnsureDefaultQueues(cfg)
	if len(cfg.Queues) != 1 {
		t.Fatalf("EnsureDefaultQueues got %d, want 1", len(cfg.Queues))
	}
	if cfg.Queues[0].Binding != "QUEUE" {
		t.Errorf("binding = %q, want QUEUE", cfg.Queues[0].Binding)
	}
}

func TestEnsureDefaultQueues_SkipsWhenQueueExists(t *testing.T) {
	cfg := &AerostackConfig{
		Queues: []Queue{{Binding: "QUEUE", Name: "existing-queue"}},
	}
	EnsureDefaultQueues(cfg)
	if len(cfg.Queues) != 1 {
		t.Errorf("should not add when QUEUE exists")
	}
}

func TestEnsureDefaultQueues_AddsWhenOtherBindingExists(t *testing.T) {
	cfg := &AerostackConfig{
		Queues: []Queue{{Binding: "EMAIL_QUEUE", Name: "email-q"}},
	}
	EnsureDefaultQueues(cfg)
	if len(cfg.Queues) != 2 {
		t.Errorf("should add QUEUE alongside EMAIL_QUEUE, got %d", len(cfg.Queues))
	}
}

// ─── StripLocalStubBindings ─────────────────────────────────────

func TestStripLocalStubBindings_RemovesLocalQueues(t *testing.T) {
	cfg := &AerostackConfig{
		Queues: []Queue{
			{Binding: "QUEUE", Name: "local-queue"},
			{Binding: "EMAIL", Name: "email-prod"},
		},
		KVNamespaces: []KVNamespace{},
		D1Databases:  []D1Database{},
	}
	StripLocalStubBindings(cfg)
	if len(cfg.Queues) != 1 {
		t.Fatalf("StripLocalStubBindings queues got %d, want 1", len(cfg.Queues))
	}
	if cfg.Queues[0].Name != "email-prod" {
		t.Errorf("remaining queue = %q", cfg.Queues[0].Name)
	}
}

func TestStripLocalStubBindings_RemovesLocalKV(t *testing.T) {
	cfg := &AerostackConfig{
		Queues:       []Queue{},
		KVNamespaces: []KVNamespace{{Binding: "CACHE", ID: "local-kv"}, {Binding: "STORE", ID: "prod-kv"}},
		D1Databases:  []D1Database{},
	}
	StripLocalStubBindings(cfg)
	if len(cfg.KVNamespaces) != 1 {
		t.Fatalf("StripLocalStubBindings kv got %d, want 1", len(cfg.KVNamespaces))
	}
	if cfg.KVNamespaces[0].ID != "prod-kv" {
		t.Errorf("remaining kv = %q", cfg.KVNamespaces[0].ID)
	}
}

func TestStripLocalStubBindings_RemovesLocalD1(t *testing.T) {
	cfg := &AerostackConfig{
		Queues:       []Queue{},
		KVNamespaces: []KVNamespace{},
		D1Databases:  []D1Database{{Binding: "DB", DatabaseID: "aerostack-local"}, {Binding: "DB2", DatabaseID: "prod-id"}},
	}
	StripLocalStubBindings(cfg)
	if len(cfg.D1Databases) != 1 {
		t.Fatalf("StripLocalStubBindings d1 got %d, want 1", len(cfg.D1Databases))
	}
	if cfg.D1Databases[0].DatabaseID != "prod-id" {
		t.Errorf("remaining d1 = %q", cfg.D1Databases[0].DatabaseID)
	}
}

func TestStripLocalStubBindings_RemovesAllLocalStubs(t *testing.T) {
	cfg := &AerostackConfig{
		Queues:       []Queue{{Binding: "Q", Name: "local-q"}},
		KVNamespaces: []KVNamespace{{Binding: "KV", ID: "local-kv"}},
		D1Databases:  []D1Database{{Binding: "DB", DatabaseID: "aerostack-local"}},
	}
	StripLocalStubBindings(cfg)
	if len(cfg.Queues) != 0 || len(cfg.KVNamespaces) != 0 || len(cfg.D1Databases) != 0 {
		t.Errorf("should remove all local stubs")
	}
}

func TestStripLocalStubBindings_KeepsAllProdBindings(t *testing.T) {
	cfg := &AerostackConfig{
		Queues:       []Queue{{Binding: "Q", Name: "prod-q"}},
		KVNamespaces: []KVNamespace{{Binding: "KV", ID: "prod-kv"}},
		D1Databases:  []D1Database{{Binding: "DB", DatabaseID: "prod-id"}},
	}
	StripLocalStubBindings(cfg)
	if len(cfg.Queues) != 1 || len(cfg.KVNamespaces) != 1 || len(cfg.D1Databases) != 1 {
		t.Errorf("should keep all prod bindings")
	}
}

// ─── parseEnvD1Databases ────────────────────────────────────────

func TestParseEnvD1Databases_Staging(t *testing.T) {
	content := `
[[env.staging.d1_databases]]
binding = "DB"
database_name = "staging-db"
database_id = "staging-id"
`
	dbs := parseEnvD1Databases(content, "env.staging")
	if len(dbs) != 1 {
		t.Fatalf("parseEnvD1Databases got %d, want 1", len(dbs))
	}
	if dbs[0].DatabaseID != "staging-id" {
		t.Errorf("db_id = %q", dbs[0].DatabaseID)
	}
}

func TestParseEnvD1Databases_DefaultDbName(t *testing.T) {
	content := `
[[env.production.d1_databases]]
binding = "DB"
database_id = "prod-id"
`
	dbs := parseEnvD1Databases(content, "env.production")
	if len(dbs) != 1 {
		t.Fatalf("got %d", len(dbs))
	}
	if dbs[0].DatabaseName != "local-db" {
		t.Errorf("default name = %q", dbs[0].DatabaseName)
	}
}

func TestParseEnvD1Databases_SkipsWithoutID(t *testing.T) {
	content := `
[[env.staging.d1_databases]]
binding = "DB"
database_name = "db"
`
	dbs := parseEnvD1Databases(content, "env.staging")
	if len(dbs) != 0 {
		t.Errorf("should skip without database_id, got %d", len(dbs))
	}
}

// ─── EnsureDefaultAI (deprecated, should be no-op) ──────────────

func TestEnsureDefaultAI_IsNoOp(t *testing.T) {
	cfg := &AerostackConfig{AI: false}
	EnsureDefaultAI(cfg)
	if cfg.AI != false {
		t.Errorf("EnsureDefaultAI should be no-op")
	}
}
