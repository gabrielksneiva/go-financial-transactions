package utils

import (
	"encoding/hex"

	"github.com/btcsuite/btcutil/base58"
)

func Base58ToHex(address string) string {
	return hex.EncodeToString(base58.Decode(address))
}
