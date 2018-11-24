package forwarder

type Header struct {
	Type uint8
}

const (
	TypeTelemetry = 1
	TypeTiming    = 2
)

type Telemetry struct {
	RPM         float32
	OilPressure float32
	Speed       float32

	FuelRemaining float32
	FuelLevel     uint8

	OilTemp        float32
	CoolantTemp    float32
	AirIntakeTemp  float32
	BatteryVoltage float32

	Latitude      float64
	Longitude     float64
	Altitude      float32
	Track         float32
	GPSSpeed      float32
	GasPedalAngle uint8
}
