package postgres

import (
	"strings"
	"testing"
)

func TestLoadSchemaSQLUsesSharedDBSchema(t *testing.T) {
	schema, err := loadSchemaSQL()
	if err != nil {
		t.Fatalf("load schema: %v", err)
	}
	if !strings.Contains(schema, "CREATE SCHEMA IF NOT EXISTS paisa") {
		t.Fatal("expected shared schema SQL to create the paisa schema")
	}
	if !strings.Contains(schema, "CREATE TABLE IF NOT EXISTS paisa.partners") {
		t.Fatal("expected shared schema SQL to include core partner table")
	}
}
