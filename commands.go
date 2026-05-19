package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func showStatus(config *Config) error {
	PrintHeader("QodeAssist Plugin Status")

	_, err := getAndPrintQtCreatorVersion(config)
	if err != nil {
		return err
	}

	installedVersion, err := getInstalledVersion(config)
	if err != nil {
		PrintFieldColored("QodeAssist plugin", "Not installed", "red")
	} else {
		PrintFieldColored("QodeAssist plugin", installedVersion.String(), "green")
	}

	PrintField("Qt Creator path", Gray(config.QtCreatorPath))

	pluginPath, err := config.GetPluginPath()
	if err != nil {
		PrintField("Plugin path", Gray(config.PluginPath)+" "+Red("(expansion failed)"))
	} else {
		PrintField("Plugin path", Gray(pluginPath))
	}

	fmt.Println()
	PrintStep("Checking for updates...")
	_, latestVersion, err := getAndPrintLatestRelease()
	if err != nil {
		return err
	}

	if installedVersion == nil {
		PrintInfo("Run with --install to install the plugin")
	} else if latestVersion.IsNewer(installedVersion) {
		PrintStatus("warning", "Update available!")
		PrintField("Current", Red(installedVersion.String()))
		PrintField("Latest", Green(latestVersion.String()))
		fmt.Println()
		PrintInfo("Run with --update to upgrade")
	} else {
		PrintStatus("success", "Plugin is up to date")
	}

	return nil
}

func installPlugin(config *Config, force bool, version string, checksum string) error {
	PrintHeader("Install QodeAssist Plugin")

	qtcVersion, err := getAndPrintQtCreatorVersion(config)
	if err != nil {
		return err
	}

	if !force {
		installedVersion, err := getInstalledVersion(config)
		if err == nil {
			PrintWarning(fmt.Sprintf("Plugin already installed (v%s)", installedVersion))
			PrintInfo("Use --update to upgrade to the latest version")
			return nil
		}
	}

	fmt.Println()
	var release *GithubRelease
	if version != "" {
		PrintStep(fmt.Sprintf("Fetching release %s...", version))
		release, err = getAndPrintSpecificRelease(version)
		if err != nil {
			return err
		}
	} else {
		PrintStep("Fetching latest release...")
		release, _, err = getAndPrintLatestRelease()
		if err != nil {
			return err
		}
	}

	return performInstallation(release, qtcVersion, config, checksum, false, false)
}

func updatePlugin(config *Config, version string, checksum string, autoYes bool) error {
	PrintHeader("Update QodeAssist Plugin")

	qtcVersion, err := getAndPrintQtCreatorVersion(config)
	if err != nil {
		return err
	}

	installedVersion, err := getInstalledVersion(config)
	if err != nil {
		PrintError("Plugin not installed")
		PrintInfo("Run with --install to install the plugin")
		return nil
	}
	PrintFieldColored("Current version", installedVersion.String(), "cyan")

	fmt.Println()
	var release *GithubRelease
	var targetVersion *Version

	if version != "" {
		PrintStep(fmt.Sprintf("Fetching release %s...", version))
		release, err = getAndPrintSpecificRelease(version)
		if err != nil {
			return err
		}
		targetVersion, err = ParseVersion(release.TagName)
		if err != nil {
			return fmt.Errorf("failed to parse release version: %w", err)
		}
	} else {
		PrintStep("Checking for updates...")
		release, targetVersion, err = getAndPrintLatestRelease()
		if err != nil {
			return err
		}
	}

	if !targetVersion.IsNewer(installedVersion) && version == "" {
		PrintStatus("success", "Already up to date!")
		return nil
	}

	if targetVersion.IsNewer(installedVersion) {
		PrintStatus("info", "Update available")
		PrintField("Current", Red(installedVersion.String()))
		PrintField("Target", Green(targetVersion.String()))
	} else if version != "" {
		PrintStatus("warning", "Installing older or same version")
		PrintField("Current", Yellow(installedVersion.String()))
		PrintField("Target", Cyan(targetVersion.String()))
	}
	fmt.Println()

	return performInstallation(release, qtcVersion, config, checksum, true, autoYes)
}

func removePlugin(config *Config, autoYes bool) error {
	PrintHeader("Remove QodeAssist Plugin")

	installedVersion, err := getInstalledVersion(config)
	if err != nil {
		PrintWarning("Plugin is not installed")
		return nil
	}
	PrintFieldColored("Version", installedVersion.String(), "yellow")

	pluginPath, err := config.GetPluginPath()

	if err != nil {
		return fmt.Errorf("failed to get plugin path: %w", err)
	}

	fmt.Println()
	PrintStep("Removing plugin...")
	PrintField("Location", Gray(pluginPath))

	if err := RemovePlugin(pluginPath, autoYes); err != nil {
		return fmt.Errorf("failed to remove plugin: %w", err)
	}

	PrintStatus("success", "Plugin removed successfully")
	return nil
}

func getInstalledVersion(config *Config) (*Version, error) {
	pluginPath, err := config.GetPluginPath()
	if err != nil {
		return nil, fmt.Errorf("plugin not installed: failed to get plugin path: %w", err)
	}

	if !CheckPluginInstalled(pluginPath, PluginName) {
		return nil, fmt.Errorf("plugin not installed")
	}

	qtcInfo, err := config.GetQtCreatorInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get Qt Creator info: %w", err)
	}

	installedVersion, err := GetInstalledPluginVersion(qtcInfo, PluginName, pluginPath, config.QtCreatorPath)
	if err != nil {
		return nil, fmt.Errorf("plugin not installed: %w", err)
	}

	return installedVersion, nil
}

func getAndPrintQtCreatorVersion(config *Config) (*Version, error) {
	qtcVersion, err := config.GetQtCreatorVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get Qt Creator version: %w", err)
	}
	PrintFieldColored("Qt Creator", qtcVersion.String(), "blue")
	return qtcVersion, nil
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

func performInstallation(release *GithubRelease, qtcVersion *Version, config *Config, checksum string, isUpdate bool, autoYes bool) error {
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

	if checksum != "" {
		PrintStep("Verifying checksum...")
	}
	if err := VerifyFileChecksum(tmpFile, checksum); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	pluginPath, err := config.GetPluginPath()
	if err != nil {
		return fmt.Errorf("failed to get plugin path: %w", err)
	}

	if isUpdate {
		PrintStep("Checking for old plugin files...")
		if err := RemoveOldQodeAssistFiles(pluginPath, !autoYes); err != nil {
			return fmt.Errorf("failed to remove old plugin files: %w", err)
		}
		fmt.Println()
	}

	PrintStep("Installing plugin...")
	if err := InstallPlugin(tmpFile, pluginPath); err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	PrintStatus("success", "Installation complete!")
	fmt.Println()
	PrintInfo("Please restart Qt Creator to load the new plugin")
	return nil
}
