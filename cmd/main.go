package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"time"

	logger_util "github.com/free5gc/util/logger"
	"github.com/free5gc/util/version"

	"github.com/HanHongChen/tsctsf/internal/logger"
	"github.com/HanHongChen/tsctsf/pkg/factory"
	"github.com/HanHongChen/tsctsf/pkg/service"
	"github.com/urfave/cli"
)

var TSCTSF *service.TsctsfApp

func main() {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.MainLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	app := cli.NewApp()
	app.Name = "tsctsf"
	app.Usage = "5G Time Sensitive Communication and Time Synchronization Function (TSCTSF)"
	app.Action = action
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Load configuration from `FILE`",
		},
		cli.StringSliceFlag{
			Name:  "log, l",
			Usage: "Output NF log to `FILE`",
		},
	}
	rand.Seed(time.Now().UnixNano())

	if err := app.Run(os.Args); err != nil {
		logger.MainLog.Errorf("TSCTSF Run Error: %v\n", err)
	}
}

func action(cliCtx *cli.Context) error {
	// Initializes log files based on command-line arguments and gets the TLS key log path
	tlsKeyLogPath, err := initLogFile(cliCtx.StringSlice("log"))
	if err != nil {
		return err
	}

	logger.MainLog.Infoln(cliCtx.App.Name)
	logger.MainLog.Infoln("TSCTSF version: ", version.GetVersion())

	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh  // Wait for interrupt signal to gracefully shutdown UPF
		cancel() // Notify each goroutine and wait them stopped
	}()

	// Reads the configuration file specified by the user
	cfg, err := factory.ReadConfig(cliCtx.String("config"))
	if err != nil {
		return err
	}
	factory.TsctsfConfig = cfg

	// Creates a new TSCTSF application instance with the configuration
	tsctsf, err := service.NewApp(ctx, cfg, tlsKeyLogPath)
	if err != nil {
		sigCh <- nil
		return err
	}

	// Stores the app instance in the global TSCTSF variable
	TSCTSF = tsctsf
	if tsctsf == nil {
		logger.MainLog.Infoln("tsctsf is nil")
	}

	// Starts the TSCTSF application with the TLS key log path
	tsctsf.Start()

	return nil
}

func initLogFile(logNfPath []string) (string, error) {
	logTlsKeyPath := ""

	for _, path := range logNfPath {
		if err := logger_util.LogFileHook(logger.Log, path); err != nil {
			return "", err
		}

		if logTlsKeyPath != "" {
			continue
		}

		nfDir, _ := filepath.Split(path)
		tmpDir := filepath.Join(nfDir, "key")
		if err := os.MkdirAll(tmpDir, 0o775); err != nil {
			logger.InitLog.Errorf("Make directory %s failed: %+v", tmpDir, err)
			return "", err
		}
		_, name := filepath.Split(factory.TsctsfDefaultTLSKeyLogPath)
		logTlsKeyPath = filepath.Join(tmpDir, name)
	}

	return logTlsKeyPath, nil
}
