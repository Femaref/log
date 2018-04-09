// +build linux darwin

package log

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func RedirectStderrToFile(appName string, force_redirect bool) {
	// if it's a regular file, it's already being redirected
	// if it's a terminal, we want to see it
	// open /dev/null and check if the inode is the same -> don't redirect

	fi, _ := os.Stderr.Stat()
	fil, _ := os.OpenFile("/dev/null", os.O_WRONLY, 0755)
	nullstat, _ := fil.Stat()

	var is_regular bool = fi.Mode().IsRegular()
	var is_terminal bool = terminal.IsTerminal(int(os.Stderr.Fd()))
	var is_dev_null bool = fi.Sys().(*syscall.Stat_t).Ino == nullstat.Sys().(*syscall.Stat_t).Ino
	fil.Close()

	should_redirect := !(is_regular || is_terminal || is_dev_null) || force_redirect

	if should_redirect {
		logFile, _ := os.OpenFile(fmt.Sprintf("%s-panic.log", appName), os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0644)
		syscall.Dup2(int(logFile.Fd()), int(os.Stderr.Fd()))
		Local.Infof("redirecting stderr to %s", logFile.Name())
	}
}
