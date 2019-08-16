package backups

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func executeWithTempfile(t *testing.T, run func(filename string)) {
	file, err := ioutil.TempFile("", "gotest-systools-manager-*")
	assert.NoError(t, err)

	filename := file.Name()
	defer os.Remove(filename)

	_, err = file.Write([]byte("Hello"))
	assert.NoError(t, err)

	err = file.Close()
	assert.NoError(t, err)

	run(filename)
}

func TestItBacksUpAFile(t *testing.T) {
	backend := NewInMemoryBackend()
	versioner := staticVersioner{"VERSION"}
	manager := NewManager(backend, versioner)

	executeWithTempfile(t, func(filename string) {
		var err error

		err = manager.Backup(filename)
		assert.NoError(t, err)

		// A versioned backup should have been stored in the back end.
		_, err = backend.Read(fmt.Sprintf("%s_%s.bak", filename, versioner.GetVersion()))
		assert.NoError(t, err)

		// An associated lock should have been stored in the back end.
		_, err = backend.Read(fmt.Sprintf("%s.lock", filename))
	})
}

func TestLockForNewBackupHasValidInfo(t *testing.T) {
	backend := NewInMemoryBackend()
	versioner := staticVersioner{"VERSION"}
	manager := NewManager(backend, versioner)

	executeWithTempfile(t, func(filename string) {
		var err error

		err = manager.Backup(filename)
		assert.NoError(t, err)

		lockBytes, err := backend.Read(fmt.Sprintf("%s.lock", filename))
		assert.NoError(t, err)

		lock, err := NewLockFromBytes(lockBytes)
		assert.NoError(t, err)
		assert.Equal(t, "", lock.Previous)
		assert.Equal(
			t,
			fmt.Sprintf("%s_%s.bak", filename, versioner.GetVersion()),
			lock.Current,
		)
	})
}

func TestLockForExistingBackupsHasPreviousAndCurrentPointers(t *testing.T) {
	backend := NewInMemoryBackend()
	versioner := staticVersioner{"VERSION"}
	manager := NewManager(backend, versioner)

	executeWithTempfile(t, func(filename string) {
		var err error

		// Create the initial backup.
		err = manager.Backup(filename)
		assert.NoError(t, err)

		// Create another backup for the same file.
		err = manager.Backup(filename)
		assert.NoError(t, err)

		lockBytes, err := backend.Read(fmt.Sprintf("%s.lock", filename))
		assert.NoError(t, err)

		lock, err := NewLockFromBytes(lockBytes)
		assert.NoError(t, err)
		assert.Equal(
			t,
			fmt.Sprintf("%s_%s.bak", filename, versioner.GetVersion()),
			lock.Previous,
		)
		assert.Equal(
			t,
			fmt.Sprintf("%s_%s.bak", filename, versioner.GetVersion()),
			lock.Current,
		)
	})
}
