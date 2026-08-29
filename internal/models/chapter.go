package models

// Chapter marks a start point in an audiobook or other timed media.
type Chapter struct {
	Index    int     `json:"index"`
	Title    string  `json:"title"`
	StartSec float64 `json:"startSec"`
}
