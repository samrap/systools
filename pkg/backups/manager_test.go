package backups

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ItBacksUpBytesAtTheGivenName(t *testing.T) {
	versioner := newStaticVersioner("VERSION")
	backend := NewInMemoryBackend()
	manager := NewManager(backend, versioner)

	contents := []byte("Nothing is certain but death and taxes.")
	err := manager.Backup("truth.txt", bytes.NewReader(contents))

	assert.NoError(t, err)
	assert.Contains(t, backend.Backups, "truth.txt_VERSION.bak")
	assert.Equal(t, contents, backend.Backups["truth.txt_VERSION.bak"])
}

func Test_ItCreatesALockForBackup(t *testing.T) {
	versioner := newStaticVersioner("VERSION")
	backend := NewInMemoryBackend()
	manager := NewManager(backend, versioner)

	contents := []byte("Nothing is certain but death and taxes.")
	err := manager.Backup("truth.txt", bytes.NewReader(contents))

	assert.NoError(t, err)
	assert.Contains(t, backend.Backups, "truth.txt.lock")
}

func TestLockForNewBackup(t *testing.T) {
	versioner := newStaticVersioner("VERSION")
	backend := NewInMemoryBackend()
	manager := NewManager(backend, versioner)

	contents := []byte("Nothing is certain but death and taxes.")
	err := manager.Backup("truth.txt", bytes.NewReader(contents))

	assert.NoError(t, err)

	lock, err := NewLockFromBytes(backend.Backups["truth.txt.lock"])

	assert.NoError(t, err)
	assert.Equal(t, "truth.txt", lock.Name)
	assert.Equal(t, "truth.txt_VERSION.bak", lock.Current)
	assert.Equal(t, "", lock.Previous)
}

func TestLockForExistingBackup(t *testing.T) {
	versioner := newStaticVersioner("VERSION")
	backend := NewInMemoryBackend()
	manager := NewManager(backend, versioner)

	contents := []byte("Nothing is certain but death and taxes.")

	// Perform two backups.
	err := manager.Backup("truth.txt", bytes.NewReader(contents))
	assert.NoError(t, err)
	err = manager.Backup("truth.txt", bytes.NewReader(contents))
	assert.NoError(t, err)

	lock, err := NewLockFromBytes(backend.Backups["truth.txt.lock"])

	assert.NoError(t, err)
	assert.Equal(t, "truth.txt", lock.Name)
	// There should be current and previous versions in the lock.
	assert.Equal(t, "truth.txt_VERSION.bak", lock.Current)
	assert.Equal(t, "truth.txt_VERSION.bak", lock.Previous)
}

func Test_ItRestoresFromTheGivenName(t *testing.T) {
	versioner := newStaticVersioner("VERSION")
	backend := NewInMemoryBackend()
	manager := NewManager(backend, versioner)

	contents := []byte("Nothing is certain but death and taxes.")

	// Create a back up.
	err := manager.Backup("truth.txt", bytes.NewReader(contents))
	assert.NoError(t, err)

	// Restore the back up.
	reader, err := manager.Restore("truth.txt")
	assert.NoError(t, err)
	assert.NotNil(t, reader)

	bytes, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, contents, bytes)
}

func Test_ItAlwaysRestoresTheCurrentVersion(t *testing.T) {
	versioner := newStaticVersioner("VERSION")
	backend := NewInMemoryBackend()
	manager := NewManager(backend, versioner)

	v1Contents := []byte("Nothing is certain but death and taxes.")

	// Create a back up.
	err := manager.Backup("truth.txt", bytes.NewReader(v1Contents))
	assert.NoError(t, err)

	v2Contents := []byte("Nobody exists on purpose. Nobody belongs anywhere. We're all going to die. Come watch TV.")

	// Create a second back up with new contents. We'll need to mutate the
	// versioner so that it stores the back up under a new version...
	manager.versioner = newStaticVersioner("VERSION_2")
	err = manager.Backup("truth.txt", bytes.NewReader(v2Contents))
	assert.NoError(t, err)
	assert.Equal(t, 3, len(backend.Backups))

	reader, err := manager.Restore("truth.txt")
	assert.NoError(t, err)
	assert.NotNil(t, reader)

	bytes, err := ioutil.ReadAll(reader)

	// The restored back up should contain the latest contents.
	assert.NoError(t, err)
	assert.Equal(t, v2Contents, bytes)
}
