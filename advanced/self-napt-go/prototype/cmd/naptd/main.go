package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"nat-handson/advanced/self-napt-go/prototype/internal/forwarder"
	"nat-handson/advanced/self-napt-go/prototype/internal/nat"
	"nat-handson/advanced/self-napt-go/prototype/internal/sys"
	"nat-handson/advanced/self-napt-go/prototype/internal/tun"
)

type config struct {
	LANIF     string
	WANIF     string
	LANCIDR   string
	TUNName   string
	WANIP     string
	PortRange string
	LogLevel  string
	EnableIpt bool
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.LANIF, "lan-if", "eth0", "LAN interface name")
	flag.StringVar(&cfg.WANIF, "wan-if", "eth1", "WAN interface name")
	flag.StringVar(&cfg.LANCIDR, "lan-cidr", "192.168.20.0/24", "LAN subnet CIDR")
	flag.StringVar(&cfg.TUNName, "tun-name", "tun0", "TUN device name")
	flag.StringVar(&cfg.WANIP, "wan-ip", "", "WAN side source address for NAPT")
	flag.StringVar(&cfg.PortRange, "port-range", "40000-49999", "translated source port range")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "log level: debug/info")
	flag.BoolVar(&cfg.EnableIpt, "enable-iptables", true, "enable ip_forward and iptables NAT/FORWARD rules")
	flag.Parse()

	wanIP := net.ParseIP(cfg.WANIP)
	if wanIP == nil {
		log.Fatal("invalid --wan-ip")
	}

	minPort, maxPort, err := parsePortRange(cfg.PortRange)
	if err != nil {
		log.Fatalf("invalid --port-range: %v", err)
	}

	logger := log.New(os.Stdout, "", 0)
	table := nat.NewTable(time.Now)
	alloc := nat.NewPortAllocator(uint16(minPort), uint16(maxPort))
	fwd := forwarder.New(table, alloc, wanIP, logger)

	printJSONLog(logger, "startup", map[string]any{
		"lan_if":     cfg.LANIF,
		"wan_if":     cfg.WANIF,
		"lan_cidr":   cfg.LANCIDR,
		"tun_name":   cfg.TUNName,
		"wan_ip":     cfg.WANIP,
		"port_range": cfg.PortRange,
	})

	if cfg.TUNName != "" {
		if err := tun.EnsureUp(cfg.TUNName); err != nil {
			log.Fatalf("tun setup failed: %v", err)
		}
		printJSONLog(logger, "tun_ready", map[string]any{"name": cfg.TUNName})
	}
	if cfg.EnableIpt {
		if err := sys.ConfigureIPForwardAndNAT(cfg.LANIF, cfg.WANIF, cfg.LANCIDR, logger); err != nil {
			log.Fatalf("iptables setup failed: %v", err)
		}
		printJSONLog(logger, "kernel_nat_ready", map[string]any{
			"lan_if":   cfg.LANIF,
			"wan_if":   cfg.WANIF,
			"lan_cidr": cfg.LANCIDR,
		})
	}

	stopGC := make(chan struct{})
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				expired := table.Sweep(30*time.Second, 5*time.Minute)
				for _, m := range expired {
					alloc.Release(m.TranslatedSrcPort)
					printJSONLog(logger, "mapping_expired", map[string]any{
						"translated_src_port": m.TranslatedSrcPort,
						"state":               m.State.String(),
					})
				}
			case <-stopGC:
				return
			}
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	close(stopGC)
	fwd.Close()
	printJSONLog(logger, "shutdown", nil)
}

func parsePortRange(portRange string) (int, int, error) {
	parts := strings.Split(portRange, "-")
	if len(parts) != 2 {
		return 0, 0, os.ErrInvalid
	}
	min, err := parseUint(parts[0])
	if err != nil {
		return 0, 0, err
	}
	max, err := parseUint(parts[1])
	if err != nil {
		return 0, 0, err
	}
	if min < 1 || max > 65535 || min > max {
		return 0, 0, os.ErrInvalid
	}
	return min, max, nil
}

func parseUint(raw string) (int, error) {
	v, err := net.LookupPort("tcp", raw)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func printJSONLog(l *log.Logger, event string, kv map[string]any) {
	record := map[string]any{
		"ts":    time.Now().UTC().Format(time.RFC3339),
		"event": event,
	}
	for k, v := range kv {
		record[k] = v
	}
	b, err := json.Marshal(record)
	if err != nil {
		l.Printf("{\"event\":\"log_error\",\"message\":%q}", err.Error())
		return
	}
	l.Println(string(b))
}
