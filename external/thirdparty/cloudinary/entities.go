package cloudinary

import "time"

type UploadInput struct {
	Bytes       []byte
	Filename    string
	Folder      string
	PublicID    string
	ContentType string
}

type UploadResponse struct {
	PublicID  string
	URL       string
	SecureURL string
	Width     int
	Height    int
	Format    string
	CreatedAt time.Time
	Resource  string
}

type DeleteInput struct {
	PublicID string
}

type DeleteResponse struct {
	Result string
}

type GetURLInput struct {
	PublicID string
}

type GetURLResponse struct {
	URL string
}
