package main

import (
	"fmt"
	"runtime"
)

func GetPluginFileName() string {
	switch runtime.GOOS {
	case "darwin":
		return "libQodeAssist.dylib"
	case "windows":
		return "QodeAssist.dll"
	default:
		return "libQodeAssist.so"
	}
}

func GetPlatformArchName() (platform, arch string, err error) {
	switch runtime.GOOS {
	case "windows":
		platform = "Windows"
	case "linux":
		platform = "Linux"
	case "darwin":
		platform = "macOS"
	default:
		return "", "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if runtime.GOOS == "darwin" {
		arch = "universal"
	} else {
		switch runtime.GOARCH {
		case "amd64":
			arch = "x64"
		case "arm64":
			arch = "arm64"
		default:
			arch = runtime.GOARCH
		}
	}

	return platform, arch, nil
}
