package sbi

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/HanHongChen/tsctsf/internal/logger"
	"github.com/HanHongChen/tsctsf/internal/sbi/consumer"
	"github.com/HanHongChen/tsctsf/internal/sbi/processor"
	"github.com/HanHongChen/tsctsf/pkg/app"
	"github.com/HanHongChen/tsctsf/pkg/factory"
	"github.com/free5gc/util/httpwrapper"
	logger_util "github.com/free5gc/util/logger"
)

type Route struct {
	Name    string
	Method  string
	Pattern string
	APIFunc gin.HandlerFunc
}

func applyRoutes(group *gin.RouterGroup, routes []Route) {
	for _, route := range routes {
		switch route.Method {
		case "GET":
			group.GET(route.Pattern, route.APIFunc)
		case "POST":
			group.POST(route.Pattern, route.APIFunc)
		case "PUT":
			group.PUT(route.Pattern, route.APIFunc)
		case "PATCH":
			group.PATCH(route.Pattern, route.APIFunc)
		case "DELETE":
			group.DELETE(route.Pattern, route.APIFunc)
		}
	}
}

type tsctsf interface {
	app.App
	Processor() *processor.Processor
	Consumer() *consumer.Consumer
	CancelContext() context.Context
}

type Server struct {
	tsctsf

	httpServer *http.Server
	router     *gin.Engine
}

func NewServer(tsctsf tsctsf, tlsKeyLogPath string) (*Server, error) {
	s := &Server{
		tsctsf: tsctsf,
		router: logger_util.NewGinWithLogrus(logger.GinLog),
	}

	qosAndTscRoutes := s.getQoSAnsTscRoutes()
	qosAndTscGroup := s.router.Group(factory.TsctsfQoSAndTscResUriPrefix)
	applyRoutes(qosAndTscGroup, qosAndTscRoutes)

	cfg := s.Config()
	bindAddr := cfg.GetSbiBindingAddr()
	logger.SBILog.Infof("Binding addr: [%s]", bindAddr)
	var err error
	if s.httpServer, err = httpwrapper.NewHttp2Server(bindAddr, tlsKeyLogPath, s.router); err != nil {
		logger.InitLog.Errorf("Initialize HTTP server failed: %v", err)
		return nil, err
	}
	s.httpServer.ErrorLog = log.New(logger.SBILog.WriterLevel(logrus.ErrorLevel), "HTTP2: ", 0)

	return s, nil
}

func (s *Server) Run(traceCtx context.Context, wg *sync.WaitGroup) error {
	var err error
	_, s.Context().NfId, err = s.Consumer().SendRegisterNFInstance(s.CancelContext())
	if err != nil {
		logger.InitLog.Errorf("tsctsf register to NRF Error[%s]", err.Error())
	}

	wg.Add(1)
	go s.startServer(wg)

	return nil
}

func (s *Server) Shutdown(traceCtx context.Context) {
	const defaultShutdownTimeout time.Duration = 2 * time.Second

	if s.httpServer != nil {
		logger.SBILog.Infof("Stop SBI server (listen on %s)", s.httpServer.Addr)
		toCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()
		if err := s.httpServer.Shutdown(toCtx); err != nil {
			logger.SBILog.Errorf("Could not close SBI server: %#v", err)
		}
	}
}

func (s *Server) startServer(wg *sync.WaitGroup) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.SBILog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
			s.Terminate()
		}

		wg.Done()
	}()

	logger.SBILog.Infof("Start SBI server (listen on %s)", s.httpServer.Addr)

	var err error
	// cfg := s.Config()
	// scheme := cfg.GetSbiScheme()
	// if scheme == "http" {
	// 	err = s.httpServer.ListenAndServe()
	// } else if scheme == "https" {
	// 	err = s.httpServer.ListenAndServeTLS(
	// 		cfg.GetCertPemPath(),
	// 		cfg.GetCertKeyPath())
	// } else {
	// 	err = fmt.Errorf("No support this scheme[%s]", scheme)
	// }
	err = s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.SBILog.Errorf("SBI server error: %v", err)
	}
	logger.SBILog.Infof("SBI server (listen on %s) stopped", s.httpServer.Addr)
}
