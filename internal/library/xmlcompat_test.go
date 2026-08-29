package library

import (
	"encoding/xml"
	"testing"
)

func TestNormalizeXMLDecl(t *testing.T) {
	input := []byte(`<?xml version="1.1" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`)

	normalized := normalizeXMLDecl(input)
	var c container
	if err := xml.Unmarshal(normalized, &c); err != nil {
		t.Fatalf("unmarshal normalized xml: %v", err)
	}
	if len(c.Rootfiles) != 1 || c.Rootfiles[0].FullPath != "OEBPS/content.opf" {
		t.Fatalf("unexpected container: %+v", c)
	}
}

func TestNormalizeXMLDeclNoOp(t *testing.T) {
	input := []byte(`<?xml version="1.0"?><root/>`)
	if got := normalizeXMLDecl(input); string(got) != string(input) {
		t.Fatalf("expected unchanged input, got %q", got)
	}
}
