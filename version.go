package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
		return nil, fmt.Errorf("invalid version format: %s", s)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, err
	}

	return &Version{Major: major, Minor: minor, Patch: patch}, nil
}

func (v *Version) IsNewer(other *Version) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch > other.Patch
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func GetQtCreatorVersion(qtcreatorPath string) (*Version, error) {
	cmd := exec.Command(qtcreatorPath, "-version")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get Qt Creator version: %w", err)
	}

	re := regexp.MustCompile(`Qt Creator (\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not parse Qt Creator version from: %s", output)
	}

	return ParseVersion(matches[1])
}

func GetInstalledPluginVersion(qtcreatorPath, pluginName string) (*Version, error) {
	cmd := exec.Command(qtcreatorPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run Qt Creator --version: %w", err)
	}

	outputStr := string(output)
	pluginNameLower := strings.ToLower(pluginName)
	lines := strings.Split(outputStr, "\n")
	
	for _, line := range lines {
		lineLower := strings.ToLower(line)
		if strings.HasPrefix(lineLower, pluginNameLower+" ") {
			versionPart := strings.TrimSpace(strings.TrimPrefix(lineLower, pluginNameLower+" "))
			parts := strings.Fields(versionPart)
			if len(parts) > 0 {
				if version, err := ParseVersion(parts[0]); err == nil {
					return version, nil
				}
			}
		}
	}

	pattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(pluginName) + `\s+(\d+\.\d+\.\d+)`)
	matches := pattern.FindStringSubmatch(outputStr)
	if len(matches) >= 2 {
		if version, err := ParseVersion(matches[1]); err == nil {
			return version, nil
		}
	}

	return nil, fmt.Errorf("plugin '%s' not found in Qt Creator plugin list", pluginName)
}

func GetInstalledPluginVersionFromPath(pluginPath, pluginName string) (*Version, error) {
	info, err := os.Stat(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("plugin path does not exist: %w", err)
	}

	if !info.IsDir() {
		pluginPath = filepath.Dir(pluginPath)
	}

	entries, err := os.ReadDir(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin directory: %w", err)
	}

	versionPattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(pluginName) + `.*?(\d+\.\d+\.\d+)`)
	pluginNameLower := strings.ToLower(pluginName)
	
	for _, entry := range entries {
		name := entry.Name()
		if strings.Contains(strings.ToLower(name), pluginNameLower) {
			matches := versionPattern.FindStringSubmatch(name)
			if len(matches) >= 2 {
				if version, err := ParseVersion(matches[1]); err == nil {
					return version, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("could not find version in plugin directory")
}

type GithubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func GetLatestGithubRelease(repo string) (*GithubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var release GithubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, err
	}

	return &release, nil
}
