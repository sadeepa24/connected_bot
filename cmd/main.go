package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	connected "github.com/sadeepa24/connected_bot"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Mainctx context.Context

func main() {

	Mainctx = context.Background()

	botoption, err := getBotOption()
	if err != nil {
		log.Fatal(err.Error())
	}
	if err = runConnected(botoption); err != nil {
		log.Fatal("bot exit with err ",  zap.Error(err))
	}
}

func runConnected(botoptions connected.Botoptions) error {
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(osSignals)

	for {
		ctx, cancel := context.WithCancel(Mainctx)
		botoptions.Ctx = ctx
		defer cancel()
		bot, err := connected.New(botoptions)
		if err != nil {
			return err
		}
		if err = bot.Start(); err != nil {
			return err
		}
		osSignal := <-osSignals

		if osSignal == syscall.SIGHUP {
			if closeErr := bot.Close(); closeErr != nil {
				botoptions.Logger.Error("Error closing bot:" + closeErr.Error())
			}
			cancel()
			runtime.GC() // to make fresh restart
			if botoptions, err = getBotOption(); err != nil {
				return errors.New("error while restarting, when option building " + err.Error())
			}
			continue
		} else {
			if closeErr := bot.Close(); closeErr != nil {
				botoptions.Logger.Error("Error closing bot:" + closeErr.Error())
				return closeErr
			}
			return nil
		}
	}
}


func getBotOption() (connected.Botoptions, error) {
	botoption, err := readBotConfig()
	if err != nil {
		return botoption, err
	}
	botoption.Ctx = Mainctx
	opt, err := readsboxconfigAT(botoption.SboxConfPath)
	if err != nil {
		return botoption, err
	}
	botoption.Sboxoption = opt.options

	if botoption.LoggerOption.Encoding == "" {
		botoption.LoggerOption.Encoding = "console"
	}
	if len(botoption.LoggerOption.Paths) == 0 {
		botoption.LoggerOption.Paths = append(botoption.LoggerOption.Paths, "stdout")
	}
	logconfig := zap.Config{
		Level: zap.NewAtomicLevelAt(zap.DebugLevel),
		Development: false,
		DisableCaller: true,
		DisableStacktrace: true,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: botoption.LoggerOption.Encoding,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       botoption.LoggerOption.Level,
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			// StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
			
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths: botoption.LoggerOption.Paths,
		// ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := logconfig.Build()
	if err != nil {
		log.Fatal("logger Building err - "+ err.Error() )
	}
	botoption.Logger = logger
	
	// if botoption.Templates, err = readTmpl(botoption.TemplatesPath); err != nil {
	// 	return botoption, err
	// }

	return botoption, nil
}

func readBotConfig() (connected.Botoptions, error) {
	file, err := os.ReadFile("./config.json")
	var opts connected.Botoptions
	if err != nil {
		return opts, err
	}
	return opts, json.Unmarshal(file, &opts)
}
