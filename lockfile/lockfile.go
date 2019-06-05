package lockfile

import (
	"os"
)

type Lockfile struct {
	filename string
}

func (l *Lockfile) Lock() error {
	// Check if file exists
	if _, err := os.Stat(l.filename); err == nil {
		// Lock is already taken
		return nil
	} else if os.IsNotExist(err) {
		// Take the lock by creating a file.
		fd, err := os.OpenFile(l.filename, os.O_EXCL|os.O_CREATE, 0444)
		if err != nil {
			return err
		}
		// Immediately close the file descriptor
		if err := fd.Close(); err != nil {
			return err
		}
	} else {
		// os.IsNotExists may be false in certain conditions even if the file exists,
		// i.e. no permissions. Return an error in this case.
		return err
	}
	return nil
}

func (l *Lockfile) Unlock() error {
	if err := os.RemoveAll(l.filename); err != nil {
		return err
	}
	return nil
}

func New(filename string) *Lockfile {
	return &Lockfile{filename}
}
