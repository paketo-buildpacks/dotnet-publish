package dotnetpublish

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/fs"
)

type DotnetRootManager struct{}

func NewDotnetRootManager() DotnetRootManager {
	return DotnetRootManager{}
}

func (m DotnetRootManager) Setup(root, existingRoot, sdkLocation string) error {
	paths, err := filepath.Glob(filepath.Join(existingRoot, "shared", "*"))
	if err != nil {
		return fmt.Errorf("could not glob %s: %w", filepath.Join(existingRoot, "shared", "*"), err)
	}

	paths = append(paths, filepath.Join(existingRoot, "host"))
	for _, path := range paths {
		relPath, err := filepath.Rel(existingRoot, path)
		if err != nil {
			return err
		}

		newPath := filepath.Join(root, relPath)
		err = os.MkdirAll(filepath.Dir(newPath), os.ModePerm)
		if err != nil {
			return fmt.Errorf("could not write to %s: %w", filepath.Dir(newPath), err)
		}

		err = os.Symlink(path, newPath)
		if err != nil {
			return fmt.Errorf("could not create symlink: %w", err)
		}
	}

	// NOTE: the dotnet CLI uses relative pathing that means we must copy it into
	// the final DOTNET_ROOT so that it can find SDKs.
	err = fs.Copy(filepath.Join(existingRoot, "dotnet"), filepath.Join(root, "dotnet"))
	if err != nil {
		return fmt.Errorf("could not copy the dotnet cli: %w", err)
	}

	err = os.Symlink(filepath.Join(sdkLocation, "sdk"), filepath.Join(root, "sdk"))
	if err != nil {
		return fmt.Errorf("could not create symlink: %w", err)
	}

	return nil
}
