//go:build !linux

package tun

import "fmt"

func EnsureUp(name string) error {
	_ = name
	return fmt.Errorf("tun setup is only supported on linux")
}
