package library

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"path"
	"strconv"
	"strings"
)

// epubMeta is the metadata extracted from an EPUB's OPF package document.
type epubMeta struct {
	Title       string
	Author      string
	Language    string
	Description string
	Series      string
	SeriesIndex float64
	// CoverData holds the raw bytes of the cover image, if found.
	CoverData []byte
	CoverMIME string
}

// container.xml points to the OPF root file.
type container struct {
	Rootfiles []struct {
		FullPath string `xml:"full-path,attr"`
	} `xml:"rootfiles>rootfile"`
}

// opfPackage is a minimal view of the EPUB OPF document.
type opfPackage struct {
	Metadata struct {
		Titles      []string `xml:"title"`
		Creators    []string `xml:"creator"`
		Language    string   `xml:"language"`
		Description string   `xml:"description"`
		Metas       []struct {
			Name     string `xml:"name,attr"`
			Content  string `xml:"content,attr"`
			Property string `xml:"property,attr"`
			Value    string `xml:",chardata"`
		} `xml:"meta"`
	} `xml:"metadata"`
	Manifest struct {
		Items []struct {
			ID         string `xml:"id,attr"`
			Href       string `xml:"href,attr"`
			MediaType  string `xml:"media-type,attr"`
			Properties string `xml:"properties,attr"`
		} `xml:"item"`
	} `xml:"manifest"`
}

// parseEPUB opens an EPUB file and extracts its metadata and cover image.
func parseEPUB(filePath string) (epubMeta, error) {
	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return epubMeta{}, err
	}
	defer zr.Close()

	files := make(map[string]*zip.File, len(zr.File))
	for _, f := range zr.File {
		files[f.Name] = f
	}

	opfPath, err := opfPathFrom(files)
	if err != nil {
		return epubMeta{}, err
	}

	pkg, err := readOPF(files[opfPath])
	if err != nil {
		return epubMeta{}, err
	}

	meta := epubMeta{Language: pkg.Metadata.Language, Description: pkg.Metadata.Description}
	if len(pkg.Metadata.Titles) > 0 {
		meta.Title = strings.TrimSpace(pkg.Metadata.Titles[0])
	}
	if len(pkg.Metadata.Creators) > 0 {
		meta.Author = strings.TrimSpace(pkg.Metadata.Creators[0])
	}

	var coverID string
	for _, m := range pkg.Metadata.Metas {
		switch {
		case m.Name == "cover":
			coverID = m.Content
		case m.Property == "belongs-to-collection":
			meta.Series = CleanSeriesName(m.Value)
		case m.Property == "group-position" && meta.SeriesIndex == 0:
			meta.SeriesIndex, _ = strconv.ParseFloat(strings.TrimSpace(m.Value), 64)
		case m.Name == "calibre:series":
			meta.Series = CleanSeriesName(m.Content)
		case m.Name == "calibre:series_index":
			meta.SeriesIndex, _ = strconv.ParseFloat(strings.TrimSpace(m.Content), 64)
		}
	}

	coverHref := coverHref(pkg, coverID)
	if coverHref != "" {
		base := path.Dir(opfPath)
		full := path.Clean(path.Join(base, coverHref))
		if cf, ok := files[full]; ok {
			if data, err := readZipFile(cf); err == nil {
				meta.CoverData = data
				meta.CoverMIME = mimeForItem(pkg, coverHref)
			}
		}
	}

	return meta, nil
}

func opfPathFrom(files map[string]*zip.File) (string, error) {
	cf, ok := files["META-INF/container.xml"]
	if !ok {
		return "", errNoContainer
	}
	data, err := readZipFile(cf)
	if err != nil {
		return "", err
	}
	var c container
	if err := xml.Unmarshal(normalizeXMLDecl(data), &c); err != nil {
		return "", err
	}
	if len(c.Rootfiles) == 0 {
		return "", errNoRootfile
	}
	return c.Rootfiles[0].FullPath, nil
}

func readOPF(f *zip.File) (opfPackage, error) {
	if f == nil {
		return opfPackage{}, errNoRootfile
	}
	data, err := readZipFile(f)
	if err != nil {
		return opfPackage{}, err
	}
	var pkg opfPackage
	if err := xml.Unmarshal(normalizeXMLDecl(data), &pkg); err != nil {
		return opfPackage{}, err
	}
	return pkg, nil
}

// coverHref resolves the cover image href either from the EPUB3
// properties="cover-image" hint or the EPUB2 meta cover id.
func coverHref(pkg opfPackage, coverID string) string {
	for _, it := range pkg.Manifest.Items {
		if strings.Contains(it.Properties, "cover-image") {
			return it.Href
		}
	}
	if coverID != "" {
		for _, it := range pkg.Manifest.Items {
			if it.ID == coverID {
				return it.Href
			}
		}
	}
	// Fall back to the first image whose id mentions "cover".
	for _, it := range pkg.Manifest.Items {
		if strings.HasPrefix(it.MediaType, "image/") && strings.Contains(strings.ToLower(it.ID), "cover") {
			return it.Href
		}
	}
	return ""
}

func mimeForItem(pkg opfPackage, href string) string {
	for _, it := range pkg.Manifest.Items {
		if it.Href == href {
			return it.MediaType
		}
	}
	return "image/jpeg"
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}
