package xkcp

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetModeConf(t *testing.T) {
	type args struct {
		mode string
	}
	tests := []struct {
		name string
		args args
		want *ModeConf
	}{
		{name: "TestNormal", args: args{mode: ModeNormal}, want: GetModeConf(ModeNormal)},
		{name: "TestFast", args: args{mode: ModeFast}, want: GetModeConf(ModeFast)},
		{name: "TestFast2", args: args{mode: ModeFast2}, want: GetModeConf(ModeFast2)},
		{name: "TestFast3", args: args{mode: ModeFast3}, want: GetModeConf(ModeFast3)},
		{name: "TestDefault", args: args{mode: ""}, want: GetModeConf(ModeNormal)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetModeConf(tt.args.mode); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetModeConf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	conf := DefaultConfig()
	require.NotNil(t, conf)

	require.Equal(t, 1200, conf.MTU)
	require.Equal(t, 1024, conf.SndWnd)
	require.Equal(t, 1024*4, conf.RcvWnd)
	require.False(t, conf.AckNodelay)
	require.Equal(t, GetModeConf(ModeNormal), conf.ModeConf)
}
