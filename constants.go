package main

import "time"

const (
	AppVersion = "0.1.0"
	AppName    = "QodeAssist Plugin Updater"

	GithubRepo = "Palm1r/QodeAssist"

	DefaultDirPermissions  = 0755
	DefaultFilePermissions = 0644
	ExecutablePermissions  = 0755

	ExecutableBitMask = 0111
	GroupWriteMask    = 0022

	MaxExtractSize = 1024 * 1024 * 1024
	MaxFiles       = 10000

	MaxDownloadSize = 500 * 1024 * 1024

	HTTPTimeout = 30 * time.Second

	MaxGitHubAPIResponseSize = 10 * 1024 * 1024
)
