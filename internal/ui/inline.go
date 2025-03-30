package ui

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/miklosn/cmdperf/internal/benchmark"
	"github.com/miklosn/cmdperf/internal/command"
	"github.com/miklosn/cmdperf/internal/ui/colorscheme"
)

type InlineUI struct {
	commands    []*benchmark.CommandStats
	totalRuns   int
	duration    time.Duration
	startTime   time.Time
	colorScheme *colorscheme.Scheme

	mu                  sync.Mutex
	lastUpdate          time.Time
	finished            bool
	cancelled           bool
	lastLines           int
	lastProgressPercent float64
	lastEta             time.Duration
}

func NewInlineUI(totalRuns int) *InlineUI {
	return &InlineUI{
		commands:   make([]*benchmark.CommandStats, 0),
		totalRuns:  totalRuns,
		startTime:  time.Now(),
		lastUpdate: time.Now(),
	}
}

func (ui *InlineUI) Update(stats []*benchmark.CommandStats, complete bool) {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	// If we're already finished, don't update again
	if ui.finished {
		return
	}

	// Only update at most once every 250ms to avoid too much output
	if time.Since(ui.lastUpdate) < 250*time.Millisecond && !complete {
		return
	}

	ui.commands = stats
	ui.lastUpdate = time.Now()

	// Set finished flag only if this is the final update
	if complete {
		ui.finished = true
	}

	// Render the UI
	ui.render()
}

func (ui *InlineUI) EventHandler(event interface{}) {
	// We don't need to handle individual events in this UI
}

func (ui *InlineUI) Cancel() {
	ui.mu.Lock()
	// Check if we're already finished to avoid duplicate output
	if ui.finished {
		ui.mu.Unlock()
		return
	}
	ui.cancelled = true
	ui.finished = true
	ui.mu.Unlock()

	// Trigger a final update with the current state
	ui.Update(ui.commands, false) // Pass false for complete
}

