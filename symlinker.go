package dotnetpublish

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Symlinker struct{}

func NewSymlinker() Symlinker {
	return Symlinker{}
}

func (s Symlinker) Link(oldname, newname string) error {
	err := os.MkdirAll(filepath.Dir(newname), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to make directory for symlink: %w", err)
	}
	return os.Symlink(oldname, newname)
}

func (s Symlinker) Unlink(path string) error {

	fileInfo, err := os.Lstat(path)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	if fileInfo.Mode()&os.ModeSymlink != os.ModeSymlink {
		return fmt.Errorf("cannot unlink %s because it is not a symlink", path)
	}
	return os.Remove(path)
}
