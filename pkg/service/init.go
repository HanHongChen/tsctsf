package service

import (
	"context"
	"io/ioutil"
	"os"
	"runtime/debug"
	"sync"

	tsctsf_context "github.com/HanHongChen/tsctsf/internal/context"
	"github.com/HanHongChen/tsctsf/internal/logger"
	"github.com/HanHongChen/tsctsf/internal/sbi"
	"github.com/HanHongChen/tsctsf/internal/sbi/consumer"
	"github.com/HanHongChen/tsctsf/internal/sbi/processor"

	// timesynchronization "github.com/HanHongChen/tsctsf/internal/sbi/timesynchronization"

	"github.com/HanHongChen/tsctsf/pkg/app"
	"github.com/HanHongChen/tsctsf/pkg/factory"
	"github.com/sirupsen/logrus"
)

var TSCTSF *TsctsfApp
var _ app.App = &TsctsfApp{}

type TsctsfApp struct {
	app.App
	cfg       *factory.Config
	tsctsfCtx *tsctsf_context.TSCTSFContext
	ctx       context.Context
	cancel    context.CancelFunc

	consumer  *consumer.Consumer
	processor *processor.Processor
	sbiServer *sbi.Server
	wg        sync.WaitGroup
}

func NewApp(
	ctx context.Context,
	cfg *factory.Config,
	tlsKeyLogPath string,
) (*TsctsfApp, error) {
	tsctsf := &TsctsfApp{
		cfg: cfg,
		wg:  sync.WaitGroup{},
	}
	tsctsf.SetLogEnable(cfg.GetLogEnable())
	tsctsf.SetLogLevel(cfg.GetLogLevel())
	tsctsf.SetReportCaller(cfg.GetLogReportCaller())

	tsctsf.ctx, tsctsf.cancel = context.WithCancel(ctx)
	tsctsf_context.Init()
	tsctsf.tsctsfCtx = tsctsf_context.GetSelf()

	//consumer
	consumer, err := consumer.NewConsumer(tsctsf)
	if err != nil {
		return tsctsf, err
	}
	tsctsf.consumer = consumer

	//processor
	p, err := processor.NewProcessor(tsctsf)
	if err != nil {
		return tsctsf, err
	}
	tsctsf.processor = p

	if tsctsf.sbiServer, err = sbi.NewServer(tsctsf, tlsKeyLogPath); err != nil {
		return nil, err
	}
	TSCTSF = tsctsf

	return tsctsf, nil
}

func (a *TsctsfApp) Start() {
	logger.InitLog.Infoln("Server started")
	a.wg.Add(1)
	go a.listenShutdownEvent()
	if err := a.sbiServer.Run(context.Background(), &a.wg); err != nil {
		logger.InitLog.Fatalf("Run SBI server failed: %+v", err)
	}
	a.WaitRoutineStopped()
}

func (a *TsctsfApp) WaitRoutineStopped() {
	a.wg.Wait()
	logger.MainLog.Infof("TSCTSF App is terminated")
}

func (a *TsctsfApp) listenShutdownEvent() {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.InitLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
		a.wg.Done()
	}()

	<-a.ctx.Done()
	a.terminateProcedure()
}

func (a *TsctsfApp) terminateProcedure() {
	logger.MainLog.Infof("Terminating TSCTSF...")
	a.CallServerStop()
	// deregister with NRF
	problemDetails, err := a.Consumer().SendDeregisterNFInstance()
	if problemDetails != nil {
		logger.InitLog.Errorf("Deregister NF instance Failed Problem[%+v]", problemDetails)
	} else if err != nil {
		logger.InitLog.Errorf("Deregister NF instance Error[%+v]", err)
	} else {
		logger.InitLog.Infof("Deregister from NRF successfully")
	}
	logger.InitLog.Infof("TSCTSF terminated")
}

func (a *TsctsfApp) CallServerStop() {
	if a.sbiServer != nil {
		a.sbiServer.Shutdown(context.Background())
	}
}

func (a *TsctsfApp) Config() *factory.Config {
	return a.cfg
}

func (a *TsctsfApp) Context() *tsctsf_context.TSCTSFContext {
	return a.tsctsfCtx
}

func (a *TsctsfApp) CancelContext() context.Context {
	return a.ctx
}

func (a *TsctsfApp) Consumer() *consumer.Consumer {
	return a.consumer
}

func (a *TsctsfApp) Processor() *processor.Processor {
	return a.processor
}

func (a *TsctsfApp) SetLogEnable(enable bool) {
	logger.MainLog.Infof("Log enable is set to [%v]", enable)
	if enable && logger.Log.Out == os.Stderr {
		return
	} else if !enable && logger.Log.Out == ioutil.Discard {
		return
	}

	a.cfg.SetLogEnable(enable)
	if enable {
		logger.Log.SetOutput(os.Stderr)
	} else {
		logger.Log.SetOutput(ioutil.Discard)
	}
}

