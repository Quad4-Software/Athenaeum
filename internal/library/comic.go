package library

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nwaples/rardecode/v2"
)

var imageExt = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
	".gif":  "image/gif",
	".avif": "image/avif",
}

type comicMeta struct {
	PageCount int
	CoverData []byte
	Pages     []comicPageEntry
}

type comicPageEntry struct {
	Name string
	Mime string
}

type comicMetaCacheEntry struct {
	modTime time.Time
	size    int64
	meta    comicMeta
}

type comicPageCacheKey struct {
	path  string
	index int
	size  int64
	mtime int64
}

type comicPageCacheEntry struct {
	key  comicPageCacheKey
	data []byte
	mime string
}

const (
	maxComicMetaEntries = 64
	maxComicPageEntries = 48
)

var (
	comicMetaMu     sync.RWMutex
	comicMetaByPath = make(map[string]comicMetaCacheEntry)
	comicMetaOrder  []string

	comicPageMu    sync.Mutex
	comicPageCache []comicPageCacheEntry
)

func getComicMeta(path string) comicMeta {
	modTime, size, err := comicFileStat(path)
	if err != nil {
		return comicMeta{}
	}

	comicMetaMu.RLock()
	if entry, ok := comicMetaByPath[path]; ok && entry.modTime.Equal(modTime) && entry.size == size {
		meta := entry.meta
		comicMetaMu.RUnlock()
		return meta
	}
	comicMetaMu.RUnlock()

	meta := parseComic(path)

	comicMetaMu.Lock()
	comicMetaByPath[path] = comicMetaCacheEntry{modTime: modTime, size: size, meta: meta}
	touchComicMetaOrder(path)
	evictComicMetaLocked()
	comicMetaMu.Unlock()
	return meta
}

func touchComicMetaOrder(path string) {
	for i, p := range comicMetaOrder {
		if p == path {
			comicMetaOrder = append(comicMetaOrder[:i], comicMetaOrder[i+1:]...)
			break
		}
	}
	comicMetaOrder = append(comicMetaOrder, path)
}

func evictComicMetaLocked() {
	for len(comicMetaOrder) > maxComicMetaEntries {
		oldest := comicMetaOrder[0]
		comicMetaOrder = comicMetaOrder[1:]
		delete(comicMetaByPath, oldest)
	}
}

func comicFileStat(path string) (time.Time, int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return time.Time{}, 0, err
	}
	return fi.ModTime(), fi.Size(), nil
}

func parseComic(path string) comicMeta {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".cbz":
		return parseCBZ(path)
	case ".cbr":
		return parseCBR(path)
	default:
		return comicMeta{}
	}
}

func parseCBZ(path string) comicMeta {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return comicMeta{}
	}
	defer zr.Close()
	var pages []comicPageEntry
	for _, f := range zr.File {
		if f.FileInfo().IsDir() || strings.HasPrefix(filepath.Base(f.Name), ".") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name))
		mime, ok := imageExt[ext]
		if !ok {
			continue
		}
		pages = append(pages, comicPageEntry{Name: f.Name, Mime: mime})
	}
	sortComicPages(pages)
	meta := comicMeta{Pages: pages, PageCount: len(pages)}
	if len(pages) > 0 {
		meta.CoverData = readZipImage(zr, pages[0].Name)
	}
	return meta
}

func readZipImage(zr *zip.ReadCloser, name string) []byte {
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil
		}
		data, err := io.ReadAll(io.LimitReader(rc, 8<<20))
		_ = rc.Close()
		if err != nil {
			return nil
		}
		return data
	}
	return nil
}

func parseCBR(path string) comicMeta {
	f, err := os.Open(path) // #nosec G304
	if err != nil {
		return comicMeta{}
	}
	defer f.Close()
	rr, err := rardecode.NewReader(f)
	if err != nil {
		return comicMeta{}
	}
	var pages []comicPageEntry
	var cover []byte
	for {
		hdr, err := rr.Next()
		if err == io.EOF {
			break
		}
		if err != nil || hdr.IsDir {
			continue
		}
		name := hdr.Name
		if strings.HasPrefix(filepath.Base(name), ".") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		mime, ok := imageExt[ext]
		if !ok {
			continue
		}
		pages = append(pages, comicPageEntry{Name: name, Mime: mime})
		if cover == nil {
			cover, _ = io.ReadAll(io.LimitReader(rr, 8<<20))
		}
	}
	sortComicPages(pages)
	return comicMeta{Pages: pages, PageCount: len(pages), CoverData: cover}
}

