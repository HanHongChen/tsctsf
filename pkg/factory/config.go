package factory

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/HanHongChen/openapi-tsctsf/models"
	"github.com/HanHongChen/tsctsf/internal/logger"
	"github.com/asaskevich/govalidator"
)

const (
	TsctsfDefaultTLSKeyLogPath = "./log/tsctsfsslkey.log"
	// TsctsfDefaultCertPemPath        = "./cert/tsctsf.pem"
	// TsctsfDefaultPrivateKeyPath     = "./cert/tsctsf.key"
	TsctsfDefaultConfigPath = "./config/tsctsfcfg.yaml"
	TsctsfSbiDefaultIPv4    = "127.0.0.50"
	TsctsfSbiDefaultPort    = 8000
	TsctsfSbiDefaultScheme  = "https"
	TsctsfDefaultNrfUri     = "https://127.0.0.10:8000"
)

type Config struct {
	Info             *Info            `yaml:"info" valid:"required"`
	Configuration    *Configuration   `yaml:"configuration" valid:"required"`
	Subscriptions    []Subscription   `yaml:"subscriptions,omitempty"`
	Logger           *Logger          `yaml:"logger" valid:"required"`
	RouterInfo       []RouterInfo     `yaml:"routerInfo" valid:"required"`
	DetNetController DetNetController `yaml:"detnetController" valid:"required"`
	sync.RWMutex
}

type DetNetController struct {
	Scheme   string `yaml:"scheme" valid:"required,scheme"`
	IPv4     string `yaml:"ipv4,omitempty" valid:"required,host"`
	Port     int    `yaml:"port,omitempty" valid:"required,port"`
	User     string `yaml:"user,omitempty" valid:"required"`
	Password string `yaml:"password,omitempty" valid:"required"`
	Uri      string `yaml:"uri,omitempty" valid:"required"`
}

type RouterInfo struct {
	UpNodeId string `yaml:"upNodeId" valid:"required"`
	MacAddr  string `yaml:"ifMac"`
	UeIp     string `yaml:"ueIp" `
	Port     int    `yaml:"port"`
	Mtu      int    `yaml:"mtu"`
}

func (c *Config) Validate() (bool, error) {
	if configuration := c.Configuration; configuration != nil {
		if result, err := configuration.validate(); err != nil {
			return result, err
		}
	}

	result, err := govalidator.ValidateStruct(c)
	return result, appendInvalid(err)
}

type Info struct {
	Version     string `yaml:"version,omitempty" valid:"required,in(1.0.1)"`
	Description string `yaml:"description,omitempty" valid:"type(string)"`
}

type Configuration struct {
	TsctsfName      string   `yaml:"tsctsfName,omitempty" valid:"required, type(string)"`
	Sbi             *Sbi     `yaml:"sbi,omitempty" valid:"required"`
	NrfUri          string   `yaml:"nrfUri,omitempty" valid:"required, url"`
	PcfUri          string   `yaml:"pcfUri,omitempty" valid:"required, url"`
	ServiceNameList []string `yaml:"serviceNameList,omitempty" valid:"required"`
}

type Subscription struct {
	SubscriptionId   string                         `yaml:"subscriptionId"`
	SubscriptionData *models.TimeSyncExposureSubsc  `yaml:"subscriptionData"`
	ConfigurationId  string                         `yaml:"configurationId"`
	SubscriptionCfg  *models.TimeSyncExposureConfig `yaml:"subscriptionCfg"`
}

type Logger struct {
	Enable       bool   `yaml:"enable" valid:"type(bool)"`
	Level        string `yaml:"level" valid:"required,in(trace|debug|info|warn|error|fatal|panic)"`
	ReportCaller bool   `yaml:"reportCaller" valid:"type(bool)"`
}

