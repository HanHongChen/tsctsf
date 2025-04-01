package logger

import (
	"github.com/sirupsen/logrus"

	logger_util "github.com/free5gc/util/logger"
)

var (
	Log            *logrus.Logger
	NfLog          *logrus.Entry
	MainLog        *logrus.Entry
	InitLog        *logrus.Entry
	CfgLog         *logrus.Entry
	CtxLog         *logrus.Entry
	GinLog         *logrus.Entry
	ConsumerLog    *logrus.Entry
	PolicyAuthLog  *logrus.Entry
	UtilLog        *logrus.Entry
	TimeSyncSubLog *logrus.Entry
	TimeSyncCfgLog *logrus.Entry
	TSCAppSessLog  *logrus.Entry
)

func init() {
	fieldsOrder := []string{
		logger_util.FieldNF,
		logger_util.FieldCategory,
	}

	Log = logger_util.New(fieldsOrder)
	NfLog = Log.WithField(logger_util.FieldNF, "TSCTSF")
	MainLog = NfLog.WithField(logger_util.FieldCategory, "Main")
	InitLog = NfLog.WithField(logger_util.FieldCategory, "Init")
	CfgLog = NfLog.WithField(logger_util.FieldCategory, "CFG")
	CtxLog = NfLog.WithField(logger_util.FieldCategory, "CTX")
	GinLog = NfLog.WithField(logger_util.FieldCategory, "GIN")
	ConsumerLog = NfLog.WithField(logger_util.FieldCategory, "Consumer")
	PolicyAuthLog = NfLog.WithField(logger_util.FieldCategory, "PolicyAuth")
	UtilLog = NfLog.WithField(logger_util.FieldCategory, "Util")
	TimeSyncSubLog = NfLog.WithField(logger_util.FieldCategory, "TimeSyncSub")
	TimeSyncCfgLog = NfLog.WithField(logger_util.FieldCategory, "TimeSyncCfg")
	TSCAppSessLog = NfLog.WithField(logger_util.FieldCategory, "TSCAppSessLog")

}