func sortComicPages(pages []comicPageEntry) {
	sort.Slice(pages, func(i, j int) bool {
		return naturalLess(pages[i].Name, pages[j].Name)
	})
}

var numChunk = regexp.MustCompile(`(\d+)`)

func naturalLess(a, b string) bool {
	aa := numChunk.FindAllString(a, -1)
	bb := numChunk.FindAllString(b, -1)
	for i := 0; i < len(aa) && i < len(bb); i++ {
		if aa[i] != bb[i] {
			return aa[i] < bb[i]
		}
	}
	return strings.ToLower(a) < strings.ToLower(b)
}

// OpenComicPage returns the image bytes for a page index.
func OpenComicPage(path string, index int) ([]byte, string, error) {
	modTime, size, err := comicFileStat(path)
	if err != nil {
		return nil, "", err
	}
	cacheKey := comicPageCacheKey{
		path:  path,
		index: index,
		size:  size,
		mtime: modTime.UnixNano(),
	}
	if data, mime, ok := getComicPageCached(cacheKey); ok {
		return data, mime, nil
	}

	meta := getComicMeta(path)
	if index < 0 || index >= len(meta.Pages) {
		return nil, "", fmt.Errorf("page out of range")
	}
	entry := meta.Pages[index]
	var data []byte
	var mime string
	switch strings.ToLower(filepath.Ext(path)) {
	case ".cbz":
		data, mime, err = openCBZPage(path, entry.Name)
	case ".cbr":
		data, mime, err = openCBRPage(path, entry.Name)
	default:
		return nil, "", fmt.Errorf("unsupported comic format")
	}
	if err != nil {
		return nil, "", err
	}
	putComicPageCached(cacheKey, data, mime)
	return data, mime, nil
}

func getComicPageCached(key comicPageCacheKey) ([]byte, string, bool) {
	comicPageMu.Lock()
	defer comicPageMu.Unlock()
	for i, entry := range comicPageCache {
		if entry.key != key {
			continue
		}
		if i < len(comicPageCache)-1 {
			comicPageCache = append(append(comicPageCache[:i], comicPageCache[i+1:]...), entry)
		}
		return entry.data, entry.mime, true
	}
	return nil, "", false
}

func putComicPageCached(key comicPageCacheKey, data []byte, mime string) {
	comicPageMu.Lock()
	defer comicPageMu.Unlock()
	for i, entry := range comicPageCache {
		if entry.key == key {
			comicPageCache[i] = comicPageCacheEntry{key: key, data: data, mime: mime}
			if i < len(comicPageCache)-1 {
				e := comicPageCache[i]
				comicPageCache = append(append(comicPageCache[:i], comicPageCache[i+1:]...), e)
			}
			return
		}
	}
	comicPageCache = append(comicPageCache, comicPageCacheEntry{key: key, data: data, mime: mime})
	if len(comicPageCache) > maxComicPageEntries {
		comicPageCache = comicPageCache[len(comicPageCache)-maxComicPageEntries:]
	}
}

func openCBZPage(path, name string) ([]byte, string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, "", err
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, "", err
		}
		data, err := io.ReadAll(io.LimitReader(rc, 32<<20))
		_ = rc.Close()
		mime := imageExt[strings.ToLower(filepath.Ext(name))]
		return data, mime, err
	}
	return nil, "", fmt.Errorf("page not found")
}

func openCBRPage(path, name string) ([]byte, string, error) {
	f, err := os.Open(path) // #nosec G304
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	rr, err := rardecode.NewReader(f)
	if err != nil {
		return nil, "", err
	}
	for {
		hdr, err := rr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", err
		}
		if hdr.Name != name {
			continue
		}
		data, err := io.ReadAll(io.LimitReader(rr, 32<<20))
		mime := imageExt[strings.ToLower(filepath.Ext(name))]
		return data, mime, err
	}
	return nil, "", fmt.Errorf("page not found")
}

// ComicManifest builds API-ready page metadata.
func ComicManifest(path string) (int, []comicPageEntry) {
	meta := getComicMeta(path)
	return meta.PageCount, meta.Pages
}
