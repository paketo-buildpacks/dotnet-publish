package dotnetpublish

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ProjectFileParser struct{}

func NewProjectFileParser() ProjectFileParser {
	return ProjectFileParser{}
}

func (p ProjectFileParser) FindProjectFile(path string) (string, error) {
	projectFiles, err := filepath.Glob(filepath.Join(path, "*.csproj"))
	if err != nil {
		return "", err
	}

	fsProjFiles, err := filepath.Glob(filepath.Join(path, "*.fsproj"))
	if err != nil {
		return "", err
	}
	projectFiles = append(projectFiles, fsProjFiles...)

	vbProjFiles, err := filepath.Glob(filepath.Join(path, "*.vbproj"))
	if err != nil {
		return "", err
	}
	projectFiles = append(projectFiles, vbProjFiles...)

	if len(projectFiles) > 0 {
		return projectFiles[0], nil
	}

	return "", nil
}

func (p ProjectFileParser) ParseVersion(path, rootDir string) (string, error) {
	version, found, err := parseVersionFromFile(path, "project file")
	if err != nil {
		return "", err
	}
	if found {
		return version, nil
	}

	rootDir = filepath.Clean(rootDir)
	for dir := filepath.Clean(filepath.Dir(path)); ; dir = filepath.Dir(dir) {
		propsPath := filepath.Join(dir, "Directory.Build.props")
		_, err = os.Stat(propsPath)
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}

		if err == nil {
			version, found, err = parseVersionFromFile(propsPath, "Directory.Build.props")
			if err != nil {
				return "", err
			}
			if found {
				return version, nil
			}
		}

		if dir == rootDir || filepath.Dir(dir) == dir {
			break
		}
	}

	return "", errors.New("failed to find version in project file: missing or invalid TargetFramework property")
}

func parseVersionFromFile(path, fileDescription string) (string, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", false, fmt.Errorf("failed to read %s: %w", fileDescription, err)
	}
	defer func() {
		_ = file.Close()
	}()

	var project struct {
		PropertyGroups []struct {
			RuntimeFrameworkVersion string
			TargetFramework         string
		} `xml:"PropertyGroup"`
	}

	err = xml.NewDecoder(file).Decode(&project)
	if err != nil {
		return "", false, fmt.Errorf("failed to parse %s: %w", fileDescription, err)
	}

	for _, group := range project.PropertyGroups {
		if group.RuntimeFrameworkVersion != "" {
			return group.RuntimeFrameworkVersion, true, nil
		}
	}

	// This regular expression matches on 'net<x>.<y>',
	// 'net<x>.<y>-<platform>' & 'netcoreapp<x>.<y>'
	targetFrameworkRe := regexp.MustCompile(`net(?:coreapp)?(?:(\d+\.\d)(?:\-?\w+)?)$`)
	for _, group := range project.PropertyGroups {
		matches := targetFrameworkRe.FindStringSubmatch(group.TargetFramework)
		if len(matches) == 2 {
			return fmt.Sprintf("%s.0", matches[1]), true, nil
		}
	}

	return "", false, nil
}

func (p ProjectFileParser) NodeIsRequired(path string) (bool, error) {
	needsNode, err := findInFile("node ", path)
	if err != nil {
		return false, err
	}

	needsNPM, err := findInFile("npm ", path)
	if err != nil {
		return false, err
	}

	return needsNode || needsNPM, nil
}

func (p ProjectFileParser) NPMIsRequired(path string) (bool, error) {
	return findInFile("npm ", path)
}

func findInFile(str, path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	var project struct {
		Targets []struct {
			Execs []struct {
				Command string `xml:",attr"`
			} `xml:"Exec"`
		} `xml:"Target"`
	}

	err = xml.NewDecoder(file).Decode(&project)
	if err != nil {
		return false, fmt.Errorf("failed to decode %s: %w", path, err)
	}

	for _, target := range project.Targets {
		for _, exec := range target.Execs {
			if strings.HasPrefix(exec.Command, str) {
				return true, nil
			}
		}
	}

	return false, nil
}
