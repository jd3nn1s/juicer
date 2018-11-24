package forwarder

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/jd3nn1s/juicer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
	"unsafe"
)

type Header struct {
	Type uint8
}

var maxTelemetrySize = int(unsafe.Sizeof(Header{}) + unsafe.Sizeof(juicer.Telemetry{}))

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

	conn    net.Conn
	fwdChan chan *juicer.Telemetry
	wg sync.WaitGroup
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
	udp := &UDPForwarder{
		Config:  &config,
		fwdChan: make(chan *juicer.Telemetry, 1),
	}
	if err = udp.connect(); err != nil {
		return nil, err
	}
	return udp, nil
}

func (udp *UDPForwarder) Close() error {
	return udp.conn.Close()
}

func (udp *UDPForwarder) Forward(newTelemetry *juicer.Telemetry, prevTelemetry *juicer.Telemetry) error {
	telemCopy := *newTelemetry
	select {
	// copy telemetry as we're processing it on another go-routine
	case udp.fwdChan <- &telemCopy:
	default:
		// if channel is full, skip
	}
	return nil
}

func (udp *UDPForwarder) Start(ctx context.Context) error {
	limiter := time.Tick(100 * time.Millisecond)
	for {
		<-limiter
		select {
		case t := <-udp.fwdChan:
			if err := udp.forward(t); err != nil {
				log.Error("unable to forward telemetry to server ", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (udp *UDPForwarder) forward(telem *juicer.Telemetry) error {
	buf := bytes.NewBuffer([]byte{})
	hdr := Header{
		Type: TypeTelemetry,
	}
	if err := binary.Write(buf, binary.LittleEndian, &hdr); err != nil {
		return errors.Wrap(err, "unable to write udp packet header")
	}
	if err := binary.Write(buf, binary.LittleEndian, telem); err != nil {
		return errors.Wrap(err, "unable to write telemetry udp packet")
	}
	return binary.Write(udp.conn, binary.LittleEndian, buf.Bytes())
}

func (udp *UDPForwarder) connect() error {
	writeBufSize := maxTelemetrySize * 2

	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d",
		udp.Config.Server,
		udp.Config.Port))
	if err != nil {
		return err
	}
	udpConn := conn.(*net.UDPConn)
	if err = udpConn.SetWriteBuffer(writeBufSize); err != nil {
		return errors.Wrapf(err, "unable to set OS write buffer to %v", writeBufSize)
	}

	udp.conn = conn
	return nil
}
