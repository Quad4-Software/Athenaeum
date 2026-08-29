// Package i18n loads and validates locale JSON files for the Athenaeum UI.
package i18n

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
)

//go:embed locales/en.json
var defaultLocaleJSON []byte

const metaNameKey = "$name"

// LocaleInfo describes one available translation bundle.
type LocaleInfo struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Source string `json:"source"`
}

// Catalog lists bundled and custom locale bundles.
type Catalog struct {
	Locales []LocaleInfo `json:"locales"`
}

// Loader reads locale files from a directory on disk.
type Loader struct {
	dir string
}

// NewLoader returns a loader for JSON files in dir. Missing directories are allowed.
func NewLoader(dir string) *Loader {
	return &Loader{dir: dir}
}

// DefaultMessages returns the embedded English message map.
func DefaultMessages() (map[string]string, error) {
	msgs, _, err := ParseFile(defaultLocaleJSON)
	return msgs, err
}

// DefaultTemplate returns a copy of the default keys with empty values for translators.
func DefaultTemplate() (map[string]string, error) {
	msgs, err := DefaultMessages()
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(msgs))
	for k := range msgs {
		out[k] = ""
	}
	return out, nil
}

// Catalog returns bundled English plus any valid custom locale files in the loader directory.
func (l *Loader) Catalog() (Catalog, error) {
	out := Catalog{
		Locales: []LocaleInfo{{Code: "en", Name: "English", Source: "bundled"}},
	}

	if l.dir == "" {
		return out, nil
	}
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return Catalog{}, err
	}

	seen := map[string]struct{}{"en": {}}
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(strings.ToLower(ent.Name()), ".json") {
			continue
		}
		code := strings.TrimSuffix(ent.Name(), filepath.Ext(ent.Name()))
		if code == "" || !validLocaleCode(code) {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		path := filepath.Join(l.dir, ent.Name())
		msgs, name, err := LoadFile(path)
		if err != nil || len(msgs) == 0 {
			continue
		}
		if name == "" {
			name = code
		}
		seen[code] = struct{}{}
		out.Locales = append(out.Locales, LocaleInfo{
			Code:   code,
			Name:   name,
			Source: "custom",
		})
	}
	return out, nil
}

// Load returns messages for a locale code. English is always served from the embedded bundle.
func (l *Loader) Load(code string) (map[string]string, error) {
	if code == "en" {
		return DefaultMessages()
	}
	if l.dir == "" {
		return nil, fs.ErrNotExist
	}
	path := filepath.Join(l.dir, code+".json")
	msgs, _, err := LoadFile(path)
	if err != nil {
		return nil, err
	}
	if len(msgs) == 0 {
		return nil, fmt.Errorf("empty locale %q", code)
	}
	return msgs, nil
}

// LoadFile parses one JSON locale file from disk.
func LoadFile(path string) (map[string]string, string, error) {
	raw, err := os.ReadFile(path) // #nosec G304 -- locale file under configured i18n directory
	if err != nil {
		return nil, "", err
	}
	return ParseFile(raw)
}

// ParseFile validates and flattens a locale JSON document.
func ParseFile(raw []byte) (map[string]string, string, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, "", err
	}
	name := ""
	if meta, ok := root[metaNameKey]; ok {
		_ = json.Unmarshal(meta, &name)
		delete(root, metaNameKey)
	}
	msgs, err := flatten(root, "")
	if err != nil {
		return nil, "", err
	}
	return msgs, name, nil
}

func flatten(node map[string]json.RawMessage, prefix string) (map[string]string, error) {
	out := make(map[string]string)
	for key, raw := range node {
		if strings.HasPrefix(key, "$") {
			continue
		}
		full := key
		if prefix != "" {
			full = prefix + "." + key
		}
		if len(raw) > 0 && raw[0] == '"' {
			var val string
			if err := json.Unmarshal(raw, &val); err != nil {
				return nil, err
			}
			out[full] = val
			continue
		}
		if string(raw) == "null" {
			return nil, fmt.Errorf("invalid value for key %q", full)
		}
		var nested map[string]json.RawMessage
		if err := json.Unmarshal(raw, &nested); err != nil {
			return nil, fmt.Errorf("invalid value for key %q", full)
		}
		child, err := flatten(nested, full)
		if err != nil {
			return nil, err
		}
		maps.Copy(out, child)
	}
	return out, nil
}

func validLocaleCode(code string) bool {
	if code == "" || len(code) > 16 {
		return false
	}
	for i, r := range code {
		if r == '-' || r == '_' {
			if i == 0 {
				return false
			}
			continue
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}
