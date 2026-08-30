package library

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FSDirEntry is one directory in a browse listing.
type FSDirEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// FSBrowseResult lists browsable directories.
type FSBrowseResult struct {
	Path    string       `json:"path"`
	Parent  string       `json:"parent,omitempty"`
	Entries []FSDirEntry `json:"entries"`
}

// BrowseDirs lists subdirectories under path, constrained to allowed roots.
func BrowseDirs(roots []string, path string) (FSBrowseResult, error) {
	roots = normalizeRoots(roots)
	if len(roots) == 0 {
		return FSBrowseResult{}, fmt.Errorf("no browse roots configured")
	}

	path = strings.TrimSpace(path)
	if path == "" {
		entries := make([]FSDirEntry, 0, len(roots))
		for _, root := range roots {
			entries = append(entries, FSDirEntry{Name: filepath.Base(root), Path: root})
		}
		sort.Slice(entries, func(i, j int) bool {
			return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
		})
		return FSBrowseResult{Entries: entries}, nil
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return FSBrowseResult{}, err
	}
	abs = filepath.Clean(abs)

	st, err := os.Stat(abs)
	if err != nil {
		return FSBrowseResult{}, err
	}
	if !st.IsDir() {
		return FSBrowseResult{}, fmt.Errorf("not a directory")
	}
	if !pathAllowed(abs, roots) {
		return FSBrowseResult{}, fmt.Errorf("path not allowed")
	}

	entries, err := listSubdirs(abs)
	if err != nil {
		return FSBrowseResult{}, err
	}

	parent := ""
	if p := filepath.Dir(abs); pathAllowed(p, roots) && p != abs {
		parent = p
	}

	return FSBrowseResult{Path: abs, Parent: parent, Entries: entries}, nil
}

func normalizeRoots(roots []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, root := range roots {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		abs, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		abs = filepath.Clean(abs)
		if _, ok := seen[abs]; ok {
			continue
		}
		if st, err := os.Stat(abs); err != nil || !st.IsDir() {
			continue
		}
		seen[abs] = struct{}{}
		out = append(out, abs)
	}
	sort.Strings(out)
	return out
}

func pathAllowed(path string, roots []string) bool {
	path = filepath.Clean(path)
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = filepath.Clean(resolved)
	}
	for _, root := range roots {
		root = filepath.Clean(root)
		if resolved, err := filepath.EvalSymlinks(root); err == nil {
			root = filepath.Clean(resolved)
		}
		if path == root {
			return true
		}
		if strings.HasPrefix(path, root+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func listSubdirs(dir string) ([]FSDirEntry, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []FSDirEntry
	for _, ent := range ents {
		if !ent.IsDir() {
			continue
		}
		name := ent.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		out = append(out, FSDirEntry{Name: name, Path: filepath.Join(dir, name)})
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}
