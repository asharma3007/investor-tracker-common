package investor_tracker_common

import "os"

func FixTemplateSourceDir() {
	fileInfo, err := os.Stat(CloudfunctionSourceDir)
	if err == nil && fileInfo.IsDir() {
		_ = os.Chdir(CloudfunctionSourceDir)
	}
}
