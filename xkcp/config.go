package xkcp

const (
	ModeFast   = "fast"
	ModeNormal = "normal"
	ModeFast2  = "fast2"
	ModeFast3  = "fast3"
)

type KcpConfig struct {
	Seed       string    `json:"seed"`
	Crypt      string    `json:"crypt"`
	MTU        int       `json:"mtu"`
	SndWnd     int       `json:"sndwnd"`
	RcvWnd     int       `json:"rcvwnd"`
	DSCP       int       `json:"dscp"`
	AckNodelay bool      `json:"acknodelay"`
	SockBuf    int       `json:"sockbuf"`
	ModeConf   *ModeConf `json:"mode"`
	FECConf    *FECConf  `json:"fec"`
}

type FECConf struct {
	DataShard   int `json:"datashard"`
	ParityShard int `json:"parityshard"`
}

type ModeConf struct {
	NoDelay      int `json:"nodelay"`
	Interval     int `json:"interval"`
	Resend       int `json:"resend"`
	NoCongestion int `json:"nc"`
}

func GetModeConf(mode string) *ModeConf {
	switch mode {
	case ModeFast:
		return &ModeConf{NoDelay: 0, Interval: 30, Resend: 2, NoCongestion: 1}
	case ModeNormal:
		return &ModeConf{NoDelay: 0, Interval: 40, Resend: 2, NoCongestion: 1}
	case ModeFast2:
		return &ModeConf{NoDelay: 1, Interval: 20, Resend: 2, NoCongestion: 1}
	case ModeFast3:
		return &ModeConf{NoDelay: 1, Interval: 10, Resend: 2, NoCongestion: 1}

	default:
		return &ModeConf{NoDelay: 0, Interval: 40, Resend: 2, NoCongestion: 1}
	}
}

func DefaultConfig() *KcpConfig {
	return &KcpConfig{
		Seed:       "test-seed",
		Crypt:      "salsa20",
		MTU:        1200,
		SndWnd:     1024,
		RcvWnd:     1024 * 4,
		AckNodelay: false,
		SockBuf:    16777217,
		ModeConf:   GetModeConf(ModeNormal),
		FECConf:    &FECConf{DataShard: 10, ParityShard: 3},
	}
}
