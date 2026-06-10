package main

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

const postInstallMonitorInterval = 15 * time.Second

func monitorPostInstall(ctx context.Context, cfg Config, ip net.IP) {
	if cfg.MonitorDuration <= 0 {
		fmt.Fprintln(output, "Post-install monitor disabled.")
		return
	}

	fmt.Fprintf(output, "\nPost-install monitor (%s):\n", cfg.MonitorDuration)
	fmt.Fprintln(output, "  Watching SSH reachability, AGNOS version state, updater/launcher processes, and recent launch logs.")

	deadline := time.Now().Add(cfg.MonitorDuration)
	lastPrinted := ""
	lastPrintAt := time.Time{}

	for time.Now().Before(deadline) {
		status, done := pollPostInstallStatus(ctx, cfg, ip)
		now := time.Now()
		if status != lastPrinted || now.Sub(lastPrintAt) >= time.Minute {
			fmt.Fprintf(output, "[%s]\n%s\n", now.Format("15:04:05"), indentBlock(status, "  "))
			lastPrinted = status
			lastPrintAt = now
		}
		if done {
			fmt.Fprintln(output, "Post-install monitor complete: AGNOS version matches and manager is visible over SSH.")
			return
		}

		sleepFor := postInstallMonitorInterval
		if remaining := time.Until(deadline); remaining < sleepFor {
			sleepFor = remaining
		}
		select {
		case <-ctx.Done():
			fmt.Fprintln(output, "Post-install monitor stopped: context canceled.")
			return
		case <-time.After(sleepFor):
		}
	}

	fmt.Fprintln(output, "Post-install monitor ended. If SSH/network access is down, the device may still be updating or rebooting.")
}

func pollPostInstallStatus(ctx context.Context, cfg Config, ip net.IP) (string, bool) {
	client, err := connectSSH(ctx, ip, cfg.Timeout, privateKey)
	if err != nil {
		return "SSH_UNREACHABLE\t" + compactError(err), false
	}
	defer client.Close()

	out, err := executeCommand(client, postInstallStatusCommand(), 8*time.Second)
	status := strings.TrimSpace(out)
	if err != nil {
		if status != "" {
			status += "\n"
		}
		status += "STATUS_ERROR\t" + compactError(err)
	}
	if status == "" {
		status = "STATUS_ERROR\tempty status response"
	}

	done := strings.Contains(status, "AGNOS_STATUS\tmatch") && strings.Contains(status, "manager.py")
	return status, done
}

func postInstallStatusCommand() string {
	return `set +e
current="$(cat /VERSION 2>/dev/null || true)"
expected="$(bash -lc 'cd /data/openpilot && unset AGNOS_VERSION && source launch_env.sh >/dev/null 2>&1 && printf "%s" "$AGNOS_VERSION"' 2>/dev/null || true)"
if [ -n "$current" ] && [ -n "$expected" ] && [ "$current" = "$expected" ]; then
  agnos_status="match"
elif [ -n "$current" ] && [ -n "$expected" ]; then
  agnos_status="update-needed"
else
  agnos_status="unknown"
fi
procs="$(ps -eo pid,comm,args 2>/dev/null | grep -Ei 'agnos|updater|launch_openpilot|launch_chffrplus|manager.py|updated.py|comma.sh' | grep -v grep | head -8 | awk '{$1=$1; print}' | tr '\n' ';')"
logs="$(tail -120 /tmp/launch_log 2>/dev/null | grep -Ei 'agnos|updater|version|reboot|verify|error|fail' | tail -6 | tr '\n' ';')"
printf 'SSH\tok\n'
printf 'CURRENT_AGNOS\t%s\n' "${current:-unknown}"
printf 'EXPECTED_AGNOS\t%s\n' "${expected:-unknown}"
printf 'AGNOS_STATUS\t%s\n' "$agnos_status"
printf 'PROCESSES\t%s\n' "${procs:-none}"
printf 'RECENT_LAUNCH_LOG\t%s\n' "${logs:-none}"`
}
