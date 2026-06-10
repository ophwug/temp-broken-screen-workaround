package main

import (
	"fmt"
	"strings"
)

func printTextReport(report *RunReport) {
	reachable := 0
	for _, device := range report.DeviceReports {
		if device.SSHReachable {
			reachable++
		}
	}

	fmt.Fprintln(output, "SSH probe complete.")
	fmt.Fprintf(output, "Targets probed: %d\n", len(report.Targets))
	fmt.Fprintf(output, "SSH reachable devices: %d\n\n", reachable)

	if reachable == 0 {
		fmt.Fprintln(output, "No comma/openpilot device accepted SSH with the embedded comma key.")
		fmt.Fprintln(output, "Make sure the device is powered on, on the same network, and still exposes SSH as user comma.")
		fmt.Fprintln(output)
		if report.LogPath != "" {
			fmt.Fprintf(output, "Install log written to: %s\n", report.LogPath)
		}
		fmt.Fprintln(output, "If the screen is usable enough to show network details, pass --ip <device-ip> and try again.")
	}
}

func indentBlock(s, prefix string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i := range lines {
		lines[i] = prefix + strings.TrimSpace(lines[i])
	}
	return strings.Join(lines, "\n")
}
