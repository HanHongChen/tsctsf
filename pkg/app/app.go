package app

import (
	tsctsf_context "github.com/HanHongChen/tsctsf/internal/context"
	"github.com/HanHongChen/tsctsf/pkg/factory"
)

type App interface {
	SetLogEnable(enable bool)
	SetLogLevel(level string)
	SetReportCaller(reportCaller bool)

	Start()
	Terminate()

	Context() *tsctsf_context.TSCTSFContext
	Config() *factory.Config
}
