package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServeDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<p>deck</p>"), 0o600); err != nil {
		t.Fatal(err)
	}

	base, stop, err := serveDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()

	if !strings.HasPrefix(base, "http://127.0.0.1:") {
		t.Errorf("base URL %q is not on loopback", base)
	}
	resp, err := http.Get(base + "/index.html")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK || string(body) != "<p>deck</p>" {
		t.Errorf("GET index.html = %d %q", resp.StatusCode, body)
	}
}

func TestWithQuery(t *testing.T) {
	saved := hdQuery
	t.Cleanup(func() { hdQuery = saved })

	for _, tt := range []struct{ query, url, want string }{
		{"capture=1", "http://h/index.html", "http://h/index.html?capture=1"},
		{"capture=1", "http://h/index.html?x=1", "http://h/index.html?x=1&capture=1"},
		{"", "http://h/index.html", "http://h/index.html"},
	} {
		hdQuery = tt.query
		if got := withQuery(tt.url); got != tt.want {
			t.Errorf("withQuery(%q) with query %q = %q, want %q", tt.url, tt.query, got, tt.want)
		}
	}
}
