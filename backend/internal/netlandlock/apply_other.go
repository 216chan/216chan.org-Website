//go:build !linux || (!amd64 && !arm64)

package netlandlock

import "fmt"

func Reexec(databasePort, listenPort uint16) error {
	return fmt.Errorf("network Landlock is supported only on Linux amd64/arm64")
}
