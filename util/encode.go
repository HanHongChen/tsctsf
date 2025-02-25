package util

import (
	"encoding/binary"

	"github.com/HanHongChen/bitbucket-openapi/models"
	tsctsf_context "github.com/HanHongChen/tsctsf/internal/context"
)

const (
	SUPPORT   bool = true
	UNSUPPORT bool = false
)

// Ethernet port management service message type definitions.
const (
	ManagePortCommand            uint8 = 0b00000001
	ManagePortComplete           uint8 = 0b00000010
	PortManagementNotify         uint8 = 0b00000011
	PortManagementNotifyAck      uint8 = 0b00000100
	PortManagementNotifyComplete uint8 = 0b00000101
	PortManagementCapability     uint8 = 0b00000110
)

// User Plane Node management service message type definitions.
const (
	ManageUserPlaneNodeCommand       uint8 = 0b00000001
	ManageUserPlaneNodeComplete      uint8 = 0b00000010
	UserPlaneNodeManagementNotify    uint8 = 0b00000011
	UserPlaneNodeManagementNotifyAck uint8 = 0b00000100
)

// IEI of MANAGE PORT COMPLETE message content
const (
	PortManagementCapabilityIEI uint8 = 70
	PortStatusIEI               uint8 = 71
	PortUpdateResultIEI         uint8 = 72
)

// IEI of MANAGE User Plane Node COMPLETE message content
const (
	UserPlaneNodeManagementCapability uint8 = 70
	UserPlaneNodeStatus               uint8 = 71
	UserPlaneNodeUpdateResult         uint8 = 72
)

// Operation code
const (
	GetCapabilities                  uint8 = 0b00000001
	ReadParameter                    uint8 = 0b00000010
	SetParameter                     uint8 = 0b00000011
	SubscribeNotifyForParameter      uint8 = 0b00000100
	UnsubscribeForParameter          uint8 = 0b00000101
	SelevtiveReadParameter           uint8 = 0b00000110
	SelevtiveSubscribeForParameter   uint8 = 0b00000111
	SelevtiveUnsubscribeForParameter uint8 = 0b00001000
	DeleteParameterEntry             uint8 = 0b00001001
)

// PMIC
const (
	TSNTimeDomainNumber uint16 = 0x00D4
)

// Supported PTP Instance Types
const (
	OrdinaryClock       uint8 = 0x00
	BoundaryClock       uint8 = 0x01
	P2PTransparentClock uint8 = 0x02
	E2ETransparentClock uint8 = 0x03
	P2PRelayInstance    uint8 = 0x04
)

var ptpInstanceTypesMap = map[string]uint8{
	"OrdinaryClock":      OrdinaryClock,
	"BOUNDARY_CLOCK":     BoundaryClock,
	"P2P_TRANS_CLOCK":    P2PTransparentClock,
	"E2E_TRANS_CLOCK":    E2ETransparentClock,
	"P2P_RELAY_INSTANCE": P2PRelayInstance,
}

// Supported transport types
const (
	IPv4     uint8 = 0b00000000
	IPv6     uint8 = 0b00000001
	Ethernet uint8 = 0b00000010
)

var transportTypeMap = map[string]uint8{
	"IPV4": IPv4,
	"IPV6": IPv6,
	"ETH":  Ethernet,
}

// Supported PTP delay mechanisms
const (
	E2E          uint8 = 0x01
	P2P          uint8 = 0x02
	COMMON_P2P   uint8 = 0x03
	SPECIAL      uint8 = 0x04
	NO_MECHANISM uint8 = 0xFE
)

var gmEnbleMap = map[bool]uint8{
	true:  1,
	false: 0,
}

// Supported PTP profile
const (
	SMPTE               uint8 = 0b00000000
	IEEE8021AS          uint8 = 0b00000001
	E2EDefault          uint8 = 0b00000010 // Default delay request-response profile
	P2PDefault          uint8 = 0b00000011 // Default delay peer-to-peer delay profile
	HighAccuracyDefault uint8 = 0b00000100 // High Accuracy Delay Request-Response Default PTP profile
)

