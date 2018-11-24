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

func TestUDPForwarder(t *testing.T) {
	pc, err := net.ListenPacket("udp", "localhost:5000")
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()
	udpAddr := pc.LocalAddr().(*net.UDPAddr)
	config := fmt.Sprintf(`
Server = "127.0.0.1"
Port = %d
`, udpAddr.Port)

	recvData := struct{
		data []byte
		len int
	}{}

	dataChan := make(chan struct{}, 1)
	go func() {
		buffer := make([]byte, 1024)
		assert.NoError(t, pc.SetReadDeadline(time.Now().Add(time.Second * 3)))
		n, _, err := pc.ReadFrom(buffer)
		assert.NoError(t, err)
		recvData.data = buffer
		recvData.len = n
		dataChan<-struct{}{}
	}()

	udp, err := NewUDPForwarderFromReader(bytes.NewBufferString(config))
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	<-dataChan
	assert.Equal(t, 63, recvData.len)

	hdr := Header{}
	recvTelem := juicer.Telemetry{}
	rdr := bytes.NewReader(recvData.data)
	assert.NoError(t, binary.Read(rdr, binary.LittleEndian, &hdr))
	assert.NoError(t, binary.Read(rdr, binary.LittleEndian, &recvTelem))
	assert.Equal(t, &newTelem, &recvTelem)
}