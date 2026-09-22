package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveTiDBDSN(t *testing.T) {
	t.Run("command line", func(t *testing.T) {
		got, err := resolveTiDBDSN("root@tcp(localhost:4000)/", "")
		if err != nil || got != "root@tcp(localhost:4000)/" {
			t.Fatalf("resolveTiDBDSN() = %q, %v", got, err)
		}
	})

	t.Run("TOML file", func(t *testing.T) {
		path := writeConfig(t, "[tidb]\ndsn = 'root@tcp(127.0.0.1:4000)/'\n")
		got, err := resolveTiDBDSN("", path)
		if err != nil || got != "root@tcp(127.0.0.1:4000)/" {
			t.Fatalf("resolveTiDBDSN() = %q, %v", got, err)
		}
	})

	for _, tc := range []struct {
		name, dsn, contents, wantError string
	}{
		{name: "missing source"},
		{name: "both sources", dsn: "root@tcp(localhost:4000)/", contents: "[tidb]\ndsn = 'root@tcp(127.0.0.1:4000)/'\n", wantError: "cannot be used together"},
		{name: "missing DSN", contents: "[tidb]\n", wantError: "non-empty [tidb] dsn"},
		{name: "unknown key", contents: "[tidb]\naddress = '127.0.0.1:4000'\n", wantError: "unknown config key"},
		{name: "malformed TOML", contents: "[tidb\n", wantError: "read config"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := ""
			if tc.contents != "" {
				path = writeConfig(t, tc.contents)
			}
			got, err := resolveTiDBDSN(tc.dsn, path)
			if tc.wantError == "" {
				if err != nil || got != "" {
					t.Fatalf("resolveTiDBDSN() = %q, %v, want empty DSN without error", got, err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("resolveTiDBDSN() error = %v, want %q", err, tc.wantError)
			}
		})
	}
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
