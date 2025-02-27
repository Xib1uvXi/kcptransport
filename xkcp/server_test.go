package xkcp

import (
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/lossyconn"
)

var baseport = uint32(10000)

func getTestAddr() string {
	port := int(atomic.AddUint32(&baseport, 1))
	return fmt.Sprintf("127.0.0.1:%v", port)
}

type testSinkHandler struct{}

func (h *testSinkHandler) Handle(conn *kcp.UDPSession) {
	defer conn.Close()

	buf := make([]byte, 65536)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			return
		}
	}
}

type testEchoHandler struct{}

func (h *testEchoHandler) Handle(conn *kcp.UDPSession) {
	defer conn.Close()

	buf := make([]byte, 65536)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		conn.Write(buf[:n])
	}
}

type testTinyBufferEchoHandler struct{}

func (h *testTinyBufferEchoHandler) Handle(conn *kcp.UDPSession) {
	defer conn.Close()

	buf := make([]byte, 2)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		conn.Write(buf[:n])
	}
}

func TestNewServer(t *testing.T) {
	server, err := NewServer(getTestAddr(), DefaultConfig(), &testSinkHandler{})
	require.NoError(t, err)
	defer func() {
		server.Close()
		time.Sleep(100 * time.Millisecond)
	}()

	time.Sleep(100 * time.Millisecond)
}

func TestNewServer2(t *testing.T) {
	server, err := NewServer(getTestAddr(), DefaultConfig(), &testEchoHandler{})
	require.NoError(t, err)
	defer func() {
		server.Close()
		time.Sleep(100 * time.Millisecond)
	}()

	time.Sleep(100 * time.Millisecond)
}

func TestNewServer3(t *testing.T) {
	server, err := NewServer(getTestAddr(), DefaultConfig(), &testTinyBufferEchoHandler{})
	require.NoError(t, err)
	defer func() {
		server.Close()
		time.Sleep(100 * time.Millisecond)
	}()

	time.Sleep(100 * time.Millisecond)
}

func TestNewServer4(t *testing.T) {
	server, err := NewServer(getTestAddr(), DefaultConfig(), nil)
	require.NoError(t, err)
	defer func() {
		server.Close()
		time.Sleep(100 * time.Millisecond)
	}()

	time.Sleep(100 * time.Millisecond)
}

func TestNewServerWithConn(t *testing.T) {
	saddr := getTestAddr()
	conf := DefaultConfig()

	conn, err := net.ListenPacket("udp", saddr)
	require.NoError(t, err)

	server, err := NewServerWithConn(conn, conf, &testSinkHandler{})
	require.NoError(t, err)
	defer func() {
		server.Close()
		time.Sleep(100 * time.Millisecond)
	}()

	time.Sleep(100 * time.Millisecond)
}

func TestNewServerWithConn2(t *testing.T) {
	conf := DefaultConfig()
	conf.Crypt = "null"
	serverLC, err := lossyconn.NewLossyConn(0.1, 100)
	if err != nil {
		t.Fatalf("failed to create lossyconn: %v", err)
	}

	server, err := NewServerWithConn(serverLC, conf, &testSinkHandler{})
	require.NoError(t, err)
	defer func() {
		server.Close()
		time.Sleep(100 * time.Millisecond)
	}()

	time.Sleep(100 * time.Millisecond)
}

func TestNewServer_Sink(t *testing.T) {
	saddr := getTestAddr()
	conf := DefaultConfig()
	conf.DSCP = 46
	server, err := NewServer(saddr, conf, &testSinkHandler{})
	require.NoError(t, err)
	defer func() {
		server.Close()
		time.Sleep(100 * time.Millisecond)
	}()

	client, err := NewClient(saddr, DefaultConfig())
	require.NoError(t, err)
	defer client.Close()

	for i := 0; i < 10; i++ {
		n, err := client.Write([]byte("hello"))
		require.NoError(t, err)
		require.Equal(t, 5, n)
	}

	time.Sleep(100 * time.Millisecond)
}

// sink
func sinkBenchmark(b *testing.B, nbytes int, serverConf *KcpConfig, clientConf *KcpConfig) {
	saddr := getTestAddr()
	server, err := NewServer(saddr, serverConf, &testSinkHandler{})
	require.NoError(b, err)
	defer server.Close()

	b.ReportAllocs()

	client, err := NewClient(saddr, clientConf)
	require.NoError(b, err)
	defer client.Close()

	// sender
	buf := make([]byte, nbytes)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.Write(buf); err != nil {
			b.Fatalf("failed to write to client: %v", err) // stop benchmark
		}
	}

	b.SetBytes(int64(nbytes))
}

