package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/ohhfishal/fishy/flashcard"
	"github.com/ohhfishal/fishy/notify"
	"github.com/ohhfishal/fishy/serve"
	konghelp "github.com/ohhfishal/kong-help"
)

var ErrDone = errors.New("program ready to exit")

type Cmd struct {
	LogConfig LogConfig             `embed:"" group:"Logging Flags:"`
	Generate  flashcard.GenerateCMD `cmd:"" default:"withargs" help:"Generate all flashcards."`
	Notify    notify.NotifyCMD      `cmd:"" help:"Use generated flashcards to notify."`
	Serve     serve.CMD             `cmd:"" help:"Run as a server to periodically send notifications."`
}

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()
	if err := Run(ctx, os.Stdin, os.Stdout, os.Stderr, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func Run(ctx context.Context, stdin io.Reader, stdout io.Writer, stderr io.Writer, args []string) error {
	var exit bool
	var cmd Cmd
	parser, err := kong.New(
		&cmd,
		kong.Exit(func(_ int) { exit = true }),
		konghelp.Help(),
		kong.BindTo(ctx, new(context.Context)),
		kong.BindTo(stdout, new(io.Writer)),
		kong.BindTo(stdin, new(io.Reader)),
	)
	if err != nil {
		return err
	}

	parser.Stdout = stdout
	parser.Stderr = stdout

	context, err := parser.Parse(
		os.Args[1:],
	)
	if errors.Is(err, ErrDone) {
		return nil
	} else if err != nil || exit {
		return err
	}

	logger := cmd.LogConfig.NewLogger(stderr)
	if err := context.Run(logger); err != nil {
		// TODO: Handle some of the run options
		logger.Error("failed to run", "error", err)
		return nil
	}
	return nil
}

type LogConfig struct {
	Disable     bool       `help:"Disable logging. Shorthand for handler=discard."`
	HandlerType string     `name:"handler" enum:"json,discard,text" env:"HANDLER" default:"json" help:"Handler to use (${enum}) (env=$$${env})"`
	Level       slog.Level `default:"info"`
	AddSource   bool       `default:"false"`
	SetDefault  bool       `default:"true" help:"Set the global slog logger to usse this config."`
}

func (config *LogConfig) AfterApply() error {
	if config.Disable {
		config.HandlerType = "discard"
	}
	return nil
}

func (config LogConfig) NewLogger(stdout io.Writer) *slog.Logger {
	logger := slog.New(config.Handler(stdout))
	if config.SetDefault {
		slog.SetDefault(logger)
	}
	return logger
}

func (config LogConfig) Handler(stdout io.Writer) slog.Handler {
	switch config.HandlerType {
	case "discard":
		return slog.DiscardHandler
	case "json":
		return slog.NewJSONHandler(stdout, &slog.HandlerOptions{
			AddSource: config.AddSource,
			Level:     config.Level,
		})
	case "text":
		fallthrough
	default:
		return slog.NewTextHandler(stdout, &slog.HandlerOptions{
			AddSource: config.AddSource,
			Level:     config.Level,
		})
	}
}
