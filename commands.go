package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func installPlugin(qtcVersion *Version, pluginDir, version, checksum string, autoYes bool) error {
	PrintHeader("Install QodeAssist Plugin")
	PrintFieldColored("Qt Creator", qtcVersion.String(), "blue")
	PrintField("Plugin directory", Gray(pluginDir))

	release, err := fetchRelease(version)
	if err != nil {
		return err
	}

	return performInstallation(release, qtcVersion, pluginDir, checksum, false, autoYes)
}

func updatePlugin(qtcVersion *Version, pluginDir, version, checksum string, autoYes bool) error {
	PrintHeader("Update QodeAssist Plugin")
	PrintFieldColored("Qt Creator", qtcVersion.String(), "blue")
	PrintField("Plugin directory", Gray(pluginDir))

	release, err := fetchRelease(version)
	if err != nil {
		return err
	}

	return performInstallation(release, qtcVersion, pluginDir, checksum, true, autoYes)
}

func removePlugin(pluginDir string, autoYes bool) error {
	PrintHeader("Remove QodeAssist Plugin")
	PrintField("Plugin directory", Gray(pluginDir))

	if err := RemovePlugin(pluginDir, autoYes); err != nil {
		return fmt.Errorf("failed to remove plugin: %w", err)
	}

	return nil
}

func fetchRelease(version string) (*GithubRelease, error) {
	fmt.Println()
	if version != "" {
		PrintStep(fmt.Sprintf("Fetching release %s...", version))
		return getAndPrintSpecificRelease(version)
	}

	PrintStep("Fetching latest release...")
	release, _, err := getAndPrintLatestRelease()
	return release, err
}

func getAndPrintLatestRelease() (*GithubRelease, *Version, error) {
	release, err := GetLatestGithubRelease(GithubRepo)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get latest release: %w", err)
	}

	latestVersion, err := ParseVersion(release.TagName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse release version: %w", err)
	}

	PrintFieldColored("Latest version", latestVersion.String(), "green")
	return release, latestVersion, nil
}

func getAndPrintSpecificRelease(version string) (*GithubRelease, error) {
	release, err := GetGithubReleaseByTag(GithubRepo, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get release %s: %w", version, err)
	}

	releaseVersion, err := ParseVersion(release.TagName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse release version: %w", err)
	}

	PrintFieldColored("Plugin version", releaseVersion.String(), "green")
	return release, nil
}

func listVersions() error {
	PrintHeader("Available QodeAssist Plugin Versions")

	minVersion, err := ParseVersion("0.5.9")
	if err != nil {
		return fmt.Errorf("failed to parse minimum version: %w", err)
	}

	PrintField("Minimum version", minVersion.String())
	fmt.Println()

	PrintStep("Fetching releases from GitHub...")
	releases, err := GetAllGithubReleases(GithubRepo)
	if err != nil {
		return fmt.Errorf("failed to fetch releases: %w", err)
	}

	type ReleaseInfo struct {
		Version     *Version
		QtCVersions []string
	}

	var validReleases []ReleaseInfo
	for _, release := range releases {
		if release.TagName == "" {
			continue
		}

		version, err := ParseVersion(release.TagName)
		if err != nil {
			continue
		}

		if version.IsGreaterOrEqual(minVersion) {
			qtcVersions := ExtractQtCreatorVersions(&release)
			validReleases = append(validReleases, ReleaseInfo{
				Version:     version,
				QtCVersions: qtcVersions,
			})
		}
	}

	if len(validReleases) == 0 {
		PrintWarning(fmt.Sprintf("No versions found >= %s", minVersion.String()))
		return nil
	}

	PrintStatus("success", fmt.Sprintf("Found %d version(s)", len(validReleases)))
	fmt.Println()

	for _, info := range validReleases {
		qtcVersionsStr := "N/A"
		if len(info.QtCVersions) > 0 {
			qtcVersionsStr = strings.Join(info.QtCVersions, ", ")
		}
		text := fmt.Sprintf("%s %s %s",
			Cyan(fmt.Sprintf("%-6s", info.Version.String())),
			Gray("→ Qt Creator:"),
			Yellow(qtcVersionsStr))
		PrintColoredListItem(Green("•"), text)
	}

	return nil
}

func performInstallation(release *GithubRelease, qtcVersion *Version, pluginDir string, checksum string, isUpdate bool, autoYes bool) error {
	assetName, assetURL, err := FindPluginAsset(release, qtcVersion)
	if err != nil {
		return fmt.Errorf("failed to find plugin asset: %w", err)
	}

	tmpFile := filepath.Join(os.TempDir(), assetName)

	fmt.Println()
	PrintStep("Downloading plugin...")
	PrintField("Asset", assetName)

	if err := DownloadPlugin(assetURL, tmpFile); err != nil {
		return fmt.Errorf("failed to download plugin: %w", err)
	}
	defer os.Remove(tmpFile)

	if err := VerifyFileChecksum(tmpFile, checksum); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	if isUpdate {
		PrintStep("Checking for old plugin files...")
		if err := RemoveOldQodeAssistFiles(pluginDir, !autoYes); err != nil {
			return fmt.Errorf("failed to remove old plugin files: %w", err)
		}
		fmt.Println()
	}

	PrintStep("Installing plugin...")
	if err := InstallPlugin(tmpFile, pluginDir); err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	PrintStatus("success", "Installation complete!")
	fmt.Println()
	PrintInfo("Please restart Qt Creator to load the new plugin")
	return nil
}
