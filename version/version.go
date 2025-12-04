package version

import (
	"runtime/debug"
)

const Repo = "https://github.com/ohhfishal/fishy"
const RepoMarkdown = "[fishy](" + Repo + ")"

func Version() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return "(unknown)"
}
