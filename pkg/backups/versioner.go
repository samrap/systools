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
}

// NewTimestampVersioner returns a new TimestampVersioner for the given time.
func NewTimestampVersioner() TimestampVersioner {
	return TimestampVersioner{}
}

// GetVersion returns a filename-safe version string in the format `YmdTHMS`.
//
// For example, if the versioner was created on August 15, 2019 at 15:00:00, the
// version string would look like `20190815T150000`.
func (v TimestampVersioner) GetVersion() string {
	return time.Now().Format("20060102T150405")
}

type staticVersioner struct {
	version string
}

func newStaticVersioner(version string) staticVersioner {
	return staticVersioner{version}
}

func (v staticVersioner) GetVersion() string {
	return v.version
}