var ptpProfileMap = map[string]uint8{
	"SMPTE":               SMPTE,
	"IEEE8021AS":          IEEE8021AS,
	"E2EDefault":          E2EDefault,
	"P2PDefault":          P2PDefault,
	"HighAccuracyDefault": HighAccuracyDefault,
}

type PortIdentity struct {
	ClockIdentity [8]uint8
	PortNumber    uint16
}

const (
	INITIALIZING uint8 = iota + 1
	FAULTY
	DISABLED
	LISTENING
	PRE_MASTER
	MASTER
	PASSIVE
	UNCALIBRATED
	SLAVE
)

// Supported PTP Instance List
type PortDataSet struct {
	PortIdentity            PortIdentity
	PortState               uint8
	LogMinDelayReqInterval  int8
	LogAnnounceInterval     int8
	AnnounceReceiptTimeout  uint8
	LogSyncInterval         int8
	DelayMechanism          uint8
	LogMinPdelayReqInterval int8
	VersionNumber           uint8
	DelayAsymmetry          int64
	PortEnable              bool
}

// PTP instance specification
const (
	PTP_profile                                    uint16 = 0x0001
	Transport_type                                 uint16 = 0x0002
	Grandmaster_enabled                            uint16 = 0x0003
	Grandmaster_on_behalf_of_DSTT_enabled          uint16 = 0x0004
	Grandmaster_candidate_enabled                  uint16 = 0x0005
	DefaultDS_clockIdentity                        uint16 = 0x0006
	DefaultDS_clockQuality_clockClass              uint16 = 0x0007
	DefaultDS_clockQuality_clockAccuracy           uint16 = 0x0008
	DefaultDS_clockQuality_offsetScaledLogVariance uint16 = 0x0009
	DefaultDS_priority1                            uint16 = 0x000A
	DefaultDS_priority2                            uint16 = 0x000B
	DefaultDS_domainNumber                         uint16 = 0x000C
	DefaultDS_sdoId                                uint16 = 0x000D
	DefaultDS_instanceEnable                       uint16 = 0x000E
	DefaultDS_externalPortConfigurationEnabled     uint16 = 0x000F
	DefaultDS_instanceType                         uint16 = 0x0010
	PortDS_PortIdentity                            uint16 = 0x0011
	PortDS_PortState                               uint16 = 0x0012
	PortDS_LogMinDelayReqInterval                  uint16 = 0x0013
	PortDS_LogAnnounceInterval                     uint16 = 0x0014
	PortDS_AnnounceReceiptTimeout                  uint16 = 0x0015
	PortDS_LogSyncInterval                         uint16 = 0x0016
	PortDS_DelayMechanism                          uint16 = 0x0017
	PortDS_LogMinPdelayReqInterval                 uint16 = 0x0018
	PortDS_VersionNumber                           uint16 = 0x0019
	PortDS_MinorVersionNumber                      uint16 = 0x001A
	PortDS_DelayAsymmetry                          uint16 = 0x001B
	PortDS_PortEnable                              uint16 = 0x001C
)

// UMIC
/* Information for 5GS Bridge(Read only) */
const (
	UserPlaneNodeAddress uint16 = 0x0001
	UserPlaneNodeID      uint16 = 0x0003
	NWTTPortNumbers      uint16 = 0x0004
)

/* Time synchronization information(Read only) */
const (
	SupportedPTPInstanceTypes           uint16 = 0x0074
	SupportedTransportTypes             uint16 = 0x0075
	SupportedDelayMechanisms            uint16 = 0x0076
	PTPGrandmasterCapable               uint16 = 0x0077
	gPTPGrandmasterCapable              uint16 = 0x0078
	SupportedPTPProfiles                uint16 = 0x0079
	NumberOfSupportedPTPInstances       uint16 = 0x007A
	DSTTPortTimeSynchronizationInfoList uint16 = 0x007B
	UMICPTPInstanceSpecification        uint16 = 0x007C
	PMICPTPInstanceList                 uint16 = 0x00E9
)

