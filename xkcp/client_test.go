package xkcp

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_genConvid(t *testing.T) {
	convid := genConvid()
	// t.Logf("Generated convid: %d", convid)
	require.True(t, convid > 0)
}

func TestNewClient(t *testing.T) {
	saddr := getTestAddr()
	server, err := NewServer(saddr, DefaultConfig(), &testSinkHandler{})
	require.NoError(t, err)
	defer server.Close()

	client, err := NewClient(saddr, DefaultConfig())
	require.NoError(t, err)
	defer client.Close()

	for i := 0; i < 10; i++ {
		n, err := client.Write([]byte("hello"))
		require.NoError(t, err)
		require.Equal(t, 5, n)
	}
}

func TestNewClientWithLocal(t *testing.T) {
	saddr := getTestAddr()
	conf := DefaultConfig()
	conf.DSCP = 46
	server, err := NewServer(saddr, conf, &testSinkHandler{})
	require.NoError(t, err)
	defer server.Close()

	client, err := NewClientWithLocal(getTestAddr(), saddr, conf)
	require.NoError(t, err)
	defer client.Close()

	for i := 0; i < 10; i++ {
		n, err := client.Write([]byte("hello"))
		require.NoError(t, err)
		require.Equal(t, 5, n)
	}
}

func TestNewClientWithConn(t *testing.T) {
	saddr := getTestAddr()
	conf := DefaultConfig()
	conf.DSCP = 46
	server, err := NewServer(saddr, conf, &testSinkHandler{})
	require.NoError(t, err)
	defer server.Close()

	conn, err := net.ListenPacket("udp", getTestAddr())
	require.NoError(t, err)

	client, err := NewClientWithConn(conn, server.lis.Addr(), conf)
	require.NoError(t, err)
	defer client.Close()

	for i := 0; i < 10; i++ {
		n, err := client.Write([]byte("hello"))
		require.NoError(t, err)
		require.Equal(t, 5, n)
	}
}
