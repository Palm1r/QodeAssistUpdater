package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func printUsage() {
	fmt.Printf("%s v%s\n", AppName, AppVersion)
	fmt.Println("=========================")
	fmt.Println("\nUsage:")
	fmt.Println("  qodeassist-updater [options] <command>")
	fmt.Println("\nCommands:")
	fmt.Println("  --install        Download and install the plugin")
	fmt.Println("  --update         Reinstall, removing old plugin files first")
	fmt.Println("  --remove         Remove installed plugin")
	fmt.Println("  --list-versions  List all available plugin versions (>= 0.5.9)")
	fmt.Println("\nOptions:")
	fmt.Println("  --qtc-version <ver>     Qt Creator version (required for install/update), e.g. 14.0.2")
	fmt.Println("  --plugin-dir <path>     Directory to install into / remove from (required for install/update/remove)")
	fmt.Println("  --plugin-version <ver>  Install specific plugin version (e.g. 0.8.1 or v0.8.1)")
	fmt.Println("  --checksum <hash>       Expected SHA256 checksum for verification (optional)")
	fmt.Println("  --wait-pid <pid>        Wait for this process to exit before touching files")
	fmt.Println("  --wait-timeout <sec>    Timeout for --wait-pid in seconds (default 120, 0 = infinite)")
	fmt.Println("  --log-file <path>       Mirror all output into a log file")
	fmt.Println("  --relaunch <path>       Launch this application after a successful install/update")
	fmt.Println("  -y, --yes               Automatic yes to prompts (non-interactive mode)")
	fmt.Println("  -h, --help              Show this help message")
	fmt.Println("  -v, --version           Show version information")
	fmt.Println("\nExamples:")
	fmt.Println("  qodeassist-updater --install --qtc-version 14.0.2 --plugin-dir ~/QtPlugins")
	fmt.Println("  qodeassist-updater --update --qtc-version 14.0.2 --plugin-dir ~/QtPlugins --yes")
	fmt.Println("  qodeassist-updater --remove --plugin-dir ~/QtPlugins")
	fmt.Println("  qodeassist-updater --list-versions")
	fmt.Println("\nDetached update from Qt Creator (wait for it to quit, then relaunch):")
	fmt.Println("  qodeassist-updater --update --qtc-version 14.0.2 --plugin-dir ~/QtPlugins \\")
	fmt.Println("      --yes --wait-pid 12345 --log-file ~/qodeassist-update.log \\")
	fmt.Println("      --relaunch \"/Applications/Qt Creator.app\"")
}

func printVersion() {
	fmt.Printf("%s version %s\n", AppName, AppVersion)
}

func resolvePluginDir(dir string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("--plugin-dir is required")
	}
	expanded, err := ExpandPath(dir)
	if err != nil {
		return "", fmt.Errorf("failed to expand plugin directory: %w", err)
	}
	return expanded, nil
}

func resolveQtcVersion(version string) (*Version, error) {
	if version == "" {
		return nil, fmt.Errorf("--qtc-version is required for install/update")
	}
	v, err := ParseVersion(version)
	if err != nil {
		return nil, fmt.Errorf("invalid --qtc-version: %w", err)
	}
	return v, nil
}

func main() {
	os.Exit(run())
}

func run() int {
	qtcVersionFlag := flag.String("qtc-version", "", "Qt Creator version (e.g. 14.0.2)")
	pluginDirFlag := flag.String("plugin-dir", "", "Directory to install into / remove from")
	pluginVersion := flag.String("plugin-version", "", "Install specific plugin version (e.g. 0.8.1 or v0.8.1)")
	checksum := flag.String("checksum", "", "Expected SHA256 checksum for verification")
	waitPid := flag.Int("wait-pid", 0, "Wait for this process to exit before touching files")
	waitTimeout := flag.Int("wait-timeout", 120, "Timeout for --wait-pid in seconds (0 = infinite)")
	logFile := flag.String("log-file", "", "Mirror all output into a log file")
	relaunchPath := flag.String("relaunch", "", "Launch this application after a successful install/update")
	installCmd := flag.Bool("install", false, "Download and install the plugin")
	updateCmd := flag.Bool("update", false, "Reinstall, removing old plugin files first")
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
		return 0
	}

	if showHelp {
		printUsage()
		return 0
	}

	if *logFile != "" {
		closeLog, err := SetupLogFile(*logFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
		defer closeLog()
	}

	if !*installCmd && !*updateCmd && !*removeCmd && !*listVersionsCmd {
		printUsage()
		return 0
	}

	if *listVersionsCmd {
		if err := listVersions(); err != nil {
			PrintFatal(fmt.Sprintf("Command failed: %v", err))
			return 1
		}
		return 0
	}

	pluginDir, err := resolvePluginDir(*pluginDirFlag)
	if err != nil {
		PrintFatal(err.Error())
		return 1
	}

	var qtcVersion *Version
	if *installCmd || *updateCmd {
		qtcVersion, err = resolveQtcVersion(*qtcVersionFlag)
		if err != nil {
			PrintFatal(err.Error())
			return 1
		}
	}

	if *waitPid > 0 {
		PrintStep(fmt.Sprintf("Waiting for process %d to exit...", *waitPid))
		timeout := time.Duration(*waitTimeout) * time.Second
		if err := WaitForProcessExit(*waitPid, timeout); err != nil {
			PrintFatal(err.Error())
			return 1
		}
		PrintSuccess("Process exited")
	}

	var cmdErr error
	switch {
	case *removeCmd:
		cmdErr = removePlugin(pluginDir, yesFlag)
	case *installCmd:
		cmdErr = installPlugin(qtcVersion, pluginDir, *pluginVersion, *checksum, yesFlag)
	case *updateCmd:
		cmdErr = updatePlugin(qtcVersion, pluginDir, *pluginVersion, *checksum, yesFlag)
	}

	if cmdErr != nil {
		PrintFatal(fmt.Sprintf("Command failed: %v", cmdErr))
		return 1
	}

	if *relaunchPath != "" && (*installCmd || *updateCmd) {
		PrintStep("Relaunching application...")
		if err := RelaunchApplication(*relaunchPath); err != nil {
			PrintWarning(fmt.Sprintf("Failed to relaunch: %v", err))
		}
	}

	return 0
}
