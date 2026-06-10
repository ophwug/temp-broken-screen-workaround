package main

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"
)

type DeviceReport struct {
	IP           string `json:"ip"`
	SSHReachable bool   `json:"ssh_reachable"`
	SSHError     string `json:"ssh_error,omitempty"`
}

func scanSSHReachability(ctx context.Context, targets []net.IP, parallel int, timeout time.Duration, key []byte) []DeviceReport {
	jobs := make(chan net.IP)
	results := make(chan DeviceReport, len(targets))

	var wg sync.WaitGroup
	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range jobs {
				results <- probeTarget(ctx, ip, timeout, key)
			}
		}()
	}

	for _, ip := range targets {
		jobs <- ip
	}
	close(jobs)
	wg.Wait()
	close(results)

	out := make([]DeviceReport, 0, len(targets))
	for res := range results {
		out = append(out, res)
	}
	return out
}

func probeTarget(ctx context.Context, ip net.IP, timeout time.Duration, key []byte) DeviceReport {
	report := DeviceReport{IP: ip.String()}
	client, err := connectSSH(ctx, ip, timeout, key)
	if err != nil {
		report.SSHError = compactError(err)
		return report
	}
	_ = client.Close()
	report.SSHReachable = true
	return report
}

func compactError(err error) string {
	if err == nil {
		return ""
	}
	text := strings.TrimSpace(err.Error())
	text = strings.ReplaceAll(text, "\n", "; ")
	if len(text) > 220 {
		return text[:217] + "..."
	}
	return text
}
