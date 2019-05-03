package log

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"syscall"

	"github.com/pkg/errors"

	logrustash "github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/femaref/log/file_signal_wrapper"
	"github.com/femaref/reliable_conn"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

type Config struct {
	Base     BaseConfig     `yaml:"base"`
	Logstash LogstashConfig `yaml:"logstash"`
}

type BaseConfig struct {
	Name   string `yaml:"name"`
	File   bool   `yaml:"file"`
	Stdout bool   `yaml:"stdout"`
}

type LogstashConfig struct {
	Host               string `yaml:"host"`
	TLS                bool   `yaml:"tls"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
}

func (this LogstashConfig) Valid() bool {
	return this.Host != ""
}

var Logger = logrus.New()
var Local = logrus.New()

func writable(f *os.File) bool {
	fi, _ := f.Stat()

	var is_regular bool = fi.Mode().IsRegular()
	var is_terminal bool = terminal.IsTerminal(int(f.Fd()))

	return is_regular || is_terminal
}

func init() {
	RedirectStdlogOutput(Logger)
	DefaultLogger = Logger
}

func Configure(cfg Config, defaults logrus.Fields) (io.Closer, error) {
	if cfg.Logstash.Valid() {
		var d reliable_conn.Dialer
		if cfg.Logstash.TLS {
			d = func(network, address string) (net.Conn, error) {
				config := &tls.Config{InsecureSkipVerify: cfg.Logstash.InsecureSkipVerify}
				return tls.Dial(network, address, config)
			}
		}
		conn, err := reliable_conn.DialWithDialer("tcp", cfg.Logstash.Host, d)

		if err != nil {
			return nil, err
		}

		if _, ok := defaults["type"]; !ok {
			defaults["type"] = cfg.Base.Name
		}

		hook := logrustash.New(conn, logrustash.DefaultFormatter(defaults))
		Logger.Hooks.Add(hook)
	}
	return Setup(cfg)
}

func Setup(cfg Config) (io.Closer, error) {

	var f io.WriteCloser
	var err error

	var w io.Writer

	if cfg.Base.File {
		// if we want to write to file, open it

		f, err = file_signal_wrapper.Append(context.Background(), fmt.Sprintf("%s.log", cfg.Base.Name), Local, syscall.SIGHUP)

		if err != nil {
			return nil, err
		}

		w = f

		// if we also want to write to stdout, add it as multiwriter
		if cfg.Base.Stdout {
			w = io.MultiWriter(w, os.Stdout)
		}
	} else if cfg.Base.Stdout {
		// are we only writing to stdout?
		w = os.Stdout
	} else {
		// nothing selected
		return nil, errors.New("at least one of File/Stdout has to be true")
	}

	Logger.Out = w
	Local.Out = w

	return f, err
}
