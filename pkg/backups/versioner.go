package backups

import (
	"time"
)

// Versioner provides methods for generating version strings for backups.
type Versioner interface {
	GetVersion() string
}

// TimestampVersioner is a Versioner that returns a timestamp that can be used to
// version a file. It takes a `time.Time` which it uses to create the version.
type TimestampVersioner struct {
	time time.Time
}

// NewTimestampVersioner returns a new TimestampVersioner for the given time.
func NewTimestampVersioner(time time.Time) TimestampVersioner {
	return TimestampVersioner{time}
}

// GetVersion returns a filename-safe version string in the format `YmdTHMS`.
//
// For example, if the versioner was created on August 15, 2019 at 15:00:00, the
// version string would look like `20190815T150000`.
func (v TimestampVersioner) GetVersion() string {
	return v.time.Format("YmdTHMS")
}

type staticVersioner struct {
	version string
}

func (v staticVersioner) GetVersion() string {
	return v.version
}
