package version

import "runtime/debug"

// Version is set by release builds.
var Version = "dev"

// String returns the best available version label.
func String() string {
	if Version != "" && Version != "dev" {
		return Version
	}
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return Version
}
