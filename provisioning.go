package main

import (
	"encoding/hex"
)

type provisioningData struct {
	PSKey [8]string
}

func makeProvisioningData(conf clientConfig) (res provisioningData) {
	for i := 0; i < 8; i++ {
		v := conf.PSKey[i*4 : (i+1)*4]
		res.PSKey[i] = hex.EncodeToString(v)
	}
	return
}