func baseSinkBenchmarkRunner(nbytes int) func(*testing.B) {
	defaultConf := DefaultConfig()
	return func(b *testing.B) {
		sinkBenchmark(b, nbytes, defaultConf, defaultConf)
	}
}

func perfSinkBenchmarkRunner(nbytes int) func(*testing.B) {
	serverConf := DefaultConfig()
	serverConf.SockBuf = 4 * 1024 * 1024
	serverConf.MTU = 1400
	serverConf.DSCP = 46
	serverConf.ModeConf = GetModeConf(ModeFast3)
	serverConf.SndWnd = 4096
	serverConf.RcvWnd = 4096
	serverConf.Crypt = "null"
	serverConf.FECConf.DataShard = 0
	serverConf.FECConf.ParityShard = 0

	clientConf := DefaultConfig()
	clientConf.Crypt = "null"
	clientConf.SndWnd = 1024
	clientConf.RcvWnd = 1024
	clientConf.SockBuf = 16 * 1024 * 1024
	clientConf.MTU = 1400
	clientConf.ModeConf = GetModeConf(ModeFast3)
	clientConf.FECConf.DataShard = 0
	clientConf.FECConf.ParityShard = 0
	return func(b *testing.B) {
		sinkBenchmark(b, nbytes, serverConf, clientConf)
	}
}

func BenchmarkSink_Base_Speed4K(b *testing.B) {
	baseSinkBenchmarkRunner(4096)(b)
}

func BenchmarkSink_Perf_Speed4K(b *testing.B) {
	perfSinkBenchmarkRunner(4096)(b)
}

func BenchmarkSink_Base_Speed1M(b *testing.B) {
	baseSinkBenchmarkRunner(1048576)(b)
}

func BenchmarkSink_Perf_Speed1M(b *testing.B) {
	serverConf := DefaultConfig()
	serverConf.SockBuf = 4 * 1024 * 1024
	serverConf.MTU = 1200
	serverConf.DSCP = 46
	serverConf.ModeConf = GetModeConf(ModeFast3)
	serverConf.SndWnd = 4096
	serverConf.RcvWnd = 4096
	serverConf.Crypt = "null"
	serverConf.FECConf.DataShard = 0
	serverConf.FECConf.ParityShard = 0

	clientConf := DefaultConfig()
	clientConf.Crypt = "null"
	clientConf.SndWnd = 1024
	clientConf.RcvWnd = 1024
	clientConf.SockBuf = 16 * 1024 * 1024
	clientConf.MTU = 1200
	clientConf.ModeConf = GetModeConf(ModeFast3)
	clientConf.FECConf.DataShard = 0
	clientConf.FECConf.ParityShard = 0
	sinkBenchmark(b, 1048576, serverConf, clientConf)
}

func BenchmarkSink_Comparison(b *testing.B) {
	b.Run("Sink", func(b *testing.B) {
		b.Run("4K", func(b *testing.B) {
			b.Run("Base", baseSinkBenchmarkRunner(4096))
			b.Run("Perf", perfSinkBenchmarkRunner(4096))
		})

		b.Run("64K", func(b *testing.B) {
			b.Run("Base", baseSinkBenchmarkRunner(65536))
			b.Run("Perf", perfSinkBenchmarkRunner(65536))
		})

		b.Run("512K", func(b *testing.B) {
			b.Run("Base", baseSinkBenchmarkRunner(524288))
			b.Run("Perf", perfSinkBenchmarkRunner(524288))
		})

		b.Run("1M", func(b *testing.B) {
			b.Run("Base", baseSinkBenchmarkRunner(1048576))
			b.Run("Perf", perfSinkBenchmarkRunner(1048576))
		})
	})
}

// echo
func echoBenchmark(b *testing.B, nbytes int, serverConf *KcpConfig, clientConf *KcpConfig) {
	saddr := getTestAddr()
	server, err := NewServer(saddr, serverConf, &testEchoHandler{})
	require.NoError(b, err)
	defer server.Close()

	b.ReportAllocs()

	client, err := NewClient(saddr, clientConf)
	require.NoError(b, err)
	defer client.Close()

	// sender
	buf := make([]byte, nbytes)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// send packet
		if _, err := client.Write(buf); err != nil {
			b.Fatalf("failed to write to client: %v", err)
		}

		// receive packet
		nrecv := 0
		for {
			n, err := client.Read(buf)
			if err != nil {
				b.Fatalf("failed to read from client: %v", err)
			} else {
				nrecv += n
				if nrecv == nbytes {
					break
				}
			}
		}
	}

	b.SetBytes(int64(nbytes))
}

