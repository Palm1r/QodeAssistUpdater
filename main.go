package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func printUsage() {
	fmt.Printf("%s v%s\n", AppName, AppVersion)
	fmt.Println("=========================")
	fmt.Println("\nUsage:")
	fmt.Println("  qodeassist-updater [options] <command>")
	fmt.Println("\nCommands:")
	fmt.Println("  --status     Show current plugin and Qt Creator versions")
	fmt.Println("  --install    Install or reinstall the latest plugin version")
	fmt.Println("  --update     Update plugin if a newer version is available")
	fmt.Println("  --remove     Remove installed plugin")
	fmt.Println("\nOptions:")
	fmt.Println("  --config <path>         Path to configuration file (default: config.yaml)")
	fmt.Println("  --plugin-version <ver>  Install specific plugin version (e.g., 1.2.3 or v1.2.3)")
	fmt.Println("  --checksum <hash>       Expected SHA256 checksum for verification (optional)")
	fmt.Println("  -h, --help              Show this help message")
	fmt.Println("  -v, --version           Show version information")
	fmt.Println("\nExamples:")
	fmt.Println("  qodeassist-updater --version")
	fmt.Println("  qodeassist-updater --status")
	fmt.Println("  qodeassist-updater --install")
	fmt.Println("  qodeassist-updater --update")
	fmt.Println("  qodeassist-updater --install --plugin-version 1.2.3")
	fmt.Println("  qodeassist-updater --update --plugin-version 1.2.3")
	fmt.Println("  qodeassist-updater --install --checksum abc123...")
	fmt.Println("  qodeassist-updater --remove")
	fmt.Println("  qodeassist-updater --config /path/to/config.yaml --update")
}

func printVersion() {
	fmt.Printf("%s version %s\n", AppName, AppVersion)
}

func resolveConfigPath(configPath string, defaultPath string) (string, error) {
	if configPath != defaultPath {
		return configPath, nil
	}

	execPath, err := os.Executable()
	if err != nil {
		execPath, err = filepath.Abs(os.Args[0])
		if err != nil {
			return "", fmt.Errorf("failed to resolve executable path: %w", err)
		}
	}
	execDir := filepath.Dir(execPath)
	configPathInExecDir := filepath.Join(execDir, defaultPath)

	if PathExists(configPathInExecDir) {
		return configPathInExecDir, nil
	}

	if PathExists(configPath) {
		return configPath, nil
	}

	PrintInfo("Config file not found, creating default...")
	fmt.Println()
	if err := CreateDefaultConfig(configPathInExecDir); err != nil {
		return "", err
	}

	return configPathInExecDir, nil
}

func main() {
	defaultConfigPath := "config.yaml"
	configPath := flag.String("config", defaultConfigPath, "Path to configuration file")
	pluginVersion := flag.String("plugin-version", "", "Install specific plugin version (e.g., 1.2.3 or v1.2.3)")
	checksum := flag.String("checksum", "", "Expected SHA256 checksum for verification")
	statusCmd := flag.Bool("status", false, "Show current plugin and Qt Creator versions")
	installCmd := flag.Bool("install", false, "Install or reinstall the latest plugin version")
	updateCmd := flag.Bool("update", false, "Update plugin if a newer version is available")
	removeCmd := flag.Bool("remove", false, "Remove installed plugin")

	var showHelp, showVersion bool
	flag.BoolVar(&showHelp, "help", false, "Show help message")
	flag.BoolVar(&showHelp, "h", false, "Show help message")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information")

	flag.Usage = printUsage
	flag.Parse()

	if showVersion {
		printVersion()
		os.Exit(0)
	}

	if showHelp {
		printUsage()
		os.Exit(0)
	}

	if !*statusCmd && !*installCmd && !*updateCmd && !*removeCmd {
		printUsage()
		os.Exit(0)
	}

	resolvedConfigPath, err := resolveConfigPath(*configPath, defaultConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resolve config path: %v\n", err)
		os.Exit(1)
	}

	config, err := LoadConfig(resolvedConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	var cmdErr error
	if *statusCmd {
		cmdErr = showStatus(config)
	} else if *installCmd {
		cmdErr = installPlugin(config, true, *pluginVersion, *checksum)
	} else if *updateCmd {
		cmdErr = updatePlugin(config, *pluginVersion, *checksum)
	} else if *removeCmd {
		cmdErr = removePlugin(config)
	}

	if cmdErr != nil {
		fmt.Fprintf(os.Stderr, "Command failed: %v\n", cmdErr)
		os.Exit(1)
	}
}
