package tun

import (
	"fmt"
	"os/exec"
	"strings"
)

func EnsureUp(name string) error {
	if err := ensureExists(name); err != nil {
		return err
	}
	if err := run("ip", "link", "set", "dev", name, "up"); err != nil {
		return fmt.Errorf("set tun up: %w", err)
	}
	return nil
}

func ensureExists(name string) error {
	cmd := exec.Command("ip", "tuntap", "add", "dev", name, "mode", "tun")
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	msg := string(out)
	if strings.Contains(msg, "File exists") {
		return nil
	}
	return fmt.Errorf("create tun: %w (%s)", err, msg)
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v: %w (%s)", name, args, err, string(out))
	}
	return nil
}
