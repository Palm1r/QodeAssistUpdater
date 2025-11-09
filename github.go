package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

type GithubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var githubClient = &http.Client{
	Timeout: HTTPTimeout,
}

func GetLatestGithubRelease(repo string) (*GithubRelease, error) {
	if strings.Contains(repo, "/..") || strings.Contains(repo, "\\") {
		return nil, fmt.Errorf("invalid repository name: %s", repo)
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s (expected owner/repo)", repo)
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	var resp *http.Response
	var err error
	maxRetries := 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
		}

		resp, err = githubClient.Get(apiURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if resp != nil {
			resp.Body.Close()
		}

		if attempt == maxRetries {
			if err != nil {
				return nil, fmt.Errorf("failed to fetch release after %d attempts: %w", maxRetries+1, err)
			}
			return nil, fmt.Errorf("github API returned status: %d after %d attempts", resp.StatusCode, maxRetries+1)
		}
	}
	defer resp.Body.Close()

	limitedReader := io.LimitReader(resp.Body, MaxGitHubAPIResponseSize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var release GithubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if release.TagName == "" {
		return nil, fmt.Errorf("release tag name is empty")
	}

	return &release, nil
}

func GetGithubReleaseByTag(repo string, tag string) (*GithubRelease, error) {
	if strings.Contains(repo, "/..") || strings.Contains(repo, "\\") {
		return nil, fmt.Errorf("invalid repository name: %s", repo)
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s (expected owner/repo)", repo)
	}

	// Ensure tag starts with 'v'
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", repo, tag)

	var resp *http.Response
	var err error
	maxRetries := 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
		}

		resp, err = githubClient.Get(apiURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if resp != nil {
			resp.Body.Close()
		}

		if attempt == maxRetries {
			if err != nil {
				return nil, fmt.Errorf("failed to fetch release after %d attempts: %w", maxRetries+1, err)
			}
			if resp.StatusCode == http.StatusNotFound {
				return nil, fmt.Errorf("release with tag %s not found", tag)
			}
			return nil, fmt.Errorf("github API returned status: %d after %d attempts", resp.StatusCode, maxRetries+1)
		}
	}
	defer resp.Body.Close()

	limitedReader := io.LimitReader(resp.Body, MaxGitHubAPIResponseSize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var release GithubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if release.TagName == "" {
		return nil, fmt.Errorf("release tag name is empty")
	}

	return &release, nil
}

func GetAllGithubReleases(repo string) ([]GithubRelease, error) {
	if strings.Contains(repo, "/..") || strings.Contains(repo, "\\") {
		return nil, fmt.Errorf("invalid repository name: %s", repo)
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s (expected owner/repo)", repo)
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases", repo)

	var resp *http.Response
	var err error
	maxRetries := 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
		}

		resp, err = githubClient.Get(apiURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if resp != nil {
			resp.Body.Close()
		}

		if attempt == maxRetries {
			if err != nil {
				return nil, fmt.Errorf("failed to fetch releases after %d attempts: %w", maxRetries+1, err)
			}
			return nil, fmt.Errorf("github API returned status: %d after %d attempts", resp.StatusCode, maxRetries+1)
		}
	}
	defer resp.Body.Close()

	limitedReader := io.LimitReader(resp.Body, MaxGitHubAPIResponseSize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var releases []GithubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return releases, nil
}

func FindPluginAsset(release *GithubRelease, qtCreatorVersion *Version) (string, string, error) {
	platformName, archName, err := GetPlatformArchName()
	if err != nil {
		return "", "", err
	}

	qtcVerFull := fmt.Sprintf("QtC%d.%d.%d", qtCreatorVersion.Major, qtCreatorVersion.Minor, qtCreatorVersion.Patch)
	qtcVerMajorMinor := fmt.Sprintf("QtC%d.%d", qtCreatorVersion.Major, qtCreatorVersion.Minor)

	var matchingAssets []struct {
		Name             string
		URL              string
		FullVersionMatch bool
	}

	for _, asset := range release.Assets {
		name := asset.Name
		hasPlatform := strings.Contains(name, platformName)
		hasArch := strings.Contains(name, archName)
		hasQtcVersionFull := strings.Contains(name, qtcVerFull)
		hasQtcVersionMajorMinor := strings.Contains(name, qtcVerMajorMinor)

		if hasPlatform && hasArch && (hasQtcVersionFull || hasQtcVersionMajorMinor) {
			matchingAssets = append(matchingAssets, struct {
				Name             string
				URL              string
				FullVersionMatch bool
			}{asset.Name, asset.BrowserDownloadURL, hasQtcVersionFull})
		}
	}

	if len(matchingAssets) == 0 {
		return "", "", fmt.Errorf("no matching asset found for %s %s Qt Creator %s",
			platformName, archName, qtCreatorVersion.String())
	}

	for _, asset := range matchingAssets {
		if asset.FullVersionMatch {
			return asset.Name, asset.URL, nil
		}
	}

	return matchingAssets[0].Name, matchingAssets[0].URL, nil
}

func ExtractQtCreatorVersions(release *GithubRelease) []string {
	qtcVersionPattern := regexp.MustCompile(`QtC(\d+\.\d+(?:\.\d+)?)`)
	versionMap := make(map[string]bool)

	for _, asset := range release.Assets {
		matches := qtcVersionPattern.FindStringSubmatch(asset.Name)
		if len(matches) > 1 {
			versionMap[matches[1]] = true
		}
	}

	versions := make([]string, 0, len(versionMap))
	for version := range versionMap {
		versions = append(versions, version)
	}

	sort.Slice(versions, func(i, j int) bool {
		vi, _ := ParseVersion(versions[i])
		vj, _ := ParseVersion(versions[j])
		if vi == nil || vj == nil {
			return versions[i] < versions[j]
		}
		return !vi.IsNewer(vj)
	})

	return versions
}
