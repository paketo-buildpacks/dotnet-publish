package dotnetpublish

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"strings"
)

type DotnetProjectFileParser struct{}

func NewDotnetProjectFileParser() DotnetProjectFileParser {
	return DotnetProjectFileParser{}
}

func (p DotnetProjectFileParser) ParseVersion(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to read project file: %w", err)
	}
	defer file.Close()

	var project struct {
		PropertyGroups []struct {
			RuntimeFrameworkVersion string
			TargetFramework         string
		} `xml:"PropertyGroup"`
	}

	err = xml.NewDecoder(file).Decode(&project)
	if err != nil {
		return "", fmt.Errorf("failed to parse project file: %w", err)
	}

	for _, group := range project.PropertyGroups {
		if group.RuntimeFrameworkVersion != "" {
			return group.RuntimeFrameworkVersion, nil
		}
	}

	for _, group := range project.PropertyGroups {
		if strings.HasPrefix(group.TargetFramework, "netcoreapp") {
			return fmt.Sprintf("%s.0", strings.TrimPrefix(group.TargetFramework, "netcoreapp")), nil
		}
	}

	return "", errors.New("failed to find version in project file: missing TargetFramework property")
}

func (p DotnetProjectFileParser) ASPNetIsRequired(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer file.Close()

	var project struct {
		SDK string `xml:"Sdk,attr"`
	}

	err = xml.NewDecoder(file).Decode(&project)
	if err != nil {
		return false, fmt.Errorf("failed to decode %s: %w", path, err)
	}

	return project.SDK == "Microsoft.NET.Sdk.Web", nil
}

func (p DotnetProjectFileParser) NodeIsRequired(path string) (bool, error) {
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

func (p DotnetProjectFileParser) NPMIsRequired(path string) (bool, error) {
	return findInFile("npm ", path)
}

func findInFile(str, path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer file.Close()

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
