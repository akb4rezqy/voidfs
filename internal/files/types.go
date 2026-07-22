package files

import "time"

type Entry struct {
	Name     string
	Path     string
	Size     int64
	IsDir    bool
	ModTime  time.Time
	MimeType string
}