func (a *TsctsfApp) SetLogLevel(level string) {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		logger.MainLog.Warnf("Log level [%s] is invalid", level)
		return
	}

	logger.MainLog.Infof("Log level is set to [%s]", level)
	if lvl == logger.Log.GetLevel() {
		return
	}

	a.cfg.SetLogLevel(level)
	logger.Log.SetLevel(lvl)
}

func (a *TsctsfApp) SetReportCaller(reportCaller bool) {
	logger.MainLog.Infof("Report Caller is set to [%v]", reportCaller)
	if reportCaller == logger.Log.ReportCaller {
		return
	}

	a.cfg.SetLogReportCaller(reportCaller)
	logger.Log.SetReportCaller(reportCaller)
}

func (a *TsctsfApp) Terminate() {
	logger.InitLog.Infof("Terminating TSCTSF...")
	// deregister with NRF
	problemDetails, err := a.Consumer().SendDeregisterNFInstance()
	if problemDetails != nil {
		logger.InitLog.Errorf("Deregister NF instance Failed Problem[%+v]", problemDetails)
	} else if err != nil {
		logger.InitLog.Errorf("Deregister NF instance Error[%+v]", err)
	} else {
		logger.InitLog.Infof("Deregister from NRF successfully")
	}

	logger.InitLog.Infof("TSCTSF terminated")
}

// func (a *TsctsfApp) Start(tlsKeyLogPath string) {
// 	logger.InitLog.Infoln("Server started")
// 	// pemPath := factory.TsctsfDefaultCertPemPath
// 	// keyPath := factory.TsctsfDefaultPrivateKeyPath
// 	sbi := factory.TsctsfConfig.Configuration.Sbi
// 	if sbi.Tls != nil {
// 		// pemPath = sbi.Tls.Pem
// 		// keyPath = sbi.Tls.Key
// 	}
// 	router := logger_util.NewGinWithLogrus(logger.GinLog)

// 	// policyauthorization.AddService(router)
// 	// timesynchronization.AddService(router)
// 	// bridgeinfomangement.AddService(router)
// 	qosandtscassistance.AddService(router)

// 	router.Use(cors.New(cors.Config{
// 		AllowMethods: []string{"GET", "POST", "OPTIONS", "PUT", "PATCH", "DELETE"},
// 		AllowHeaders: []string{
// 			"Origin", "Content-Length", "Content-Type", "User-Agent",
// 			"Referrer", "Host", "Token", "X-Requested-With",
// 		},
// 		ExposeHeaders:    []string{"Content-Length"},
// 		AllowCredentials: true,
// 		AllowAllOrigins:  true,
// 		MaxAge:           86400,
// 	}))

// 	self := a.tsctsfCtx
// 	// Register to NRF
// 	profile, err := consumer.BuildNFInstance(self)
// 	if err != nil {
// 		logger.InitLog.Error("Build TSCTSF Profile Error")
// 	}
// 	_, self.NfId, err = consumer.SendRegisterNFInstance(self.NrfUri, self.NfId, profile)
// 	if err != nil {
// 		logger.InitLog.Errorf("TSCTSF register to NRF Error[%s]", err.Error())
// 	}

// 	// Handle terminal process
// 	signalChannel := make(chan os.Signal, 1)
// 	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
// 	go func() {
// 		defer func() {
// 			if p := recover(); p != nil {
// 				// Print stack for panic to log. Fatalf() will let program exit.
// 				logger.InitLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
// 			}
// 		}()

// 		<-signalChannel
// 		a.Terminate()
// 		os.Exit(0)
// 	}()

// 	HTTPAddr := fmt.Sprintf("%s:%d", factory.TsctsfConfig.Configuration.Sbi.BindingIPv4, factory.TsctsfConfig.Configuration.Sbi.Port)
// 	server, err := httpwrapper.NewHttp2Server(HTTPAddr, tlsKeyLogPath, router)
// 	if server == nil {
// 		logger.InitLog.Error("Initialize HTTP server failed:", err)
// 		return
// 	}
// 	if err != nil {
// 		logger.InitLog.Warnln("Initialize HTTP server:", err)
// 	}

// 	serverScheme := factory.TsctsfConfig.Configuration.Sbi.Scheme
// 	if serverScheme == "http" {
// 		err = server.ListenAndServe()
// 		// } else if serverScheme == "https" {
// 		// err = server.ListenAndServeTLS(pemPath, keyPath)
// 	}

// 	if err != nil {
// 		logger.InitLog.Fatalf("HTTP server setup failed: %+v", err)
// 	}
// }
