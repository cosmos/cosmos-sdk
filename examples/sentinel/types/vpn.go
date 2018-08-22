package types

type Registervpn struct {
	Ip         string
	NetSpeed   NetSpeed
	PricePerGb int64
	EncMethod  string
	Location   Location
	NodeType   string
	Version    string
}

type NetSpeed struct {
	UploadSpeed   int64
	DownloadSpeed int64
}
type Location struct {
	Latitude  int64
	Longitude int64
	City      string
	Country   string
}

func NewVpnRegister(ip string, upload int64, download int64, ppgb int64, method string, latitude int64, long int64, city string, country string, nodetype string, version string) Registervpn {
	return Registervpn{
		Ip: ip,
		NetSpeed: NetSpeed{
			UploadSpeed:   upload,
			DownloadSpeed: download,
		},
		PricePerGb: ppgb,
		EncMethod:  method,
		Location: Location{
			Latitude:  latitude,
			Longitude: long,
			City:      city,
			Country:   country,
		},
		NodeType: nodetype,
		Version:  version,
	}
}
