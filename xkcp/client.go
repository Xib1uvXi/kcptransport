package xkcp

import (
	"crypto/rand"
	"encoding/binary"
	"net"

	"github.com/xtaci/kcp-go/v5"
)

type Client struct {
	*kcp.UDPSession
}

// NewClient creates a new xkcp client
func NewClient(remoteAddr string, conf *KcpConfig) (*Client, error) {
	// default UDP connection
	kcpconn, err := kcp.DialWithOptions(remoteAddr, GetBlockCrypt(conf.Seed, conf.Crypt), conf.FECConf.DataShard, conf.FECConf.ParityShard)
	if err != nil {
		return nil, err
	}

	client, err := newClient(kcpconn, conf)
	if err != nil {
		kcpconn.Close()
		return nil, err
	}

	return client, nil
}

// NewClientWithLocal creates a new xkcp client with local address
func NewClientWithLocal(local, remote string, conf *KcpConfig) (*Client, error) {
	localAddr, err := net.ResolveUDPAddr("udp", local)
	if err != nil {
		return nil, err
	}

	remoteAddr, err := net.ResolveUDPAddr("udp", remote)
	if err != nil {
		return nil, err
	}

	localUdpConn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, err
	}

	kcpconn, err := kcp.NewConn4(genConvid(), remoteAddr, GetBlockCrypt(conf.Seed, conf.Crypt), conf.FECConf.DataShard, conf.FECConf.ParityShard, true, localUdpConn)
	if err != nil {
		return nil, err
	}

	client, err := newClient(kcpconn, conf)
	if err != nil {
		kcpconn.Close()
		return nil, err
	}

	return client, nil
}

// NewClientWithConn creates a new xkcp client with a packet connection
func NewClientWithConn(conn net.PacketConn, remoteAddr net.Addr, conf *KcpConfig) (*Client, error) {
	kcpconn, err := kcp.NewConn4(genConvid(), remoteAddr, GetBlockCrypt(conf.Seed, conf.Crypt), conf.FECConf.DataShard, conf.FECConf.ParityShard, true, conn)
	if err != nil {
		return nil, err
	}

	client, err := newClient(kcpconn, conf)
	if err != nil {
		kcpconn.Close()
		return nil, err
	}

	return client, nil
}

// newClient creates a new xkcp client
func newClient(kcpconn *kcp.UDPSession, conf *KcpConfig) (*Client, error) {
	kcpconn.SetWriteDelay(false)
	kcpconn.SetNoDelay(conf.ModeConf.NoDelay, conf.ModeConf.Interval, conf.ModeConf.Resend, conf.ModeConf.NoCongestion)
	kcpconn.SetWindowSize(conf.SndWnd, conf.RcvWnd)
	kcpconn.SetMtu(conf.MTU)
	kcpconn.SetACKNoDelay(conf.AckNodelay)

	if conf.DSCP > 0 {
		_ = kcpconn.SetDSCP(conf.DSCP)
	}

	_ = kcpconn.SetReadBuffer(conf.SockBuf)
	_ = kcpconn.SetWriteBuffer(conf.SockBuf)

	client := &Client{
		UDPSession: kcpconn,
	}

	return client, nil
}

// genConvid generates a unique conversation id
func genConvid() uint32 {
	var convid uint32
	binary.Read(rand.Reader, binary.LittleEndian, &convid)
	return convid
}