func baseEchoBenchmarkRunner(nbytes int) func(*testing.B) {
	defaultConf := DefaultConfig()
	return func(b *testing.B) {
		echoBenchmark(b, nbytes, defaultConf, defaultConf)
	}
}

func perfEchoBenchmarkRunner(nbytes int) func(*testing.B) {
	serverConf := DefaultConfig()
	serverConf.SockBuf = 4 * 1024 * 1024
	serverConf.MTU = 1400
	serverConf.DSCP = 46
	serverConf.ModeConf = GetModeConf(ModeFast3)
	serverConf.SndWnd = 4096
	serverConf.RcvWnd = 4096
	serverConf.Crypt = "null"
	serverConf.FECConf.DataShard = 0
	serverConf.FECConf.ParityShard = 0

	clientConf := DefaultConfig()
	clientConf.Crypt = "null"
	clientConf.SndWnd = 1024
	clientConf.RcvWnd = 1024
	clientConf.SockBuf = 16 * 1024 * 1024
	clientConf.MTU = 1400
	clientConf.ModeConf = GetModeConf(ModeFast3)
	clientConf.FECConf.DataShard = 0
	clientConf.FECConf.ParityShard = 0
	return func(b *testing.B) {
		echoBenchmark(b, nbytes, serverConf, clientConf)
	}
}

func BenchmarkEcho_Base_Speed4K(b *testing.B) {
	baseEchoBenchmarkRunner(4096)(b)
}

func BenchmarkEchoPerf_Speed4K(b *testing.B) {
	perfEchoBenchmarkRunner(4096)(b)
}

func BenchmarkEcho_Base_Speed1M(b *testing.B) {
	baseEchoBenchmarkRunner(1048576)(b)
}

func BenchmarkEcho_Perf_Speed1M(b *testing.B) {
	serverConf := DefaultConfig()
	serverConf.SockBuf = 4 * 1024 * 1024
	serverConf.MTU = 1200
	serverConf.DSCP = 46
	serverConf.ModeConf = GetModeConf(ModeFast3)
	serverConf.SndWnd = 4096
	serverConf.RcvWnd = 4096
	serverConf.Crypt = "null"
	serverConf.FECConf.DataShard = 0
	serverConf.FECConf.ParityShard = 0

	clientConf := DefaultConfig()
	clientConf.Crypt = "null"
	clientConf.SndWnd = 1024
	clientConf.RcvWnd = 1024
	clientConf.SockBuf = 16 * 1024 * 1024
	clientConf.MTU = 1200
	clientConf.ModeConf = GetModeConf(ModeFast3)
	clientConf.FECConf.DataShard = 0
	clientConf.FECConf.ParityShard = 0
	echoBenchmark(b, 1048576, serverConf, clientConf)
}

func BenchmarkEcho_Comparison(b *testing.B) {
	b.Run("Echo", func(b *testing.B) {
		b.Run("4K", func(b *testing.B) {
			b.Run("Base", baseEchoBenchmarkRunner(4096))
			b.Run("Perf", perfEchoBenchmarkRunner(4096))
		})

		b.Run("64K", func(b *testing.B) {
			b.Run("Base", baseEchoBenchmarkRunner(65536))
			b.Run("Perf", perfEchoBenchmarkRunner(65536))
		})

		b.Run("512K", func(b *testing.B) {
			b.Run("Base", baseEchoBenchmarkRunner(524288))
			b.Run("Perf", perfEchoBenchmarkRunner(524288))
		})

		b.Run("1M", func(b *testing.B) {
			b.Run("Base", baseEchoBenchmarkRunner(1048576))
			b.Run("Perf", perfEchoBenchmarkRunner(1048576))
		})
	})
}

// sink
func sinkWithLossyConnBenchmark(b *testing.B, loss float64, delay int, nbytes int, serverConf *KcpConfig, clientConf *KcpConfig) {
	clientLC, err := lossyconn.NewLossyConn(loss, delay)
	if err != nil {
		b.Fatalf("failed to create lossyconn: %v", err)
	}

	serverLC, err := lossyconn.NewLossyConn(loss, delay)
	if err != nil {
		b.Fatalf("failed to create lossyconn: %v", err)
	}

	server, err := NewServerWithConn(serverLC, serverConf, &testSinkHandler{})
	require.NoError(b, err)
	defer server.Close()

	b.ReportAllocs()

	client, err := NewClientWithConn(clientLC, serverLC.LocalAddr(), clientConf)
	require.NoError(b, err)
	defer client.Close()

	// sender
	buf := make([]byte, nbytes)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.Write(buf); err != nil {
			b.Fatalf("failed to write to client: %v", err) // stop benchmark
		}
	}

	b.SetBytes(int64(nbytes))
}

