package file_signal_wrapper

import (
	"context"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
)

type Options struct {
	Name string
	Flag int
	Perm os.FileMode
}

func Open(ctx context.Context, name string, logger *logrus.Logger, signals ...os.Signal) (*FileSignalWrapper, error) {
	return New(ctx, Options{Name: name, Flag: os.O_RDONLY, Perm: 0}, logger, signals...)
}

func Append(ctx context.Context, name string, logger *logrus.Logger, signals ...os.Signal) (*FileSignalWrapper, error) {
	return New(ctx, Options{Name: name, Flag: os.O_APPEND | os.O_CREATE | os.O_WRONLY, Perm: 0644}, logger, signals...)
}

func New(ctx context.Context, opts Options, logger *logrus.Logger, signals ...os.Signal) (*FileSignalWrapper, error) {
	name, err := filepath.Abs(opts.Name)

	if err != nil {
		return nil, err
	}
	ctx, fn := context.WithCancel(ctx)

	fsw := &FileSignalWrapper{
		name: name,
		flag: opts.Flag,
		perm: opts.Perm,

		ctx:    ctx,
		cancel: fn,
		logger: logger,
	}

	err = fsw.cycle()

	if err != nil {
		return nil, err
	}

	go fsw.run(signals...)

	return fsw, nil
}

type FileSignalWrapper struct {
	name string
	flag int
	perm os.FileMode

	f *os.File
	m sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	logger *logrus.Logger
}

func (this *FileSignalWrapper) run(signals ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)
	for {
		select {
		case <-c:
			err := this.cycle()
			if this.logger != nil {
				this.logger.Error(err)
			}
		case <-this.ctx.Done():
			return
		}
	}
}
func (this *FileSignalWrapper) cycle() error {
	this.m.Lock()
	defer this.m.Unlock()

	var err error

	if this.f != nil {
		err = this.f.Close()

		if err != nil {
			return err
		}
	}

	this.f, err = os.OpenFile(this.name, this.flag, this.perm)

	if err != nil {
		return err
	}

	return nil
}

// check impl
var _ = io.Reader((*FileSignalWrapper)(nil))
var _ = io.Writer((*FileSignalWrapper)(nil))
var _ = io.Closer((*FileSignalWrapper)(nil))

func (this *FileSignalWrapper) Read(p []byte) (int, error) {
	this.m.RLock()
	defer this.m.RUnlock()

	return this.f.Read(p)
}
func (this *FileSignalWrapper) Write(p []byte) (int, error) {
	this.m.RLock()
	defer this.m.RUnlock()

	return this.f.Write(p)
}

func (this *FileSignalWrapper) Close() error {
	this.m.Lock()
	defer this.m.Unlock()

	this.cancel()
	if this.f != nil {
		err := this.f.Close()

		if err != nil {
			return err
		}

	}
	return nil
}
