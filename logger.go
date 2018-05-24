package log

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/femaref/reliable_conn"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

type LoggingConfig struct {
	Host               string `yaml:"host"`
	TLS                bool   `yaml:"tls"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
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
	Logger.Out = os.Stdout
	Local.Out = os.Stdout

	/*if writable(os.Stdout) {
		Logger.Out = os.Stdout
		Local.Out = os.Stdout
	}*/

	RedirectStdlogOutput(Logger)
	DefaultLogger = Logger
}

func Configure(appName string, defaults logrus.Fields, cfg LoggingConfig) (io.Closer, error) {
	if cfg.Host != "" {
		var d reliable_conn.Dialer
		if cfg.TLS {
			d = func(network, address string) (net.Conn, error) {
				config := &tls.Config{InsecureSkipVerify: cfg.InsecureSkipVerify}
				return tls.Dial(network, address, config)
			}
		}
		conn, err := reliable_conn.DialWithDialer("tcp", cfg.Host, d)

		if err != nil {
			return nil, err
		}

		if _, ok := defaults["type"]; !ok {
			defaults["type"] = appName
		}

		hook := logrustash.New(conn, logrustash.DefaultFormatter(defaults))
		Logger.Hooks.Add(hook)
	}
	return Setup(appName)
}

func Setup(appName string) (io.Closer, error) {
	f, err := os.OpenFile(fmt.Sprintf("%s.log", appName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		if Logger.Out != nil {
			Logger.Out = io.MultiWriter(f, Logger.Out)
		} else {
			Logger.Out = f
		}
		if Local.Out != nil {
			Local.Out = io.MultiWriter(f, Local.Out)
		} else {
			Local.Out = f
		}
	}

	return f, err
}
