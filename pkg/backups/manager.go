package backups

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
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

	err := json.Unmarshal(bytes, &l)

	return l, err
}

// Filename returns the filename this lock should be stored under.
func (l Lock) Filename() string {
	return fmt.Sprintf("%s.lock", l.Name)
}

// Shift returns a new Lock advanced forward to the next version.
func (l Lock) Shift(next string) Lock {
	return NewLock(l.Name, next, l.Current)
}

// Manager performs versioned backup and restoration of files.
type Manager struct {
	backend   Backend
	versioner Versioner
}

// NewManager returns a new Manager with the given backend and versioner.
func NewManager(backend Backend, versioner Versioner) Manager {
	return Manager{
		backend:   backend,
		versioner: versioner,
	}
}

// Backup stores a versioned backup of `path` in the manager's back end and
// updates its lock to point to this version.
func (m Manager) Backup(path string) error {
	// Get lock
	currentLock, err := m.getCurrentLock(path)
	if err != nil {
		return err
	}

	// Read file bytes
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	bytes, err := ioutil.ReadAll(file)

	// Determine backup name
	backupFilename := fmt.Sprintf("%s_%s.bak", path, m.versioner.GetVersion())

	// Create lock
	var newLock Lock
	if currentLock == nil {
		newLock = NewLock(path, backupFilename, "")
	} else {
		newLock = currentLock.Shift(backupFilename)
	}

	// Upload backup
	if err = m.backend.Store(backupFilename, bytes); err != nil {
		return err
	}

	// Upload lock
	lockBytes, err := json.Marshal(newLock)
	if err != nil {
		return err
	}
	if err = m.backend.Store(newLock.Filename(), lockBytes); err != nil {
		return err
	}

	return nil
}

func (m Manager) getCurrentLock(name string) (*Lock, error) {
	lockBytes, err := m.backend.Read(fmt.Sprintf("%s.lock", name))
	if err != nil {
		return nil, nil
	}

	lock, err := NewLockFromBytes(lockBytes)

	return &lock, err
}

// Restore: TODO
func (m Manager) Restore(name string) error {
	return errors.New("Not implemented")
}
