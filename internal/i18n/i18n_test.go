package i18n

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFileNested(t *testing.T) {
	raw := []byte(`{
		"$name": "Deutsch",
		"nav": { "allBooks": "Alle Bücher" },
		"app.title": "Athenaeum"
	}`)
	msgs, name, err := ParseFile(raw)
	if err != nil {
		t.Fatal(err)
	}
	if name != "Deutsch" {
		t.Fatalf("name=%q", name)
	}
	if msgs["nav.allBooks"] != "Alle Bücher" {
		t.Fatalf("nav.allBooks=%q", msgs["nav.allBooks"])
	}
	if msgs["app.title"] != "Athenaeum" {
		t.Fatalf("app.title=%q", msgs["app.title"])
	}
}

func TestParseFileRejectsInvalid(t *testing.T) {
	cases := []string{
		`not json`,
		`["array"]`,
		`{"key": 1}`,
		`{"key": null}`,
	}
	for _, raw := range cases {
		if _, _, err := ParseFile([]byte(raw)); err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
}

func TestDefaultMessages(t *testing.T) {
	msgs, err := DefaultMessages()
	if err != nil {
		t.Fatal(err)
	}
	if msgs["nav.allBooks"] == "" {
		t.Fatal("missing nav.allBooks")
	}
}

func TestDefaultTemplate(t *testing.T) {
	tmpl, err := DefaultTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if len(tmpl) == 0 {
		t.Fatal("empty template")
	}
	for k, v := range tmpl {
		if v != "" {
			t.Fatalf("template value for %q should be empty", k)
		}
	}
}

func TestLoaderCatalogSkipsInvalid(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "de.json"), []byte(`{"$name":"Deutsch","nav":{"allBooks":"Alle"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte(`{"x":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte(`{`), 0o644); err != nil {
		t.Fatal(err)
	}

	cat, err := NewLoader(dir).Catalog()
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Locales) != 2 {
		t.Fatalf("locales=%d want 2 (en+de)", len(cat.Locales))
	}
}

func TestLoaderLoadCustom(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "fr.json"), []byte(`{"greet":"Bonjour"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	msgs, err := NewLoader(dir).Load("fr")
	if err != nil {
		t.Fatal(err)
	}
	if msgs["greet"] != "Bonjour" {
		t.Fatalf("greet=%q", msgs["greet"])
	}
}

func TestLoaderLoadMissing(t *testing.T) {
	dir := t.TempDir()
	if _, err := NewLoader(dir).Load("xx"); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidLocaleCode(t *testing.T) {
	if !validLocaleCode("de") || !validLocaleCode("pt-BR") {
		t.Fatal("expected valid codes")
	}
	if validLocaleCode("") || validLocaleCode("../x") || validLocaleCode("bad code") {
		t.Fatal("expected invalid codes")
	}
}
