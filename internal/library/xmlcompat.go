package library

import (
	"bytes"
	"regexp"
)

var xmlVersion11 = regexp.MustCompile(`(?i)(<\?xml[^>]*version\s*=\s*["'])1\.1(["'])`)

// normalizeXMLDecl rewrites XML 1.1 declarations to 1.0 so encoding/xml can parse them.
func normalizeXMLDecl(data []byte) []byte {
	if !bytes.Contains(data, []byte("1.1")) {
		return data
	}
	return xmlVersion11.ReplaceAll(data, []byte(`${1}1.0${2}`))
}
