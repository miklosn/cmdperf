package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/miklosn/cmdperf/internal/benchmark"
	"github.com/miklosn/cmdperf/internal/command"
	"github.com/miklosn/cmdperf/internal/output"
	"github.com/miklosn/cmdperf/internal/ui"
	"github.com/miklosn/cmdperf/internal/ui/colorscheme"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

var cli struct {
	Commands         []string      `arg:"" name:"command" help:"Command(s) to benchmark" optional:""`
	Runs             int           `short:"n" name:"runs" help:"Number of runs to perform" default:"10"`
	Concurrency      int           `short:"c" name:"concurrency" help:"Number of concurrent executions" default:"1"`
	ColorScheme      string        `name:"color-scheme" help:"${color_scheme_help}" default:"auto"`
	ListColorSchemes bool          `name:"list-color-schemes" help:"List available color schemes"`
	Timeout          time.Duration `short:"t" name:"timeout" help:"Timeout for each command execution" default:"1m"`
	Duration         time.Duration `short:"d" name:"duration" help:"Total benchmark duration (overrides --runs)"`
	Shell            string        `short:"s" name:"shell" help:"Shell to use for command execution" default:"/bin/sh"`
	ShellOptions     []string      `name:"shell-opt" help:"Shell option (can be repeated)" default:"-c"`
	NoShell          bool          `short:"N" name:"no-shell" help:"Execute commands directly without a shell"`
	CSVOutput        string        `name:"csv" help:"Write results to CSV file"`
	MarkdownOutput   string        `name:"markdown" help:"Write results to Markdown file"`
	Version          bool          `name:"version" help:"Show version information"`
	FailOnError      bool          `name:"fail-on-error" help:"Exit with non-zero status if any command returns non-zero exit code"`
	CPUProfile       string        `name:"cpu-profile" help:"Write CPU profile to file"`
	MemProfile       string        `name:"mem-profile" help:"Write memory profile to file"`
	BlockProfile     string        `name:"block-profile" help:"Write goroutine blocking profile to file"`
	PprofServer      bool          `name:"pprof-server" help:"Start pprof HTTP server on :6060"`
	Rate             float64       `short:"r" name:"rate" help:"Maximum rate of requests per second per worker (0 = unlimited)"`
}

