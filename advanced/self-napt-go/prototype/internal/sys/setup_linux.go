package sys

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func ConfigureIPForwardAndNAT(lanIF, wanIF, lanCIDR string, logger *log.Logger) error {
	if err := ensureIPForward(logger); err != nil {
		return fmt.Errorf("enable ip_forward: %w", err)
	}

	// nat POSTROUTING MASQUERADE
	if err := ensureRule(logger, true, "nat", "POSTROUTING", "-s", lanCIDR, "-o", wanIF, "-j", "MASQUERADE"); err != nil {
		return fmt.Errorf("ensure MASQUERADE: %w", err)
	}
	// filter FORWARD LAN->WAN
	if err := ensureRule(logger, false, "", "FORWARD", "-i", lanIF, "-o", wanIF, "-j", "ACCEPT"); err != nil {
		return fmt.Errorf("ensure FORWARD lan->wan: %w", err)
	}
	// filter FORWARD WAN->LAN established
	if err := ensureRule(logger, false, "", "FORWARD", "-i", wanIF, "-o", lanIF, "-m", "conntrack", "--ctstate", "ESTABLISHED,RELATED", "-j", "ACCEPT"); err != nil {
		return fmt.Errorf("ensure FORWARD wan->lan established: %w", err)
	}
	return nil
}

func ensureIPForward(logger *log.Logger) error {
	b, err := os.ReadFile("/proc/sys/net/ipv4/ip_forward")
	if err == nil && strings.TrimSpace(string(b)) == "1" {
		return nil
	}
	if err := run(logger, "sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
		b2, readErr := os.ReadFile("/proc/sys/net/ipv4/ip_forward")
		if readErr == nil && strings.TrimSpace(string(b2)) == "1" {
			return nil
		}
		return err
	}
	return nil
}

func ensureRule(logger *log.Logger, withTable bool, table, chain string, ruleArgs ...string) error {
	checkArgs := []string{}
	addArgs := []string{}
	if withTable {
		checkArgs = append(checkArgs, "-t", table)
		addArgs = append(addArgs, "-t", table)
	}
	checkArgs = append(checkArgs, "-C", chain)
	checkArgs = append(checkArgs, ruleArgs...)
	if exec.Command("iptables", checkArgs...).Run() == nil {
		return nil
	}

	addArgs = append(addArgs, "-A", chain)
	addArgs = append(addArgs, ruleArgs...)
	return run(logger, "iptables", addArgs...)
}

func run(logger *log.Logger, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if logger != nil {
		logger.Printf("{\"event\":\"sys_cmd\",\"cmd\":%q,\"output\":%q}", name+" "+joinArgs(args), string(out))
	}
	if err != nil {
		return fmt.Errorf("%s %v: %w (%s)", name, args, err, string(out))
	}
	return nil
}

func joinArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	s := args[0]
	for i := 1; i < len(args); i++ {
		s += " " + args[i]
	}
	return s
}
