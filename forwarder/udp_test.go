package forwarder

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/jd3nn1s/juicer"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"testing"
	"time"
)

type udpRecv struct {
	data []byte
	len  int
}

func startServer(ctx context.Context, t *testing.T) (int, net.PacketConn, chan udpRecv) {
	pc, err := net.ListenPacket("udp", "localhost:5000")
	if err != nil {
		log.Fatal(err)
	}
	udpAddr := pc.LocalAddr().(*net.UDPAddr)

	go func() {
		select {
		case <-ctx.Done():
			pc.Close()
		}
	}()

	dataChan := make(chan udpRecv, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			buffer := make([]byte, 1024)
			assert.NoError(t, pc.SetReadDeadline(time.Now().Add(time.Second*3)))
			n, _, _ := pc.ReadFrom(buffer)
			dataChan <- udpRecv{
				len:  n,
				data: buffer,
			}
		}
	}()
	return udpAddr.Port, pc, dataChan
}

func config(port int) string {
	return fmt.Sprintf(`
Server = "127.0.0.1"
Port = %d
`, port)
}

func decodePacket(t *testing.T, buf []byte) *juicer.Telemetry {
	hdr := Header{}
	recvTelem := juicer.Telemetry{}
	rdr := bytes.NewReader(buf)
	assert.NoError(t, binary.Read(rdr, binary.LittleEndian, &hdr))
	assert.Equal(t, TypeTelemetry, hdr.Type)
	assert.NoError(t, binary.Read(rdr, binary.LittleEndian, &recvTelem))
	return &recvTelem
}

func TestTicker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	port, pc, dataChan := startServer(ctx, t)
	defer pc.Close()

	oldMinSendDelay := minSendDelay
	oldSendRateLimit := sendRateLimit
	defer func() {
		minSendDelay = oldMinSendDelay
		sendRateLimit = oldSendRateLimit
	}()

	minSendDelay = time.Millisecond * 10
	sendRateLimit = time.Millisecond

	udp, err := NewUDPForwarderFromReader(bytes.NewBufferString(config(port)))
	assert.NoError(t, err)

	go func() {
		_ = udp.Start(ctx)
	}()

	telem := juicer.Telemetry{
		Speed: 3,
	}
	assert.NoError(t, udp.Forward(&telem, &juicer.Telemetry{}))

	recvData := <-dataChan
	assert.Equal(t, 63, recvData.len)

	for n := 0; n < 3; n++ {
		start := time.Now()
		recvData := <-dataChan
		delay := time.Now().Sub(start)
		assert.Equal(t, 63, recvData.len)
		assert.True(t, delay >= minSendDelay)
		assert.True(t, delay < minSendDelay+time.Millisecond*10)
		assert.Equal(t, &telem, decodePacket(t, recvData.data))
	}
}

func TestUDPForwarder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port, pc, dataChan := startServer(ctx, t)
	defer pc.Close()

	udp, err := NewUDPForwarderFromReader(bytes.NewBufferString(config(port)))
	assert.NoError(t, err)

	go func() {
		_ = udp.Start(ctx)
	}()

	newTelem := juicer.Telemetry{
		RPM:            1,
		OilPressure:    2,
		Speed:          3,
		FuelRemaining:  4,
		FuelLevel:      5,
		OilTemp:        6,
		CoolantTemp:    7,
		AirIntakeTemp:  8,
		BatteryVoltage: 9,
		Latitude:       10,
		Longitude:      11,
		Altitude:       12,
		Track:          13,
		GPSSpeed:       14,
		GasPedalAngle:  15,
	}
	prevTelem := juicer.Telemetry{}
	assert.NoError(t, udp.Forward(&newTelem, &prevTelem))

	recvData := <-dataChan
	assert.Equal(t, 63, recvData.len)

	recvTelem := decodePacket(t, recvData.data)
	assert.Equal(t, &newTelem, &recvTelem)
	assert.NoError(t, udp.Close())
}
