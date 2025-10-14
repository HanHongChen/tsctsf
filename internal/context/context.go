package context

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/HanHongChen/openapi-tsctsf/models"
	"github.com/HanHongChen/tsctsf/internal/logger"
	"github.com/HanHongChen/tsctsf/pkg/factory"
	"github.com/google/uuid"
)

var tsnContext *TSCTSFContext

func Init() {

	config := factory.TsctsfConfig
	logger.InitLog.Infof("tsctsfConfig Info: Version[%s] Description[%s]\n", config.Info.Version, config.Info.Description)
	configuration := config.Configuration
	sbi := configuration.Sbi

	tsnContext = new(TSCTSFContext)
	tsnContext.NfId = uuid.New().String()
	tsnContext.NrfUri = configuration.NrfUri
	tsnContext.UriScheme = factory.TsctsfSbiDefaultScheme  // default uri scheme
	tsnContext.RegisterIPv4 = factory.TsctsfSbiDefaultIPv4 // default localhost
	tsnContext.SBIPort = factory.TsctsfSbiDefaultPort      // default port
	if sbi != nil {
		if sbi.RegisterIPv4 != "" {
			tsnContext.RegisterIPv4 = sbi.RegisterIPv4
		}
		if sbi.Port != 0 {
			tsnContext.SBIPort = sbi.Port
		}

		if sbi.Scheme == "https" {
			tsnContext.UriScheme = models.UriScheme_HTTPS
		} else {
			tsnContext.UriScheme = models.UriScheme_HTTP
		}

		tsnContext.BindingIPv4 = os.Getenv(sbi.BindingIPv4)
		if tsnContext.BindingIPv4 != "" {
			logger.InitLog.Info("Parsing ServerIPv4 address from ENV Variable.")
		} else {
			tsnContext.BindingIPv4 = sbi.BindingIPv4
			if tsnContext.BindingIPv4 == "" {
				logger.InitLog.Warn("Error parsing ServerIPv4 address as string. Using the 0.0.0.0 address as default.")
				tsnContext.BindingIPv4 = "0.0.0.0"
			}
		}
	}
	tsnContext.Url = string(tsnContext.UriScheme) + "://" + tsnContext.RegisterIPv4 + ":" + strconv.Itoa(tsnContext.SBIPort)
	serviceNameList := configuration.ServiceNameList
	tsnContext.PcfUri = configuration.PcfUri
	// context.NfService
	tsnContext.NfService = make(map[models.ServiceName]models.NfService)
	tsnContext.InitNFService(serviceNameList, config.Info.Version)
	tsnContext.Bridges = make(map[uint64]Bridge_info)
	tsnContext.SubscripSession = make(map[string]string)
	// tsnContext.NwttIndex = make(map[int]int)

	routers := config.RouterInfo // []RouterInfo
	tsnContext.Routers = make(map[string]RouterInfo)
	for _, router := range routers {
		tsnContext.Routers[router.UpNodeId] = RouterInfo{
			UpNodeId: router.UpNodeId,
			MacAddr:  router.MacAddr,
			UeIp:     router.UeIp,
			Port:     router.Port,
			Mtu:      router.Mtu,
		}
		fmt.Println(router)
	}

	detnetController := config.DetNetController
	if detnetController == (factory.DetNetController{}) {
		logger.InitLog.Warn("DetNetController is not configured")
	} else {
		tsnContext.DetNetController = DetNetController{
			Scheme:   detnetController.Scheme,
			IPv4:     detnetController.IPv4,
			Port:     detnetController.Port,
			User:     detnetController.User,
			Password: detnetController.Password,
			Uri:      detnetController.Uri,
		}
		logger.InitLog.Infof("DetNetController Info: Scheme[%s] IPv4[%s] Port[%d] User[%s] Uri[%s]\n",
			detnetController.Scheme, detnetController.IPv4, detnetController.Port, detnetController.User, detnetController.Uri)
	}
}

