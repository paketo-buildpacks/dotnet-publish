package dotnetpublish

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2/fs"
)

type DotnetSourceRemover struct{}

func NewDotnetSourceRemover() DotnetSourceRemover {
	return DotnetSourceRemover{}
}

func (m DotnetSourceRemover) Remove(workingDir, publishOutputDir string, excludedFiles ...string) error {
	workingDirFiles, err := filepath.Glob(filepath.Join(workingDir, "*"))
	if err != nil {
		return fmt.Errorf("could not glob %s: %w", filepath.Join(workingDir, "*"), err)
	}

	for _, file := range workingDirFiles {
		protectFile := false
		for _, excludedFileName := range excludedFiles {
			if filepath.Base(file) != excludedFileName {
				continue
			}
			protectFile = true
		}

		if !protectFile {
			err = os.RemoveAll(file)
			if err != nil {
				return fmt.Errorf("could not remove %s: %w", file, err)
			}
		}
	}

	outputFiles, err := filepath.Glob(filepath.Join(publishOutputDir, "*"))
	if err != nil {
		return fmt.Errorf("could not glob %s: %w", filepath.Join(publishOutputDir, "*"), err)
	}
	for _, file := range outputFiles {
		err = fs.Move(file, filepath.Join(workingDir, filepath.Base(file)))
		if err != nil {
			return err
		}
	}
	return nil
}
