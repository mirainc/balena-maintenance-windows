package lockfile

import (
	"os"
)

type Lockfile struct {
	filename string
	fd       *os.File
}

func (l *Lockfile) Lock() error {
	if l.fd == nil {
		fd, err := os.OpenFile(l.filename, os.O_EXCL|os.O_CREATE, 0444)
		if err != nil {
			return err
		}
		l.fd = fd
	}
	return nil
}

func (l *Lockfile) Unlock() error {
	if l.fd != nil {
		if err := l.fd.Close(); err != nil {
			return err
		}
		if err := os.Remove(l.filename); err != nil {
			return err
		}
		l.fd = nil
	}
	return nil
}

func New(filename string) *Lockfile {
	return &Lockfile{filename, nil}
}
