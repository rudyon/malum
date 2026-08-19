package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rudyon/malum/internal/catalog"
	"github.com/rudyon/malum/internal/httpapi"
	"github.com/rudyon/malum/internal/ingest/webpage"
	"github.com/rudyon/malum/internal/library"
	"github.com/rudyon/malum/internal/safefetch"
	authorstore "github.com/rudyon/malum/internal/storage/author"
	documentstore "github.com/rudyon/malum/internal/storage/document"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := run(logger, os.Args[1:]); err != nil {
		logger.Error("Malum stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger, arguments []string) error {
	flags := flag.NewFlagSet("malum", flag.ContinueOnError)
	dataRoot := flags.String("data-dir", "", "directory containing Malum's database and documents")
	listenAddress := flags.String("listen", "127.0.0.1:8080", "HTTP listen address")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if strings.TrimSpace(*dataRoot) == "" {
		return errors.New("--data-dir is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	catalogue, err := catalog.Open(ctx, *dataRoot)
	if err != nil {
		return err
	}
	defer catalogue.Close()

	client := safefetch.NewClient()
	service := library.New(
		webpage.NewImporter(client),
		catalogue,
		documentstore.New(*dataRoot),
		authorstore.New(*dataRoot),
	)
	server := &http.Server{
		Addr:              *listenAddress,
		Handler:           httpapi.New(service, logger),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	serveResult := make(chan error, 1)
	go func() {
		logger.Info("Malum API listening", "address", *listenAddress, "dataDir", *dataRoot)
		serveResult <- server.ListenAndServe()
	}()

	select {
	case err := <-serveResult:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve Malum API: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shut down Malum API: %w", err)
		}
		return nil
	}
}
