package backups

import (
	"encoding/json"
	"fmt"
	"time"
)

// Lock is a data structure representing a backup's lock file.
//
// A single back up may have many versions. The lock file is used to point to
// the current version to restore from as well as the previous version as
// an easy way to roll back. It also provides flexibility to point to
// an even older version by locking Lock.Current to any version.
type Lock struct {
	Name      string    `json:"name"`
	Current   string    `json:"current"`
	Previous  string    `json:"previous"`
	CreatedAt time.Time `json:"created_at"`
}

// NewLock creates a new lock with a name, current and previous versions.
func NewLock(name string, current string, previous string) Lock {
	return Lock{
		Name:      name,
		Current:   current,
		Previous:  previous,
		CreatedAt: time.Now(),
	}
}

// NewLockFromBytes unmarshals a lock file's JSON bytes into a Lock struct.
func NewLockFromBytes(bytes []byte) (Lock, error) {
	var l Lock

	return l, json.Unmarshal(bytes, &l)
}

// ID returns the ID this lock should be stored under.
func (l Lock) ID() string {
	return fmt.Sprintf("%s.lock", l.Name)
}

// Shift returns a new Lock advanced forward to the next version.
func (l Lock) Shift(next string) Lock {
	return NewLock(l.Name, next, l.Current)
}
