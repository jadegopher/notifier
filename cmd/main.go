package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notifier"
	"notifier/log"
	"notifier/log/tag"
)

type Config struct {
	URL      string
	Interval time.Duration
}

func main() {
	cfg := Config{}

	err := cfg.ParseFlags()
	if err != nil {
		os.Exit(1)
	}

	slog.SetLogLoggerLevel(slog.LevelDebug)

	// Initialize the notifier
	// The library handles buffering internally via FlushInterval
	n := notifier.Default(
		cfg.URL,
		notifier.Options{
			FlushInterval: cfg.Interval, // Use the parsed interval
		},
	)

	if err = run(n); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run orchestrates the lifecycle: Start -> Wait for Input/Signal -> Stop
func run(n *notifier.Notifier) error {
	n.Start()
	defer n.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)

	done := make(chan struct{})

	go func() {
		readStdin(n)
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-sigChan:
		log.Info("Gracefully shutting down...")
		return nil
	}
}

// readStdin scans standard input line by line
func readStdin(n *notifier.Notifier) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		n.Notify(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Error("scanner returned an error", tag.Err, err.Error())
	}
}

func (config *Config) ParseFlags() error {
	url := flag.String("url", "http://localhost:8080/notify", "Target URL for notifications")
	interval := flag.Duration("i", 5*time.Second, "Notification interval")

	flag.Usage = func() {
		flag.PrintDefaults()
	}

	flag.Parse()

	config.URL = *url
	config.Interval = *interval

	return nil
}
