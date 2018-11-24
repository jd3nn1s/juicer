package juicer

type gpsData struct {
	Latitude  int
	Longitude int
	Altitude  int
	Track     float64
	Speed     float64
}

type ecuData struct {
	GasPedalAngle  int
	RPM            float32
	OilPressure    float32
	Speed          int
	CoolantTemp    float32
	AirIntakeTemp  float32
	BatteryVoltage float32
}

type canSensorData struct {
	FuelRemaining float32
	FuelLevel     int
	CoolantTemp   int
	OilTemp       int
}

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
