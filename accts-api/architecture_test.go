package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplicationDomainAndPortsDoNotImportPostgresAdapter(t *testing.T) {
	roots := []string{"application", "domain", "ports"}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.Contains(string(data), "accts-api/adapters/postgres") {
				t.Fatalf("%s imports or references the Postgres adapter", path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