func (c *Configuration) validate() (bool, error) {
	if c.Sbi != nil {
		if _, err := c.Sbi.validate(); err != nil {
			return false, err
		}
	}

	for index, serviceName := range c.ServiceNameList {
		switch {
		case serviceName == "ntsctsf_timesynchronization":
		case serviceName == "ntsctsf_qosandtscassistance":
		default:
			err := errors.New("Invalid serviceNameList[" + strconv.Itoa(index) + "]: " +
				serviceName + ", should be ntsctsf_timesynchronization.")
			return false, err
		}
	}

	// if result := govalidator.IsTime(c.TimeFormat, PcfTimeFormatLayout); !result {
	// 	err := fmt.Errorf("Invalid TimeFormat: %s, should be in 2019-01-02 15:04:05 format.", c.TimeFormat)
	// 	return false, err
	// }

	// if c.Mongodb != nil {
	// 	if _, err := c.Mongodb.validate(); err != nil {
	// 		return false, err
	// 	}
	// }

	if _, err := govalidator.ValidateStruct(c); err != nil {
		return false, appendInvalid(err)
	}

	return true, nil
}

type Service struct {
	ServiceName string `yaml:"serviceName" valid:"required, service"`
	SuppFeat    string `yaml:"suppFeat,omitempty" valid:"-"`
}
type Tls struct {
	Pem string `yaml:"pem,omitempty" valid:"type(string),minstringlength(1),required"`
	Key string `yaml:"key,omitempty" valid:"type(string),minstringlength(1),required"`
}

func (t *Tls) validate() (bool, error) {
	result, err := govalidator.ValidateStruct(t)
	return result, err
}

type Sbi struct {
	Scheme       string `yaml:"scheme" valid:"required,scheme"`
	RegisterIPv4 string `yaml:"registerIPv4,omitempty" valid:"required,host"` // IP that is registered at NRF.
	// IPv6Addr  string `yaml:"ipv6Addr,omitempty"`
	BindingIPv4 string `yaml:"bindingIPv4,omitempty" valid:"required,host"` // IP used to run the server in the node.
	Port        int    `yaml:"port,omitempty" valid:"required,port"`
	Tls         *Tls   `yaml:"tls,omitempty" valid:"optional"`
}

func (s *Sbi) validate() (bool, error) {
	govalidator.TagMap["scheme"] = govalidator.Validator(func(str string) bool {
		return str == "https" || str == "http"
	})

	if tls := s.Tls; tls != nil {
		if result, err := tls.validate(); err != nil {
			return result, err
		}
	}

	if _, err := govalidator.ValidateStruct(s); err != nil {
		return false, appendInvalid(err)
	}

	return true, nil
}

func appendInvalid(err error) error {
	var errs govalidator.Errors

	if err == nil {
		return nil
	}

	es := err.(govalidator.Errors).Errors()
	for _, e := range es {
		errs = append(errs, fmt.Errorf("Invalid %w", e))
	}

	return error(errs)
}

func (c *Config) GetVersion() string {
	c.RLock()
	defer c.RUnlock()

	if c.Info.Version != "" {
		return c.Info.Version
	}
	return ""
}

func (c *Config) SetLogEnable(enable bool) {
	c.Lock()
	defer c.Unlock()

	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		c.Logger = &Logger{
			Enable: enable,
			Level:  "info",
		}
	} else {
		c.Logger.Enable = enable
	}
}

func (c *Config) SetLogLevel(level string) {
	c.Lock()
	defer c.Unlock()

	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		c.Logger = &Logger{
			Level: level,
		}
	} else {
		c.Logger.Level = level
	}
}

func (c *Config) SetLogReportCaller(reportCaller bool) {
	c.Lock()
	defer c.Unlock()

	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		c.Logger = &Logger{
			Level:        "info",
			ReportCaller: reportCaller,
		}
	} else {
		c.Logger.ReportCaller = reportCaller
	}
}

func (c *Config) GetLogEnable() bool {
	c.RLock()
	defer c.RUnlock()
	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		return false
	}
	return c.Logger.Enable
}

func (c *Config) GetLogLevel() string {
	c.RLock()
	defer c.RUnlock()
	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		return "info"
	}
	return c.Logger.Level
}

func (c *Config) GetLogReportCaller() bool {
	c.RLock()
	defer c.RUnlock()
	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		return false
	}
	return c.Logger.ReportCaller
}
