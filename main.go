package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("QodeAssist Plugin Updater")
	fmt.Println("=========================")

	qtcVersion, err := GetQtCreatorVersion(config.QtCreatorPath)
	if err != nil {
		log.Fatalf("Failed to get Qt Creator version: %v", err)
	}
	fmt.Printf("Qt Creator version: %s\n", qtcVersion)

	installedVersion, err := GetInstalledPluginVersion(config.QtCreatorPath, PluginName)
	if err != nil {
		fmt.Printf("Trying alternative method to detect plugin version...\n")
		installedVersion, err = GetInstalledPluginVersionFromPath(config.PluginPath, PluginName)
		if err != nil {
			fmt.Printf("Plugin not installed or error: %v\n", err)
			installedVersion = &Version{0, 0, 0}
		} else {
			fmt.Printf("Installed plugin version (from files): %s\n", installedVersion)
		}
	} else {
		fmt.Printf("Installed plugin version: %s\n", installedVersion)
	}

	fmt.Printf("Checking latest release from %s...\n", GithubRepo)
	release, err := GetLatestGithubRelease(GithubRepo)
	if err != nil {
		log.Fatalf("Failed to get latest release: %v", err)
	}

	latestVersion, err := ParseVersion(release.TagName)
	if err != nil {
		log.Fatalf("Failed to parse release version: %v", err)
	}
	fmt.Printf("Latest plugin version: %s\n", latestVersion)

	if !latestVersion.IsNewer(installedVersion) {
		fmt.Println("Plugin is up to date!")
		return
	}

	fmt.Println("Update available!")

	assetName, assetURL, err := FindPluginAsset(release, qtcVersion)
	if err != nil {
		log.Fatalf("Failed to find plugin asset: %v", err)
	}
	fmt.Printf("Found asset: %s\n", assetName)

	tmpFile := filepath.Join(os.TempDir(), assetName)
	fmt.Printf("Downloading to %s...\n", tmpFile)
	if err := DownloadPlugin(assetURL, tmpFile); err != nil {
		log.Fatalf("Failed to download plugin: %v", err)
	}
	defer os.Remove(tmpFile)

	fmt.Printf("Installing to %s...\n", config.PluginPath)
	if err := InstallPlugin(tmpFile, config.PluginPath); err != nil {
		log.Fatalf("Failed to install plugin: %v", err)
	}

	fmt.Println("Plugin updated successfully!")
	fmt.Println("Please restart Qt Creator to load the new version.")
}
