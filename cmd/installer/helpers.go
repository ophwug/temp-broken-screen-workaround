package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

func chooseSingleTarget(reports []DeviceReport, reader *bufio.Reader, allowPrompt bool, actionName string) (DeviceReport, error) {
	reachable := make([]DeviceReport, 0, len(reports))
	for _, report := range reports {
		if report.SSHReachable {
			reachable = append(reachable, report)
		}
	}
	if len(reachable) == 0 {
		return DeviceReport{}, fmt.Errorf("no SSH-reachable device found for %s", actionName)
	}
	if len(reachable) == 1 {
		return reachable[0], nil
	}
	if !allowPrompt {
		return DeviceReport{}, fmt.Errorf("multiple SSH-reachable devices found; pass --ip <device-ip> for %s", actionName)
	}

	for {
		fmt.Fprintf(output, "Choose the device to %s:\n", actionName)
		for i, report := range reachable {
			fmt.Fprintf(output, "  %d. %s\n", i+1, report.IP)
		}
		fmt.Fprintf(output, "Select 1-%d: ", len(reachable))
		choice, _ := reader.ReadString('\n')
		idx, err := strconv.Atoi(strings.TrimSpace(choice))
		if err == nil && idx >= 1 && idx <= len(reachable) {
			fmt.Fprintf(output, "Target selected for %s: %s\n\n", actionName, reachable[idx-1].IP)
			return reachable[idx-1], nil
		}
		fmt.Fprintln(output, "Please choose one of the listed devices.")
		fmt.Fprintln(output)
	}
}

func promptYesNo(reader *bufio.Reader, question string, defaultYes bool) (bool, error) {
	suffix := " [Y/n]: "
	if !defaultYes {
		suffix = " [y/N]: "
	}
	for {
		fmt.Fprint(output, question+suffix)
		answer, _ := reader.ReadString('\n')
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer == "" {
			return defaultYes, nil
		}
		switch answer {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintln(output, "Please answer yes or no.")
		}
	}
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
