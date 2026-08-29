package models

import "time"

const (
	// LibraryBackendLocal is a host directory mount.
	LibraryBackendLocal = "local"
	// LibraryBackendS3 is a MinIO-compatible object store mount.
	LibraryBackendS3 = "s3"
)

// Library is a mounted filesystem root indexed as a separate catalog.
type Library struct {
	ID        int64            `json:"id"`
	Name      string           `json:"name"`
	MountPath string           `json:"mountPath"`
	Backend   string           `json:"backend"`
	S3        *LibraryS3Public `json:"s3,omitempty"`
	SortOrder int              `json:"sortOrder"`
	BookCount int64            `json:"bookCount"`
	CreatedAt time.Time        `json:"createdAt"`
}

// LibraryS3Public is S3 config returned by the API without the secret key.
type LibraryS3Public struct {
	Endpoint     string `json:"endpoint"`
	Region       string `json:"region"`
	Bucket       string `json:"bucket"`
	Prefix       string `json:"prefix"`
	AccessKey    string `json:"accessKey"`
	UsePathStyle bool   `json:"usePathStyle"`
	TLS          bool   `json:"tls"`
	HasSecretKey bool   `json:"hasSecretKey"`
}

// LibraryS3Input is S3 config accepted on create/update.
type LibraryS3Input struct {
	Endpoint     string `json:"endpoint"`
	Region       string `json:"region"`
	Bucket       string `json:"bucket"`
	Prefix       string `json:"prefix"`
	AccessKey    string `json:"accessKey"`
	SecretKey    string `json:"secretKey"`
	UsePathStyle bool   `json:"usePathStyle"`
	TLS          bool   `json:"tls"`
}

// LibraryCreate is the payload for creating a library mount.
type LibraryCreate struct {
	Name      string          `json:"name"`
	MountPath string          `json:"mountPath"`
	Backend   string          `json:"backend"`
	S3        *LibraryS3Input `json:"s3,omitempty"`
}
