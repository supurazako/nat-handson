//go:build !linux

package sys

import (
	"fmt"
	"log"
)

func ConfigureIPForwardAndNAT(lanIF, wanIF, lanCIDR string, logger *log.Logger) error {
	_ = lanIF
	_ = wanIF
	_ = lanCIDR
	_ = logger
	return fmt.Errorf("ConfigureIPForwardAndNAT is only supported on linux")
}
