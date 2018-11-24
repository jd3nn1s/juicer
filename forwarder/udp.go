package forwarder

import (
	"github.com/BurntSushi/toml"
	"github.com/jd3nn1s/juicer"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Header struct {
	Type uint8
}

const (
	TypeTelemetry = 1
	TypeTiming    = 2
)

type UDPConfig struct {
	Server string
	Port   int
}

type UDPForwarder struct {
	Config *UDPConfig
}

func NewUDPForwarder(fileName string) (*UDPForwarder, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to determine binary location")
	}
	file, err := os.Open(filepath.Join(dir, fileName))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to open file %s", fileName)
	}
	return NewUDPForwarderFromReader(file)
}

func NewUDPForwarderFromReader(configReader io.Reader) (*UDPForwarder, error) {
	configData, err := ioutil.ReadAll(configReader)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read config reader")
	}
	config := UDPConfig{}
	if _, err := toml.Decode(string(configData), &config); err != nil {
		return nil, errors.Wrapf(err, "unable to load udp forwarder configuration")
	}
	return &UDPForwarder{
		Config: &config,
	}, nil
}

func (udp *UDPForwarder) Forward(newTelemetry *juicer.Telemetry, prevTelemetry *juicer.Telemetry) error {
	return nil
}
