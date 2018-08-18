package types

type Registervpn struct {
	Ip       string
	Netspeed int64
	Ppgb     int64
	Location string
}

func NewVpnRegister(ip, location string, ppgb, netspeed int64) Registervpn {
	return Registervpn{
		Ip:       ip,
		Netspeed: netspeed,
		Ppgb:     ppgb,
		Location: location,
	}
}
