package library

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"athenaeum/internal/brand"
)

const maxCoverDownload = 8 << 20

// FetchCoverImage downloads a cover image from an HTTP(S) URL.
func FetchCoverImage(ctx context.Context, rawURL string) ([]byte, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, fmt.Errorf("empty cover url")
	}
	u, err := ValidateOutboundURL(rawURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", brand.UserAgent(""))

	res, err := coverHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cover download failed: %s", res.Status)
	}

	ct := strings.ToLower(res.Header.Get("Content-Type"))
	if ct != "" && !strings.HasPrefix(ct, "image/") {
		return nil, fmt.Errorf("cover url is not an image")
	}

	data, err := io.ReadAll(io.LimitReader(res.Body, maxCoverDownload+1))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty cover image")
	}
	if len(data) > maxCoverDownload {
		return nil, fmt.Errorf("cover image too large")
	}
	return data, nil
}
