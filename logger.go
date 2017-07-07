package log

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/femaref/reliable_conn"
	"io"
	"os"
	"net"
	"crypto/tls"
)

type LoggingConfig struct {
	Host string `yaml:"host"`
	TLS bool `yaml:"tls"`
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
}

var Logger = logrus.New()
var Local = logrus.New()

func init() {
	// Output to stderr instead of stdout, could also be a file.
	Logger.Out = os.Stdout
	Local.Out = os.Stdout
	RedirectStdlogOutput(Logger)
	DefaultLogger = Logger
}

func Configure(appName string, cfg LoggingConfig) (io.Closer, error) {
	if cfg.Host != "" {
        var d reliable_conn.Dialer
	    if cfg.TLS {
            d = func(network, address string) (net.Conn, error) {
                config := &tls.Config{InsecureSkipVerify:cfg.InsecureSkipVerify}
                return tls.Dial(network, address, config)
            }
	    }
		conn, err := reliable_conn.DialWithDialer("tcp", cfg.Host, d)

		if err != nil {
			return nil, err
		}

		hook := logrustash.New(conn, logrustash.DefaultFormatter(logrus.Fields{"type": appName}))
		Logger.Hooks.Add(hook)
	}
	return Setup(appName)
}

func Setup(appName string) (io.Closer, error) {
	f, err := os.OpenFile(fmt.Sprintf("%s.log", appName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		Logger.Out = io.MultiWriter(f, Logger.Out)
		Local.Out = io.MultiWriter(f, Local.Out)
	}

	return f, err
}
