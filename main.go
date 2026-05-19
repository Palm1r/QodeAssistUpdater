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
	fmt.Println("  --status         Show current plugin and Qt Creator versions")
	fmt.Println("  --install        Install or reinstall the latest plugin version")
	fmt.Println("  --update         Update plugin if a newer version is available")
	fmt.Println("  --remove         Remove installed plugin")
	fmt.Println("  --list-versions  List all available plugin versions (>= 0.5.9)")
	fmt.Println("\nOptions:")
	fmt.Println("  --config <path>         Path to configuration file (default: config.yaml)")
	fmt.Println("  --qtc-path <path>       Path to Qt Creator (run without config.yaml)")
	fmt.Println("  --plugin-path <path>    Path to plugin directory (run without config.yaml)")
	fmt.Println("  --plugin-version <ver>  Install specific plugin version (e.g., 0.8.1 or v0.8.1)")
	fmt.Println("  --checksum <hash>       Expected SHA256 checksum for verification (optional)")
	fmt.Println("  -y, --yes               Automatic yes to prompts (non-interactive mode)")
	fmt.Println("  -h, --help              Show this help message")
	fmt.Println("  -v, --version           Show version information")
	fmt.Println("\nExamples:")
	fmt.Println("  qodeassist-updater --version")
	fmt.Println("  qodeassist-updater --status")
	fmt.Println("  qodeassist-updater --install")
	fmt.Println("  qodeassist-updater --update")
	fmt.Println("  qodeassist-updater --update --yes")
	fmt.Println("  qodeassist-updater --install --plugin-version 0.8.1")
	fmt.Println("  qodeassist-updater --update --plugin-version 0.8.0")
	fmt.Println("  qodeassist-updater --install --checksum abc123...")
	fmt.Println("  qodeassist-updater --remove")
	fmt.Println("  qodeassist-updater --remove --yes")
	fmt.Println("  qodeassist-updater --list-versions")
	fmt.Println("  qodeassist-updater --config /path/to/config.yaml --update")
	fmt.Println("  qodeassist-updater --update --qtc-path ~/qtcreator-19.0.1 --plugin-path ~/path/to/plugins")
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
	qtcPathFlag := flag.String("qtc-path", "", "Path to Qt Creator (run without config.yaml)")
	pluginPathFlag := flag.String("plugin-path", "", "Path to plugin directory (run without config.yaml)")
	pluginVersion := flag.String("plugin-version", "", "Install specific plugin version (e.g., 0.8.1 or v0.8.1)")
	checksum := flag.String("checksum", "", "Expected SHA256 checksum for verification")
	statusCmd := flag.Bool("status", false, "Show current plugin and Qt Creator versions")
	installCmd := flag.Bool("install", false, "Install or reinstall the latest plugin version")
	updateCmd := flag.Bool("update", false, "Update plugin if a newer version is available")
	removeCmd := flag.Bool("remove", false, "Remove installed plugin")
	listVersionsCmd := flag.Bool("list-versions", false, "List all available plugin versions")

	var yesFlag bool
	flag.BoolVar(&yesFlag, "yes", false, "Automatic yes to prompts (non-interactive mode)")
	flag.BoolVar(&yesFlag, "y", false, "Automatic yes to prompts (non-interactive mode)")

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

	if !*statusCmd && !*installCmd && !*updateCmd && !*removeCmd && !*listVersionsCmd {
		printUsage()
		os.Exit(0)
	}

	// Handle list-versions command (doesn't need config)
	if *listVersionsCmd {
		cmdErr := listVersions()
		if cmdErr != nil {
			fmt.Fprintf(os.Stderr, "Command failed: %v\n", cmdErr)
			os.Exit(1)
		}
		os.Exit(0)
	}

	var config *Config
	var err error

	if *qtcPathFlag != "" || *pluginPathFlag != "" {
		config, err = NewConfigFromArgs(*qtcPathFlag, *pluginPathFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to build configuration: %v\n", err)
			os.Exit(1)
		}
	} else {
		resolvedConfigPath, resolveErr := resolveConfigPath(*configPath, defaultConfigPath)
		if resolveErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to resolve config path: %v\n", resolveErr)
			os.Exit(1)
		}

		config, err = LoadConfig(resolvedConfigPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
			os.Exit(1)
		}
	}

	var cmdErr error
	if *statusCmd {
		cmdErr = showStatus(config)
	} else if *installCmd {
		cmdErr = installPlugin(config, true, *pluginVersion, *checksum)
	} else if *updateCmd {
		cmdErr = updatePlugin(config, *pluginVersion, *checksum, yesFlag)
	} else if *removeCmd {
		cmdErr = removePlugin(config, yesFlag)
	}

	if cmdErr != nil {
		fmt.Fprintf(os.Stderr, "Command failed: %v\n", cmdErr)
		os.Exit(1)
	}
}
