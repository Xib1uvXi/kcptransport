package xkcp

import (
	"crypto/sha1"

	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
)

const salt = "kcptransport"

func GetBlockCrypt(seed string, crypt string) kcp.BlockCrypt {
	pass := pbkdf2.Key([]byte(seed), []byte(salt), 4096, 32, sha1.New)

	var block kcp.BlockCrypt

	switch crypt {
	case "null":
		block = nil
	case "sm4":
		block, _ = kcp.NewSM4BlockCrypt(pass[:16])
	case "tea":
		block, _ = kcp.NewTEABlockCrypt(pass[:16])
	case "xor":
		block, _ = kcp.NewSimpleXORBlockCrypt(pass)
	case "none":
		block, _ = kcp.NewNoneBlockCrypt(pass)
	case "aes-128":
		block, _ = kcp.NewAESBlockCrypt(pass[:16])
	case "aes-192":
		block, _ = kcp.NewAESBlockCrypt(pass[:24])
	case "blowfish":
		block, _ = kcp.NewBlowfishBlockCrypt(pass)
	case "twofish":
		block, _ = kcp.NewTwofishBlockCrypt(pass)
	case "cast5":
		block, _ = kcp.NewCast5BlockCrypt(pass[:16])
	case "3des":
		block, _ = kcp.NewTripleDESBlockCrypt(pass[:24])
	case "xtea":
		block, _ = kcp.NewXTEABlockCrypt(pass[:16])
	case "salsa20":
		block, _ = kcp.NewSalsa20BlockCrypt(pass)
	default:
		block, _ = kcp.NewAESBlockCrypt(pass)
	}

	return block
}
