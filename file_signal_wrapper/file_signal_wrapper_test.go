package file_signal_wrapper

import (
	"context"
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsNewFilePointer(t *testing.T) {
	fsw, err := Append(context.Background(), "test.log", nil)
	defer os.Remove("test.log")

	assert.NoError(t, err)

	old_f := fsw.f
	err = fsw.cycle()

	assert.NoError(t, err)

	assert.NotEqual(t, old_f, fsw.f)
}

func TestClose(t *testing.T) {
	fsw, err := Append(context.Background(), "test.log", nil)
	defer os.Remove("test.log")

	assert.NoError(t, err)

	before_routines := runtime.NumGoroutine()
	err = fsw.Close()
	assert.NoError(t, err)
	time.Sleep(time.Second)
	after_routines := runtime.NumGoroutine()

	assert.Equal(t, before_routines-1, after_routines)
}

func TestReactsToSignal(t *testing.T) {
	fsw, err := Append(context.Background(), "test.log", nil, syscall.SIGHUP)
	defer os.Remove("test.log")

	assert.NoError(t, err)

	old_f := fsw.f

	pid := os.Getpid()

	process, err := os.FindProcess(pid)

	assert.NoError(t, err)
	err = process.Signal(syscall.SIGHUP)
	time.Sleep(1000 * time.Millisecond)
	assert.NoError(t, err)

	assert.NotEqual(t, old_f, fsw.f)
}