func CreatePTPInstanceListForPMIC(ptpinfo models.TimeSyncExposureConfig, ptpInstanceID uint16, dstt_or_nwtt bool) models.PortManagementContainer {
	ptpInstance := []byte{}
	parameterName := make([]byte, 2)
	buffer := []byte{}

	// TODO: support multiple ptp instances

	/*  PTP instance parameters list */
	if dstt_or_nwtt { //nwtt
		/* TODO: */
		/* portDS.PortIdentity */
		/* portDS.PortState */
		/* portDS.LogMinDelayReqInterval */
		/* portDS.LogAnnounceInterval */
		/* portDS.AnnounceReceiptTimeout */
		/* portDS.LogSyncInterval */
		/* portDS.DelayMechanism */
		/* portDS.LogMinPdelayReqInterval */
		/* portDS.VersionNumber */
		/* portDS.MinorVersionNumber */
		/* portDS.DelayAsymmetry */
		/* portDS.PortEnable */
	} else { //dstt
		len := make([]byte, 2)
		/* PTP profile */
		if val, ok := ptpProfileMap[ptpinfo.ReqPtpIns.PtpProfile]; ok {
			binary.BigEndian.PutUint16(parameterName, PTP_profile)

			/* PTP instance parameter  */
			buffer = append(buffer, parameterName...)
			binary.BigEndian.PutUint16(len, 1)
			buffer = append(buffer, len...)
			buffer = append(buffer, val)
		}
		/* Transport type */
		if val, ok := transportTypeMap[string(ptpinfo.ReqPtpIns.Protocol)]; ok {
			binary.BigEndian.PutUint16(parameterName, Transport_type)

			/* PTP instance parameter  */
			buffer = append(buffer, parameterName...)
			binary.BigEndian.PutUint16(len, 1)
			buffer = append(buffer, len...)
			buffer = append(buffer, val)
		}
		/* Grandmaster enable */
		if ptpinfo.ReqPtpIns.InstanceType == "BOUNDARY_CLOCK" {
			binary.BigEndian.PutUint16(parameterName, Grandmaster_enabled)

			/* PTP instance parameter  */
			buffer = append(buffer, parameterName...)
			binary.BigEndian.PutUint16(len, 1)
			buffer = append(buffer, len...)
			buffer = append(buffer, 00000001) //true
		} else {
			binary.BigEndian.PutUint16(parameterName, Grandmaster_enabled)

			/* PTP instance parameter  */
			buffer = append(buffer, parameterName...)
			binary.BigEndian.PutUint16(len, 1)
			buffer = append(buffer, len...)
			buffer = append(buffer, 00000000) //false
		}

		/* TODO: */
		/* defaultDS.clockIdentity */
		/* defaultDS.clockQuality.clockClass */
		/* defaultDS.clockQuality.clockAccurac */
		/* defaultDS.clockQuality.offsetScaled */
		/* defaultDS.priority1 */
		/* defaultDS.priority2 */
		/* defaultDS.domainNumber */
		/* defaultDS.sdoId */
		/* defaultDS.instanceEnable */
		/* defaultDS.instanceType */
		/* portDS.PortIdentity */
		/* portDS.PortState */
		/* portDS.LogMinDelayReqInterval */
		/* portDS.LogAnnounceInterval */
		/* portDS.LogSyncInterval */
		/* portDS.DelayMechanism */
		/* portDS.LogMinPdelayReqInterval */
		/* portDS.VersionNumber */
		/* portDS.MinorVersionNumber */
		/* portDS.DelayAsymmetry */
		/* portDS.PortEnable */
	}

	ptpID := make([]byte, 2)
	binary.BigEndian.PutUint16(ptpID, ptpInstanceID)
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(buffer)))

	/* PTP instance */
	ptpInstance = append(ptpInstance, length...)

	ptpInstance = append(ptpInstance, ptpID...)

	ptpInstance = append(ptpInstance, buffer...)

	listLength := make([]byte, 2)
	binary.BigEndian.PutUint16(listLength, uint16(len(ptpInstance)))
	ptpInstanceList := []byte{}
	/* PTP instance list information element */
	ptpInstanceIEI := make([]byte, 2)
	binary.BigEndian.PutUint16(ptpInstanceIEI, PMICPTPInstanceList)

	ptpInstanceList = append(ptpInstanceList, ptpInstanceIEI...) // IEI

	ptpInstanceList = append(ptpInstanceList, listLength...) // Length

	ptpInstanceList = append(ptpInstanceList, ptpInstance...) // Content

	var pmicByte []byte

	// 9.1 port management service message type
	pmicByte = append(pmicByte, ManagePortCommand)

	// 9.2 port management list
	// iei (nothing)

	Len := IntToBytes(len(ptpInstanceList) + 1)
	pmicByte = append(pmicByte, Len[0], Len[1])

	pmicByte = append(pmicByte, SetParameter)
	// 9.15 ptp intance list
	pmicByte = append(pmicByte, ptpInstanceList...)

	var pmic models.PortManagementContainer

	tsctsf_self := tsctsf_context.GetSelf()

	if dstt_or_nwtt { //nwtt
		for key := range tsctsf_self.Bridges[ptpinfo.UpNodeId].Nwtt_ports {
			pmic.PortManCont = pmicByte
			pmic.PortNum = key
		}
	} else { //dstt
		for key := range tsctsf_self.Bridges[ptpinfo.UpNodeId].Dstt_ports {
			pmic.PortManCont = pmicByte
			pmic.PortNum = key
		}
	}

	return pmic
}

