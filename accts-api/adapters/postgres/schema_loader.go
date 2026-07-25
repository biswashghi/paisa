package postgres

import (
	"fmt"
	"os"
)

var schemaPathCandidates = []string{
	"db/schema.sql",
	"../db/schema.sql",
	"../../db/schema.sql",
	"../../../db/schema.sql",
}

func loadSchemaSQL() (string, error) {
	if path := os.Getenv("PAISA_SCHEMA_PATH"); path != "" {
		return readSchema(path)
	}

	for _, path := range schemaPathCandidates {
		if data, err := os.ReadFile(path); err == nil {
			return string(data), nil
		}
	}

	return "", fmt.Errorf("could not find shared schema file; set PAISA_SCHEMA_PATH or run from one of %v", schemaPathCandidates)
}

func readSchema(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("could not read schema file %q: %w", path, err)
	}
	return string(data), nil
}
