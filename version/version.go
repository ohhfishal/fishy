package version

import (
    "runtime/debug"
)

func Version() string {
    if info, ok := debug.ReadBuildInfo(); ok {
        return info.Main.Version
    }
    return "(unknown)"
}