func CreatePTPInstanceListForUMIC(ptpinfo models.TimeSyncExposureConfig, ptpInstanceID uint16) models.BridgeManagementContainer {
	ptpInstance := []byte{}
	parameterName := make([]byte, 2)
	ptpID := make([]byte, 2)
	// TODO: support multiple ptpinstance
	binary.BigEndian.PutUint16(ptpID, ptpInstanceID)
	buffer := []byte{}

	valueLength := make([]byte, 2)
	/* PTP profile */
	if val, ok := ptpProfileMap[ptpinfo.ReqPtpIns.PtpProfile]; ok {
		binary.BigEndian.PutUint16(parameterName, PTP_profile)

		/* PTP instance parameter  */
		buffer = append(buffer, parameterName...)
		binary.BigEndian.PutUint16(valueLength, 1)
		buffer = append(buffer, valueLength...)
		buffer = append(buffer, val)
	}
	/* Transport type */
	if val, ok := transportTypeMap[string(ptpinfo.ReqPtpIns.Protocol)]; ok {
		binary.BigEndian.PutUint16(parameterName, Transport_type)

		/* PTP instance parameter  */
		buffer = append(buffer, parameterName...)
		binary.BigEndian.PutUint16(valueLength, 1)
		buffer = append(buffer, valueLength...)
		buffer = append(buffer, val)
	}
	dsttPortTSinfoList := []byte{}

	/* Grandmaster candidate enabled
	k.2.2.4 Configuration for PTP grandmaster function
	To enable either option a) or option b) for a PTP instance, the TSN AF or TSCTSF sets the element "Grandmaster candidate enabled" TRUE (per PTP instance) in UMIC
	*/
	if ptpinfo.ReqPtpIns.InstanceType == "BOUNDARY_CLOCK" {
		binary.BigEndian.PutUint16(parameterName, Grandmaster_candidate_enabled)
		/* PTP instance parameter  */
		buffer = append(buffer, parameterName...)
		binary.BigEndian.PutUint16(valueLength, 1)
		buffer = append(buffer, valueLength...)
		buffer = append(buffer, 00000001) //true

		tsctsf_self := tsctsf_context.GetSelf()
		var dsttPortNum uint32
		for _, bridge := range tsctsf_self.Bridges {
			for port, info := range bridge.Dstt_ports {
				if info.PTPInstanceId == ptpInstanceID {
					dsttPortNum = port
					/* CREATE DS-TT port time synchronization information list */
					dsttPortTSinfoList = CreateDSTTPortTimeSynhronizationInfoList(dsttPortNum, ptpInstanceID)
				}
			}
		}

	} else {
		binary.BigEndian.PutUint16(parameterName, Grandmaster_candidate_enabled)
		/* PTP instance parameter  */
		buffer = append(buffer, parameterName...)
		binary.BigEndian.PutUint16(valueLength, 1)
		buffer = append(buffer, valueLength...)
		buffer = append(buffer, 00000000) //false

		dsttPortTSinfoList = nil
	}

	/* defaultDS.instanceType */
	if val, ok := ptpInstanceTypesMap[string(ptpinfo.ReqPtpIns.InstanceType)]; ok {
		binary.BigEndian.PutUint16(parameterName, DefaultDS_instanceType)

		/* PTP instance parameter  */
		buffer = append(buffer, parameterName...)
		binary.BigEndian.PutUint16(valueLength, 1)
		buffer = append(buffer, valueLength...)
		buffer = append(buffer, val)
	}

	/* TODO: */
	/* defaultDS.clockIdentity */
	/* defaultDS.clockQuality.clockClass */
	/* defaultDS.clockQuality.clockAccurac */
	/* defaultDS.clockQuality.offsetScaled */
	/* defaultDS.priority1 */
	/* defaultDS.priority2 */
	/* defaultDS.domainNumber */
	/* defaultDS.sdoId */
	/* defaultDS.instanceEnable */
	/* defaultDS.externalPortConfiguration */
	/* defaultDS.instanceType */

	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(buffer)))

	/* PTP instance */
	ptpInstance = append(ptpInstance, length...)

	ptpInstance = append(ptpInstance, ptpID...)

	ptpInstance = append(ptpInstance, buffer...)

	listLength := make([]byte, 2)
	binary.BigEndian.PutUint16(listLength, uint16(len(ptpInstance)))
	ptpInstanceList := []byte{}
	/* PTP instance list information element */
	ptpInstanceIEI := make([]byte, 2)
	binary.BigEndian.PutUint16(ptpInstanceIEI, UMICPTPInstanceSpecification)

	ptpInstanceList = append(ptpInstanceList, ptpInstanceIEI...) // IEI

	ptpInstanceList = append(ptpInstanceList, listLength...) // Length

	ptpInstanceList = append(ptpInstanceList, ptpInstance...) // Content

	var umicByte []byte

	// 9.1 port management service message type
	umicByte = append(umicByte, ManageUserPlaneNodeCommand)

	// 9.2 port management list
	// iei (nothing)
	lenUmic := len(ptpInstanceList) + 1
	if dsttPortTSinfoList != nil {
		lenUmic += len(dsttPortTSinfoList) + 1
	}
	LenUmic := IntToBytes(lenUmic)
	umicByte = append(umicByte, LenUmic[0], LenUmic[1])

	umicByte = append(umicByte, SetParameter)
	// 9.15 ptp intance list
	umicByte = append(umicByte, ptpInstanceList...)

	// 9.16 DS-TT port time synchronization information list
	if dsttPortTSinfoList != nil {
		umicByte = append(umicByte, SetParameter)
		umicByte = append(umicByte, dsttPortTSinfoList...)
	}
	var umic models.BridgeManagementContainer
	umic.BridgeManCont = umicByte
	return umic
}