func (context *TSCTSFContext) InitNFService(srvNameList []string, version string) {
	tmpVersion := strings.Split(version, ".")
	versionUri := "v" + tmpVersion[0]
	for index, nameString := range srvNameList {
		name := models.ServiceName(nameString)
		context.NfService[name] = models.NfService{
			ServiceInstanceId: strconv.Itoa(index),
			ServiceName:       name,
			Versions: &[]models.NfServiceVersion{
				{
					ApiFullVersion:  version,
					ApiVersionInUri: versionUri,
				},
			},
			Scheme:          context.UriScheme,
			NfServiceStatus: models.NfServiceStatus_REGISTERED,
			ApiPrefix:       context.Url,
			IpEndPoints: &[]models.IpEndPoint{
				{
					Ipv4Address: context.RegisterIPv4,
					Transport:   models.TransportProtocol_TCP,
					Port:        int32(context.SBIPort),
				},
			},
		}
	}
}

type TSCTSFContext struct {
	NfId             string
	SBIPort          int
	RegisterIPv4     string
	BindingIPv4      string
	Url              string
	UriScheme        models.UriScheme
	NrfUri           string
	NfService        map[models.ServiceName]models.NfService
	AppSessionIdPool sync.Map
	Bridges          map[uint64]Bridge_info // key is Bridge_ID
	SubscripSession  map[string]string      // key is Session_ID
	Routers          map[string]RouterInfo
	DetNetController DetNetController
	PcfUri           string
	// NwttIndex        map[int]int
}

type RouterInfo struct {
	UpNodeId string
	MacAddr  string
	UeIp     string
	Port     int
	Mtu      int
}

type DetNetController struct {
	Scheme   string
	IPv4     string
	Port     int
	User     string
	Password string
	Uri      string
}

type Bridge_info struct {
	Bridge_Id      uint64
	Total_port_num int
	Nwtt_port_num  uint32
	UpNode_info    UpNode_info
	Port_list      []uint32
	Nwtt_ports     map[uint32]Nwtt_port_info //key is nwtt port number
	Dstt_ports     map[uint32]Dstt_port_info //key is dstt port number
	// Bridge_delay             uint32
	// Tsn_traffic_class        uint8
	Ingress_port_num uint8
	Egress_port_num  uint8
	Register         bool
}

type Dstt_port_info struct {
	// Register                 bool
	Update bool
	// Ingress                  bool
	// Nwtt_port_num uint32
	PTPInstanceId          uint16
	Dstt_addr              string
	Ue_dstt_residence_time [8]uint8

	TxPropagationDelay            uint32
	Traffic_class_table           []uint8
	Traffic_class_num             int
	SupportedPTPInstanceTypes     []uint8
	SupportedTransportTypes       []uint8
	SupportedDelayMechanisms      []uint8
	PTPGrandmasterCapable         bool
	GPTPGrandmasterCapable        bool
	SupportedPTPProfiles          []uint8
	NumberOfSupportedPTPInstances uint16
	PTPInstanceList               []uint8
}

type UpNode_info struct {
	PTPInstanceId                 uint16
	Update                        bool
	UpNodeAddr                    net.HardwareAddr
	UpNodeID                      uint64 //same as bridge_id
	SupportedPTPInstanceTypes     []uint8
	SupportedTransportTypes       []uint8
	SupportedDelayMechanisms      []uint8
	PTPGrandmasterCapable         bool
	GPTPGrandmasterCapable        bool
	SupportedPTPProfiles          []uint8
	NumberOfSupportedPTPInstances uint16
	PTPInstanceList               []uint8
}

// TODO: should be port etc data....
type Nwtt_port_info struct {
	Update                        bool
	PTPInstanceId                 uint16
	TxPropagationDelay            uint32
	Traffic_class_table           []uint8
	Traffic_class_num             int
	SupportedPTPInstanceTypes     []uint8
	SupportedTransportTypes       []uint8
	SupportedDelayMechanisms      []uint8
	PTPGrandmasterCapable         bool
	GPTPGrandmasterCapable        bool
	SupportedPTPProfiles          []uint8
	NumberOfSupportedPTPInstances uint16
	PTPInstanceList               []uint8
}

// Create new TSN context
func GetSelf() *TSCTSFContext {
	return tsnContext
}

func (a *TSCTSFContext) GetSelfID() string {
	return a.NfId
}