func splitCommandRespectingQuotes(cmd string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for _, r := range cmd {
		switch {
		case (r == '"' || r == '\'') && !inQuotes:
			inQuotes = true
			quoteChar = r
		case r == quoteChar && inQuotes:
			inQuotes = false
			quoteChar = rune(0)
		case r == ' ' && !inQuotes:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func main() {
	colorSchemeHelp := fmt.Sprintf("Color scheme to use (%s)", strings.Join(colorscheme.ListSchemes(), ", "))

	ctx := kong.Parse(&cli,
		kong.Name("cmdperf"),
		kong.Description("✨ Command Performance Benchmarking ✨"),
		kong.UsageOnError(),
		kong.Vars{
			"version":           version,
			"color_scheme_help": colorSchemeHelp,
		},
	)

	if cli.Version {
		fmt.Printf("cmdperf version %s (built %s)\n", version, buildTime)
		os.Exit(0)
	}

	if cli.ListColorSchemes {
		fmt.Print(colorscheme.FormatSchemeList())
		os.Exit(0)
	}

	if len(cli.Commands) == 0 {
		fmt.Println("Error: at least one command is required")
		ctx.PrintUsage(false)
		os.Exit(1)
	}

	if cli.PprofServer {
		go func() {
			log.Println("Starting pprof server on :6060")
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	if cli.CPUProfile != "" {
		f, err := os.Create(cli.CPUProfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating CPU profile: %v\n", err)
			os.Exit(1)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	commands := make([]*command.Command, len(cli.Commands))
	for i, cmdStr := range cli.Commands {
		if cli.NoShell {
			parts := splitCommandRespectingQuotes(cmdStr)
			if len(parts) == 0 {
				fmt.Fprintf(os.Stderr, "Error: empty command\n")
				os.Exit(1)
			}

			commands[i] = &command.Command{
				Raw:         cmdStr,
				DirectExec:  true,
				Command:     parts[0],
				Args:        parts[1:],
				Timeout:     cli.Timeout,
				Parallelism: cli.Concurrency,
			}

		} else {
			commands[i] = &command.Command{
				Raw:          cmdStr,
				Shell:        cli.Shell,
				ShellOptions: cli.ShellOptions,
				Timeout:      cli.Timeout,
				Parallelism:  cli.Concurrency,
			}
		}
	}

	options := benchmark.Options{
		Iterations:  cli.Runs,
		Parallelism: cli.Concurrency,
		Timeout:     cli.Timeout,
		Duration:    cli.Duration,
		Rate:        cli.Rate,
	}

	runner, err := benchmark.NewRunner(commands, options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating benchmark runner: %v\n", err)
		os.Exit(1)
	}

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nBenchmark interrupted, cleaning up...")

		if inlineUI, ok := ui.GetGlobalInlineUI(); ok {
			// Use the Cancel method to properly mark the UI as cancelled
			inlineUI.Cancel()
		}

		cancel()

		go func() {
			time.Sleep(2 * time.Second)
			fmt.Println("Forced exit due to slow shutdown")
			os.Exit(1)
		}()
	}()

	err = ui.StartInlineUI(cli.Runs, cli.Duration, cli.ColorScheme)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting UI: %v\n", err)
		os.Exit(1)
	}

	runner.SetProgressCallback(func(stats []*benchmark.CommandStats, complete bool) {
		if inlineUI, ok := ui.GetGlobalInlineUI(); ok {
			inlineUI.Update(stats, complete)
		}
	})

	runner.Run(runCtx)

	if cli.CSVOutput != "" {
		absPath, _ := filepath.Abs(cli.CSVOutput)

		file, err := os.Create(cli.CSVOutput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating CSV file at %s: %v\n", absPath, err)
		} else {
			defer file.Close()

			csvWriter, _ := output.GetWriter("csv")
			if err := csvWriter.Write(file, runner.Results); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing CSV results to %s: %v\n", absPath, err)
			} else {
				fmt.Printf("CSV results written to %s\n", absPath)
			}
		}
	}

	if cli.MarkdownOutput != "" {
		absPath, _ := filepath.Abs(cli.MarkdownOutput)

		file, err := os.Create(cli.MarkdownOutput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating Markdown file at %s: %v\n", absPath, err)
		} else {
			defer file.Close()

			mdWriter, _ := output.GetWriter("markdown")
			if err := mdWriter.Write(file, runner.Results); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing Markdown results to %s: %v\n", absPath, err)
			} else {
				fmt.Printf("Markdown results written to %s\n", absPath)
			}
		}
	}

	if cli.FailOnError {
		hasNonZeroExit := false
		for _, stat := range runner.Results {
			for exitCode, count := range stat.ExitCodes {
				if exitCode != 0 && count > 0 {
					hasNonZeroExit = true
					break
				}
			}
			if hasNonZeroExit {
				break
			}
		}

		if hasNonZeroExit {
			fmt.Fprintf(os.Stderr, "Error: Some commands returned non-zero exit codes\n")
			os.Exit(1)
		}
	}

	if cli.MemProfile != "" {
		f, err := os.Create(cli.MemProfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating memory profile: %v\n", err)
			return
		}
		defer f.Close()

		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing memory profile: %v\n", err)
		}
	}

	if cli.BlockProfile != "" {
		f, err := os.Create(cli.BlockProfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating block profile: %v\n", err)
			return
		}
		defer f.Close()

		if err := pprof.Lookup("block").WriteTo(f, 0); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing block profile: %v\n", err)
		}
	}
}
