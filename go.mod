module github.com/jd3nn1s/juicer

require (
	github.com/brutella/can v0.0.0-20180117080637-818f1bc3aba8
	github.com/jd3nn1s/kw1281 v0.0.0-20181112031746-452ca1eb6c0b
	github.com/jd3nn1s/skytraq v0.0.0-20181112024350-f8051ca383d0
	github.com/onsi/gomega v1.4.2 // indirect
	github.com/pkg/errors v0.8.0
	github.com/sirupsen/logrus v1.0.6
	github.com/stretchr/testify v1.2.2
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
)

replace github.com/jd3nn1s/skytraq => ../skytraq

replace github.com/jd3nn1s/kw1281 => ../kw1281

replace github.com/brutella/can => ../can
