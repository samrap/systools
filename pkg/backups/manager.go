package backups

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

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

// Backup creates and stores a backup for `name` with the contents of `reader`
// in the Manager's backend. Backups are versioned and stored under a name
// in the format "$name_$version.bak", where $name is the name passed
// to this function and $version is calculated from `m.versioner`.
//
// A lockfile is created for each name and points to the latest stored backup
// under that name. The lockfile is used to restore from the latest backup.
func (m Manager) Backup(name string, reader io.Reader) error {
	currentLock, err := m.getCurrentLock(name)
	if err != nil {
		return err
	}

	backupFilename := fmt.Sprintf("%s_%s.bak", name, m.versioner.GetVersion())
	if err = m.backend.Store(backupFilename, reader); err != nil {
		return err
	}

	var newLock Lock
	if currentLock == nil {
		newLock = NewLock(name, backupFilename, "")
	} else {
		newLock = currentLock.Shift(backupFilename)
	}

	lockBytes, err := json.Marshal(newLock)
	if err != nil {
		return err
	}
	if err = m.backend.Store(newLock.ID(), bytes.NewReader(lockBytes)); err != nil {
		return err
	}

	return nil
}

// Restore attempts to restore the latest backup under `name` by looking for an
// associated lock and returning an `io.Reader` for the contents of backup
// that the lock points to. If no lock exists for the given name, this
// function is unable to find a back up and will return an error.
func (m Manager) Restore(name string) (io.Reader, error) {
	lock, err := m.getCurrentLock(name)
	if err != nil {
		return nil, err
	}

	if lock == nil {
		return nil, fmt.Errorf("No backup exists for file %s", name)
	}

	return m.backend.Read(lock.Current)
}

func (m Manager) getCurrentLock(name string) (*Lock, error) {
	lockReader, err := m.backend.Read(fmt.Sprintf("%s.lock", name))
	if err != nil {
		// If the we get an error because the lock does not exist, we'll simply
		// return a nil lock with no error. The caller must determine if it
		// is ok for a lock not to exist for the given name.
		if _, ok := err.(NoSuchName); ok {
			return nil, nil
		}
		return nil, err
	}

	lockBytes, err := ioutil.ReadAll(lockReader)
	if err != nil {
		return nil, err
	}

	lock, err := NewLockFromBytes(lockBytes)

	return &lock, err
}