func CreateDSTTPortTimeSynhronizationInfoList(dsttPortNum uint32, ptpInstanceID uint16) []byte {
	ptpInstance := []byte{}
	parameterName := make([]byte, 2)
	ptpID := make([]byte, 2)
	// TODO: support multiple ptpinstance
	binary.BigEndian.PutUint16(ptpID, ptpInstanceID)
	buffer := []byte{}
	length := make([]byte, 2)
	valueLength := make([]byte, 2)

	tsctsf_self := tsctsf_context.GetSelf()
	var upNodeInfo tsctsf_context.UpNode_info
	for _, bridge := range tsctsf_self.Bridges {
		if bridge.UpNode_info.PTPInstanceId == ptpInstanceID {
			upNodeInfo = bridge.UpNode_info
		}
	}
	/* Grandmaster on behalf of DSTT enabled */
	/* Grandmaster candidate enabled */
	upNodeInfo.PTPGrandmasterCapable = true
	if upNodeInfo.PTPGrandmasterCapable {
		binary.BigEndian.PutUint16(parameterName, Grandmaster_on_behalf_of_DSTT_enabled)
		/* PTP instance parameter  */
		buffer = append(buffer, parameterName...)
		binary.BigEndian.PutUint16(valueLength, 1)
		buffer = append(buffer, valueLength...)
		buffer = append(buffer, 00000001) //true

		binary.BigEndian.PutUint16(parameterName, Grandmaster_candidate_enabled)
		/* PTP instance parameter  */
		buffer = append(buffer, parameterName...)
		binary.BigEndian.PutUint16(valueLength, 1)
		buffer = append(buffer, valueLength...)
		buffer = append(buffer, 00000001) //true
	} else {
		binary.BigEndian.PutUint16(parameterName, Grandmaster_on_behalf_of_DSTT_enabled)
		/* PTP instance parameter  */
		buffer = append(buffer, parameterName...)
		binary.BigEndian.PutUint16(valueLength, 1)
		buffer = append(buffer, valueLength...)
		buffer = append(buffer, 00000000) //false

		binary.BigEndian.PutUint16(parameterName, Grandmaster_candidate_enabled)
		/* PTP instance parameter  */
		buffer = append(buffer, parameterName...)
		binary.BigEndian.PutUint16(valueLength, 1)
		buffer = append(buffer, valueLength...)
		buffer = append(buffer, 00000000) //false
	}

	/* TODO: */
	/* portDS.PortIdentity */
	/* portDS.PortState */
	/* portDS.LogMinDelayReqInterval */
	/* portDS.LogAnnounceInterval */
	/* portDS.AnnounceReceiptTimeout */
	/* portDS.LogSyncInterval */
	/* portDS.DelayMechanism */
	/* portDS.LogMinPdelayReqInterval */
	/* portDS.VersionNumber */
	/* portDS.MinorVersionNumber */
	/* portDS.DelayAsymmetry */
	/* portDS.PortEnable */

	binary.BigEndian.PutUint16(length, uint16(len(buffer)))
	/* PTP instance list*/
	ptpInstance = append(ptpInstance, length...)

	ptpInstance = append(ptpInstance, ptpID...)

	ptpInstance = append(ptpInstance, buffer...)

	dsttPortTSInfoList := []byte{}
	dsttPortTSInfoListIEI := make([]byte, 2)
	binary.BigEndian.PutUint16(dsttPortTSInfoListIEI, DSTTPortTimeSynchronizationInfoList)

	portnum := make([]byte, 2)
	binary.BigEndian.PutUint16(portnum, uint16(dsttPortNum))

	/* DS-TT port time synchronization information list IEI */
	dsttPortTSInfoList = append(dsttPortTSInfoList, dsttPortTSInfoListIEI...)
	/* Length of DS-TT port time synchronization information list contents */
	binary.BigEndian.PutUint16(length, uint16(len(ptpInstance)+len(portnum)+2))
	dsttPortTSInfoList = append(dsttPortTSInfoList, length...)
	/* DS-TT port time synchronization information */
	binary.BigEndian.PutUint16(length, uint16(len(ptpInstance)))
	dsttPortTSInfoList = append(dsttPortTSInfoList, length...)

	dsttPortTSInfoList = append(dsttPortTSInfoList, portnum...)

	dsttPortTSInfoList = append(dsttPortTSInfoList, ptpInstance...)

	return dsttPortTSInfoList
}
