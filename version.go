package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	qtCreatorVersionRegex = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	pluginLineRegex       = regexp.MustCompile(`^\s+(\S+)\s+(\d+\.\d+\.\d+)\b`)
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func ParseVersion(s string) (*Version, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid version format: %s (expected X.Y.Z)", s)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %w", err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %w", err)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %w", err)
	}

	return &Version{Major: major, Minor: minor, Patch: patch}, nil
}

func (v *Version) IsNewer(other *Version) bool {
	if other == nil {
		return true
	}
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch > other.Patch
}

func (v *Version) IsGreaterOrEqual(other *Version) bool {
	if other == nil {
		return true
	}
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch >= other.Patch
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

type QtCreatorInfo struct {
	Version        *Version
	PluginVersions map[string]*Version
}

func GetQtCreatorInfo(qtCreatorRootPath string) (*QtCreatorInfo, error) {
	execPath, err := GetQtCreatorExecutablePath(qtCreatorRootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find Qt Creator executable: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), CommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, execPath, "--version")
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("qtcreator --version timed out after %v", CommandTimeout)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to run qtcreator --version: %w", err)
	}

	return parseQtCreatorVersionOutput(string(output))
}

func parseQtCreatorVersionOutput(output string) (*QtCreatorInfo, error) {
	info := &QtCreatorInfo{PluginVersions: make(map[string]*Version)}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if rest, ok := strings.CutPrefix(trimmed, "Version:"); ok {
			if match := qtCreatorVersionRegex.FindString(rest); match != "" {
				if version, err := ParseVersion(match); err == nil {
					info.Version = version
				}
			}
			continue
		}

		if m := pluginLineRegex.FindStringSubmatch(line); m != nil {
			if version, err := ParseVersion(m[2]); err == nil {
				info.PluginVersions[strings.ToLower(m[1])] = version
			}
		}
	}

	if info.Version == nil {
		return nil, fmt.Errorf("could not parse Qt Creator version from output: %q", output)
	}

	return info, nil
}

func GetQtCreatorVersion(qtCreatorRootPath string) (*Version, error) {
	info, err := GetQtCreatorInfo(qtCreatorRootPath)
	if err != nil {
		return nil, err
	}
	return info.Version, nil
}

func findPluginFile(pluginPath, pluginName string) string {
	info, err := os.Stat(pluginPath)
	if err != nil {
		return ""
	}

	var searchDir string
	if !info.IsDir() {
		searchDir = filepath.Dir(pluginPath)
	} else {
		searchDir = pluginPath
	}

	pluginNameLower := strings.ToLower(pluginName)
	var pluginFile string

	platformConfig, _ := GetPlatformConfig()
	if platformConfig == nil {
		return ""
	}

	err = filepath.WalkDir(searchDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		fileName := strings.ToLower(d.Name())
		if strings.Contains(fileName, pluginNameLower) {
			ext := filepath.Ext(fileName)
			if ext == platformConfig.LibExtension {
				pluginFile = path
				return filepath.SkipAll
			}
		}
		return nil
	})

	return pluginFile
}

func CheckPluginInstalled(pluginPath, pluginName string) bool {
	return findPluginFile(pluginPath, pluginName) != ""
}

func GetInstalledPluginVersion(qtCreatorInfo *QtCreatorInfo, pluginName string) (*Version, error) {
	version, ok := qtCreatorInfo.PluginVersions[strings.ToLower(pluginName)]
	if !ok {
		return nil, fmt.Errorf("plugin %q not reported by Qt Creator", pluginName)
	}
	return version, nil
}
