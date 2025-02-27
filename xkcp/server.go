package xkcp

import (
	"errors"
	"io"
	"log"
	"net"

	"github.com/xtaci/kcp-go/v5"
)

type ServerConnHandler interface {
	Handle(conn *kcp.UDPSession)
}

type Server struct {
	addr    string
	conf    *KcpConfig
	lis     *kcp.Listener
	handler ServerConnHandler

	rawConn net.PacketConn
}

// NewServer creates a new xkcp server
func NewServer(addr string, conf *KcpConfig, handler ServerConnHandler) (*Server, error) {
	lis, err := kcp.ListenWithOptions(addr, GetBlockCrypt(conf.Seed, conf.Crypt), conf.FECConf.DataShard, conf.FECConf.ParityShard)
	if err != nil {
		return nil, err
	}

	if conf.DSCP > 0 {
		if err := lis.SetDSCP(conf.DSCP); err != nil {
			lis.Close()
			return nil, err
		}
	}

	if err := lis.SetReadBuffer(conf.SockBuf); err != nil {
		lis.Close()
		return nil, err
	}

	if err := lis.SetWriteBuffer(conf.SockBuf); err != nil {
		lis.Close()
		return nil, err
	}

	s := &Server{
		addr:    addr,
		conf:    conf,
		lis:     lis,
		handler: handler,
	}

	go s.loop()

	return s, nil
}

// NewServerWithConn
func NewServerWithConn(conn net.PacketConn, conf *KcpConfig, handler ServerConnHandler) (*Server, error) {
	lis, err := kcp.ServeConn(GetBlockCrypt(conf.Seed, conf.Crypt), conf.FECConf.DataShard, conf.FECConf.ParityShard, conn)
	if err != nil {
		return nil, err
	}

	if conf.DSCP > 0 {
		_ = lis.SetDSCP(conf.DSCP)
	}

	_ = lis.SetReadBuffer(conf.SockBuf)
	_ = lis.SetWriteBuffer(conf.SockBuf)

	s := &Server{
		addr:    conn.LocalAddr().String(),
		conf:    conf,
		lis:     lis,
		handler: handler,
		rawConn: conn,
	}

	go s.loop()

	return s, nil
}

// loop
func (s *Server) loop() {
	for {
		conn, err := s.lis.AcceptKCP()
		if err != nil {
			if errors.Is(err, io.ErrClosedPipe) {
				return
			}

			log.Printf("%+v\n", err)
			continue
		}

		conn.SetWriteDelay(false)
		conn.SetNoDelay(s.conf.ModeConf.NoDelay, s.conf.ModeConf.Interval, s.conf.ModeConf.Resend, s.conf.ModeConf.NoCongestion)
		conn.SetMtu(s.conf.MTU)
		conn.SetWindowSize(s.conf.SndWnd, s.conf.RcvWnd)
		conn.SetACKNoDelay(s.conf.AckNodelay)

		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn *kcp.UDPSession) {
	if s.handler != nil {
		s.handler.Handle(conn)
	}
}

// Close closes the server
func (s *Server) Close() {
	s.lis.Close()

	if s.rawConn != nil {
		s.rawConn.Close()
	}
}