func (ui *InlineUI) render() {
	// Calculate overall progress
	totalCompleted := 0
	totalExpected := 0

	for _, cmd := range ui.commands {
		if cmd != nil {
			totalCompleted += cmd.TotalRuns
			totalExpected += ui.totalRuns
		}
	}

	// Calculate progress percentage and ETA
	progressPercent := 0.0
	var eta time.Duration
	var maxEta time.Duration

	if ui.duration > 0 {
		// For duration-based benchmarks, calculate progress based on elapsed time
		elapsed := time.Since(ui.startTime)
		if elapsed >= ui.duration {
			progressPercent = 1.0
		} else {
			progressPercent = float64(elapsed) / float64(ui.duration)
		}

		// For duration-based benchmarks, ETA is just the remaining time
		if progressPercent > 0 && progressPercent < 1.0 {
			remainingTime := ui.duration - elapsed
			eta = remainingTime
		}
	} else if totalExpected > 0 {
		// For iteration-based benchmarks, use the existing approach
		progressPercent = float64(totalCompleted) / float64(totalExpected)

		// If no commands have completed yet, estimate progress based on time
		if totalCompleted == 0 {
			// Get the timeout from the first command
			var cmdTimeout time.Duration
			if len(ui.commands) > 0 && ui.commands[0] != nil && ui.commands[0].Command != nil {
				cmdTimeout = ui.commands[0].Command.Timeout
			} else {
				cmdTimeout = 60 * time.Second // Default timeout
			}

			// Estimate progress based on elapsed time vs timeout
			elapsed := time.Since(ui.startTime)
			if elapsed < cmdTimeout {
				// Estimate progress as a fraction of the timeout
				// This gives a sense of progress even when no commands have completed
				progressPercent = math.Min(0.9, float64(elapsed)/float64(cmdTimeout))
			}
		}

		// Store the progress percentage if not cancelled
		if !ui.cancelled {
			ui.lastProgressPercent = progressPercent

			// Calculate ETA for each command independently
			commandEtas := make(map[*command.Command]time.Duration)

			for _, cmd := range ui.commands {
				if cmd != nil && cmd.Command != nil {
					// Skip completed commands
					if cmd.TotalRuns >= ui.totalRuns {
						continue
					}

					// Calculate progress for this command
					cmdProgress := float64(cmd.TotalRuns) / float64(ui.totalRuns)

					// Only calculate if we have some results
					if cmd.SuccessfulRuns > 0 && cmdProgress > 0 {
						// Calculate average time per run for this command
						avgRunTime := float64(cmd.Mean.Nanoseconds()) / 1e9 // in seconds

						// Calculate runs remaining for this command
						runsRemaining := ui.totalRuns - cmd.TotalRuns

						// Calculate total time remaining for this command
						// Account for parallelism
						effectiveParallelism := float64(cmd.Command.Parallelism)
						if effectiveParallelism <= 0 {
							effectiveParallelism = 1
						}

						// Time remaining = (avg time per run * runs remaining) / parallelism
						cmdEtaSeconds := (avgRunTime * float64(runsRemaining)) / effectiveParallelism
						cmdEta := time.Duration(cmdEtaSeconds * float64(time.Second))

						// Store this command's ETA
						commandEtas[cmd.Command] = cmdEta

						// Update max ETA if this is longer
						if cmdEta > maxEta {
							maxEta = cmdEta
						}
					}
				}
			}

			// Use the longest ETA for the overall display
			if maxEta > 0 {
				// Apply smoothing with previous estimate
				if ui.lastEta > 0 {
					// 80% previous, 20% new estimate for stability
					eta = time.Duration(0.8*float64(ui.lastEta) + 0.2*float64(maxEta))
				} else {
					eta = maxEta
				}
				ui.lastEta = eta

				// Round to a reasonable value
				if eta > time.Hour {
					eta = eta.Round(time.Minute)
				} else if eta > time.Minute {
					eta = eta.Round(time.Second)
				} else {
					eta = eta.Round(100 * time.Millisecond)
				}

				// Sanity check - cap unreasonably large ETAs
				if eta > 24*time.Hour {
					eta = 24 * time.Hour
				}
			}
		} else {
			// If cancelled, use the last stored progress percentage
			progressPercent = ui.lastProgressPercent
		}
	}

	// Build output
	var output strings.Builder

	// Use color scheme or fall back to default if not set
	scheme := ui.colorScheme
	if scheme == nil {
		scheme = colorscheme.Default()
	}

	headerColor := scheme.Header
	subheaderColor := scheme.Subheader
	commandColor := scheme.Command
	labelColor := scheme.Label
	valueColor := scheme.Value
	progressColor := scheme.Progress
	completedColor := scheme.Completed
	cancelledColor := scheme.Error

	// Get terminal width
	termWidth := getTerminalWidth()

	// Calculate dynamic column widths based on terminal width
	runsWidth := int(float64(termWidth) * 0.10)
	meanWidth := int(float64(termWidth) * 0.25)
	rangeWidth := int(float64(termWidth) * 0.25)
	throughputWidth := int(float64(termWidth) * 0.15)
	errorsWidth := int(float64(termWidth) * 0.15)

	// Ensure minimum widths
	runsWidth = max(runsWidth, 8)
	meanWidth = max(meanWidth, 15)
	rangeWidth = max(rangeWidth, 15)
	throughputWidth = max(throughputWidth, 10)
	errorsWidth = max(errorsWidth, 8)

	// Adjust if terminal is too narrow
	totalWidth := runsWidth + meanWidth + rangeWidth + throughputWidth + errorsWidth
	if totalWidth > termWidth-10 {
		// Scale down proportionally
		scale := float64(termWidth-10) / float64(totalWidth)
		runsWidth = int(float64(runsWidth) * scale)
		meanWidth = int(float64(meanWidth) * scale)
		rangeWidth = int(float64(rangeWidth) * scale)
		throughputWidth = int(float64(throughputWidth) * scale)
		errorsWidth = int(float64(errorsWidth) * scale)
	}

	// Always print the header
	output.WriteString(headerColor("✨ cmdperf - Command Performance Benchmarking ✨\n"))
	repeatCount := min(termWidth-5, 60)
	if repeatCount < 1 {
		repeatCount = 1
	}
	output.WriteString(strings.Repeat("─", repeatCount) + "\n\n")

	// For very narrow terminals, hide columns progressively
	if termWidth < 80 {
		// Hide range column first
		rangeWidth = 0

		// If still too narrow, hide throughput
		if termWidth < 60 {
			throughputWidth = 0

			// If extremely narrow, hide errors and reduce other columns
			if termWidth < 40 {
				errorsWidth = 0
				meanWidth = max(meanWidth-5, 10)
			}
		}
	}

	// Recalculate header format with potentially hidden columns
	headerFormat := ""
	if runsWidth > 0 {
		headerFormat += fmt.Sprintf("  %%-%ds ", runsWidth)
	}
	if meanWidth > 0 {
		headerFormat += fmt.Sprintf("%%-%ds ", meanWidth)
	}
	if rangeWidth > 0 {
		headerFormat += fmt.Sprintf("%%-%ds ", rangeWidth)
	}
	if throughputWidth > 0 {
		headerFormat += fmt.Sprintf("%%-%ds ", throughputWidth)
	}
	if errorsWidth > 0 {
		headerFormat += fmt.Sprintf("%%-%ds", errorsWidth)
	}
	headerFormat += "\n"

	// Build header arguments based on visible columns
	headerArgs := []interface{}{}
	if runsWidth > 0 {
		headerArgs = append(headerArgs, "Runs")
	}
	if meanWidth > 0 {
		headerArgs = append(headerArgs, "Mean ± StdDev")
	}
	if rangeWidth > 0 {
		headerArgs = append(headerArgs, "Range (min … max)")
	}
	if throughputWidth > 0 {
		headerArgs = append(headerArgs, "Throughput")
	}
	if errorsWidth > 0 {
		headerArgs = append(headerArgs, "Errors")
	}

	// Apply color to the headers
	headerLine := fmt.Sprintf(headerFormat, headerArgs...)
	output.WriteString(subheaderColor(headerLine))
	repeatCount = min(termWidth-5, 100)
	if repeatCount < 1 {
		repeatCount = 1
	}
	output.WriteString(strings.Repeat("━", repeatCount) + "\n")

	// Print command progress
	for _, cmd := range ui.commands {
		if cmd == nil || cmd.Command == nil {
			continue
		}

		// Print command on its own line with color
		cmdName := cmd.Command.Raw
		if ui.duration > 0 {
			output.WriteString(fmt.Sprintf("%s %s %s\n",
				labelColor("Command:"),
				commandColor(cmdName),
				subheaderColor(fmt.Sprintf("(running for %s)", ui.duration))))
		} else {
			output.WriteString(fmt.Sprintf("%s %s\n",
				labelColor("Command:"),
				commandColor(cmdName)))
		}

		var runs string
		if ui.duration > 0 {
			// When using duration mode, just show the total runs without a target
			runs = fmt.Sprintf("%d", cmd.TotalRuns)
		} else {
			// When using iterations mode, show progress as X/Y
			runs = fmt.Sprintf("%d/%d", cmd.TotalRuns, ui.totalRuns)
		}

		meanStdDev, timeRange, throughput := "-", "-", "-"
		if cmd.SuccessfulRuns > 0 {
			// Format the values first without color
			meanStr := formatDuration(cmd.Mean)
			stdDevStr := formatDuration(cmd.StdDev)
			minStr := formatDuration(cmd.Min)
			maxStr := formatDuration(cmd.Max)

			// Create the formatted strings without colors first
			meanStdDev = fmt.Sprintf("%s ± %s", meanStr, stdDevStr)
			timeRange = fmt.Sprintf("%s … %s", minStr, maxStr)
			throughput = formatThroughput(cmd.Throughput)
		}

		// Format error count
		errorStr := "-"
		if cmd.ErrorCount > 0 {
			errorStr = fmt.Sprintf("%d", cmd.ErrorCount)
		}

		// Add timeout information
		timeoutInfo := ""
		for exitCode, count := range cmd.ExitCodes {
			if exitCode != 0 {
				// For duration-based benchmarks, don't show exit codes from context cancellation
				if ui.duration > 0 && exitCode == -1 {
					// Skip showing these exit codes
					continue
				}
				timeoutInfo += fmt.Sprintf(" [Exit %d: %d]", exitCode, count)
			}
		}
		if cmd.ErrorCount > 0 {
			errorStr = fmt.Sprintf("%d%s", cmd.ErrorCount, timeoutInfo)
		}

		// Create a formatted line with proper spacing using dynamic widths
		lineArgs := []interface{}{}
		lineFormat := ""

		if runsWidth > 0 {
			lineFormat += fmt.Sprintf("  %%-%ds ", runsWidth)
			lineArgs = append(lineArgs, runs)
		}
		if meanWidth > 0 {
			lineFormat += fmt.Sprintf("%%-%ds ", meanWidth)
			lineArgs = append(lineArgs, meanStdDev)
		}
		if rangeWidth > 0 {
			lineFormat += fmt.Sprintf("%%-%ds ", rangeWidth)
			lineArgs = append(lineArgs, timeRange)
		}
		if throughputWidth > 0 {
			lineFormat += fmt.Sprintf("%%-%ds ", throughputWidth)
			lineArgs = append(lineArgs, throughput)
		}
		if errorsWidth > 0 {
			lineFormat += fmt.Sprintf("%%-%ds", errorsWidth)
			lineArgs = append(lineArgs, errorStr)
		}
		lineFormat += "\n"

		// Apply color to the entire line
		line := fmt.Sprintf(lineFormat, lineArgs...)
		output.WriteString(valueColor(line))
	}

	// Create a progress bar with dynamic width
	barWidth := min(30, termWidth/3)
	completedWidth := int(float64(barWidth) * progressPercent)

	// Add a pulsing indicator for slow commands
	var progressBar string
	if totalCompleted == 0 {
		// For commands that haven't completed yet, show estimated progress
		// with a pulsing indicator at the end of the progress bar
		var pulseChar string
		pulseIndex := int(time.Since(ui.startTime).Seconds()*2) % 3
		switch pulseIndex {
		case 0:
			pulseChar = "▒"
		case 1:
			pulseChar = "▓"
		case 2:
			pulseChar = "█"
		}

		// Show estimated progress with a pulsing indicator
		remainingWidth := barWidth - 1
		if remainingWidth < 0 {
			remainingWidth = 0
		}
		progressBar = fmt.Sprintf("[%s%s] ",
			pulseChar,
			strings.Repeat("░", remainingWidth),
		)
	} else {
		if completedWidth < 0 {
			completedWidth = 0
		}
		var remainingWidth int
		remainingWidth = barWidth - completedWidth
		if remainingWidth < 0 {
			remainingWidth = 0
		}
		progressBar = fmt.Sprintf("[%s%s] %.0f%%",
			strings.Repeat("█", completedWidth),
			strings.Repeat("░", remainingWidth),
			progressPercent*100)
	}

	// Print progress information at the bottom
	repeatCount = min(termWidth-5, 60)
	if repeatCount < 1 {
		repeatCount = 1
	}
	output.WriteString(strings.Repeat("─", repeatCount) + "\n")
	elapsed := time.Since(ui.startTime).Round(time.Second)

	// Format the ETA string
	etaStr := "calculating..."
	if progressPercent >= 1.0 {
		etaStr = "done"
	} else if progressPercent > 0 && !ui.cancelled && totalCompleted > 0 {
		if ui.duration > 0 {
			// For duration mode, format with only one fractional digit
			remainingTime := ui.duration - time.Since(ui.startTime)
			if remainingTime < 0 {
				remainingTime = 0
			}

			// Format with appropriate units and single fractional digit
			if remainingTime >= time.Hour {
				etaStr = fmt.Sprintf("%.1fh", remainingTime.Hours())
			} else if remainingTime >= time.Minute {
				etaStr = fmt.Sprintf("%.1fm", remainingTime.Minutes())
			} else {
				etaStr = fmt.Sprintf("%.1fs", remainingTime.Seconds())
			}
		} else {
			// For iteration mode, use the standard duration format
			etaStr = eta.String()
		}
	}

	output.WriteString(fmt.Sprintf("%s %s | %s %s | %s %s\n",
		labelColor("Elapsed:"),
		valueColor(elapsed.String()),
		labelColor("Progress:"),
		progressColor(progressBar),
		labelColor("ETA:"),
		valueColor(etaStr)))

	// Print footer
	if ui.finished {
		if ui.cancelled {
			// Only show cancelled message if explicitly cancelled
			output.WriteString("\n" + cancelledColor("⚠️  Benchmark cancelled!") + "\n")
		} else if ui.duration > 0 {
			// For duration-based benchmarks, always show completed
			output.WriteString("\n" + completedColor("✅ Benchmark completed!") + "\n")
		} else if totalCompleted < totalExpected {
			// For iteration-based benchmarks, check if all iterations completed
			output.WriteString("\n" + cancelledColor("⚠️  Benchmark cancelled!") + "\n")
		} else {
			output.WriteString("\n" + completedColor("✅ Benchmark completed!") + "\n")
		}
	} else {
		output.WriteString("\n" + valueColor("Press Ctrl+C to interrupt") + "\n")
	}

	// For the final render, don't clear previous output
	// This ensures we don't get duplicate headers when cancelling
	if ui.finished {
		// For the final output, just print directly without clearing
		// But first clear the previous output to avoid duplication
		if ui.lastLines > 0 {
			// Move cursor up by the number of lines we printed last time
			fmt.Printf("\033[%dA", ui.lastLines)
			// Clear from cursor to end of screen
			fmt.Print("\033[J")
		}

		// Print the final output
		fmt.Print(output.String())

		// Reset lastLines to avoid issues with future renders
		ui.lastLines = 0
	} else {
		// Normal update during benchmark - clear previous output
		if ui.lastLines > 0 {
			// Move cursor up by the number of lines we printed last time
			fmt.Printf("\033[%dA", ui.lastLines)
			// Clear from cursor to end of screen
			fmt.Print("\033[J")
		}

		// Print the new output
		fmt.Print(output.String())

		// Count the number of lines we just printed
		ui.lastLines = strings.Count(output.String(), "\n")
	}
}

func StartInlineUI(runs int, duration time.Duration, colorSchemeName string) error {
	// Create an inline UI instance
	ui := NewInlineUI(runs)
	ui.duration = duration

	// Set the color scheme
	if colorSchemeName != "" {
		scheme, err := colorscheme.GetScheme(colorSchemeName)
		if err != nil {
			return fmt.Errorf("invalid color scheme: %v (use --help to see available schemes)", err)
		}
		ui.colorScheme = scheme
	} else {
		// Use default scheme
		ui.colorScheme = colorscheme.Default()
	}

	// Store the UI instance globally for the event handler
	globalInlineUI = ui

	return nil
}

// Global inline UI instance
var globalInlineUI *InlineUI

func GetGlobalInlineUI() (*InlineUI, bool) {
	if globalInlineUI == nil {
		return nil, false
	}
	return globalInlineUI, true
}