func baseSinkLossyConnBenchmarkRunner(nbytes int, loss float64, delay int) func(*testing.B) {
	defaultConf := DefaultConfig()
	return func(b *testing.B) {
		sinkWithLossyConnBenchmark(b, loss, delay, nbytes, defaultConf, defaultConf)
	}
}

func perfSinkLossyConnBenchmarkRunner(nbytes int, loss float64, delay int) func(*testing.B) {
	serverConf := DefaultConfig()
	serverConf.SockBuf = 4 * 1024 * 1024
	serverConf.MTU = 1400
	serverConf.DSCP = 46
	serverConf.ModeConf = GetModeConf(ModeFast3)
	serverConf.SndWnd = 4096
	serverConf.RcvWnd = 4096
	serverConf.Crypt = "null"
	serverConf.FECConf.DataShard = 0
	serverConf.FECConf.ParityShard = 0

	clientConf := DefaultConfig()
	clientConf.Crypt = "null"
	clientConf.SndWnd = 1024
	clientConf.RcvWnd = 1024
	clientConf.SockBuf = 16 * 1024 * 1024
	clientConf.MTU = 1400
	clientConf.ModeConf = GetModeConf(ModeFast3)
	clientConf.FECConf.DataShard = 0
	clientConf.FECConf.ParityShard = 0
	return func(b *testing.B) {
		sinkWithLossyConnBenchmark(b, loss, delay, nbytes, serverConf, clientConf)
	}
}

func BenchmarkSinkWithLossyConn_Comparison(b *testing.B) {
	// testing loss rate 10%, rtt 200ms -> 2.x MB/s

	loss := 0.1
	delay := 100

	b.Run("Sink_LossyConn", func(b *testing.B) {
		b.Run("4K", func(b *testing.B) {
			b.Run("Base", baseSinkLossyConnBenchmarkRunner(4096, loss, delay))
			b.Run("Perf", perfSinkLossyConnBenchmarkRunner(4096, loss, delay))
		})

		b.Run("64K", func(b *testing.B) {
			b.Run("Base", baseSinkLossyConnBenchmarkRunner(65536, loss, delay))
			b.Run("Perf", perfSinkLossyConnBenchmarkRunner(65536, loss, delay))
		})

		b.Run("512K", func(b *testing.B) {
			b.Run("Base", baseSinkLossyConnBenchmarkRunner(524288, loss, delay))
			b.Run("Perf", perfSinkLossyConnBenchmarkRunner(524288, loss, delay))
		})

		b.Run("1M", func(b *testing.B) {
			b.Run("Base", baseSinkLossyConnBenchmarkRunner(1048576, loss, delay))
			b.Run("Perf", perfSinkLossyConnBenchmarkRunner(1048576, loss, delay))
		})
	})
}

func BenchmarkSinkWithLossyConn_Comparison_Seed(b *testing.B) {
	// testing loss rate 10%, rtt 200ms -> 2.x MB/s

	b.Run("Sink_LossyConn", func(b *testing.B) {
		b.Run("1M_real_no_loss", func(b *testing.B) {
			b.Run("Limit", baseSinkLossyConnBenchmarkRunner(1048576, 0.0, 100))
			b.Run("No-Limit", perfSinkLossyConnBenchmarkRunner(1048576, 0.0, 100))
		})

		b.Run("1M_real_has_loss_0.1_100", func(b *testing.B) {
			b.Run("Limit", baseSinkLossyConnBenchmarkRunner(1048576, 0.1, 50))
			b.Run("No-Limit", perfSinkLossyConnBenchmarkRunner(1048576, 0.1, 50))
		})

		b.Run("1M_real_has_loss_0.1_200", func(b *testing.B) {
			b.Run("Limit", baseSinkLossyConnBenchmarkRunner(1048576, 0.1, 100))
			b.Run("No-Limit", perfSinkLossyConnBenchmarkRunner(1048576, 0.1, 100))
		})

		b.Run("1M_real_has_loss_0.2_200", func(b *testing.B) {
			b.Run("Limit", baseSinkLossyConnBenchmarkRunner(1048576, 0.2, 100))
			b.Run("No-Limit", perfSinkLossyConnBenchmarkRunner(1048576, 0.2, 100))
		})

		b.Run("1M_real_has_loss_0.3_200", func(b *testing.B) {
			b.Run("Limit", baseSinkLossyConnBenchmarkRunner(1048576, 0.3, 100))
			b.Run("No-Limit", perfSinkLossyConnBenchmarkRunner(1048576, 0.3, 100))
		})
	})
}