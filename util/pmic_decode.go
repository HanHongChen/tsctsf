package util

import (
	"bytes"
	"encoding/binary"
	"math"
	"strconv"

	tsctsf_context "github.com/HanHongChen/tsctsf/internal/context"

	"github.com/HanHongChen/openapi-tsctsf/models"
	//github.com/HanHongChen/bitbucket-openapi/models
	"github.com/HanHongChen/tsctsf/internal/logger"
)

var (
	dstt_capability = map[string]string{
		// "1":   "txPropagationDelay",                //0001H
		// "2":   "traffic class table",               //0002H
		// "224": "Stream filter instance table",      //00E0H
		// "225": "Stream gate instance table",        //00E01H
		"226": "Supported PTP instance types",      // 0x00E2
		"227": "Supported transport types",         //0x00E3
		"228": "Supported delay mechanisms",        //0x00E4
		"229": "PTP grandmaster capable",           //0x00E5
		"230": "gPTP grandmaster capable",          //0x00E6
		"231": "Supported PTP profiles",            //0x00E7
		"232": "Number of supported PTP instances", //0x00E8
		"233": "PTP instance list",                 //0x00E9
	}

	nwtt_capability = map[string]string{
		"17": "PortDS_PortIdentity",            // 0x0011
		"18": "PortDS_PortState",               // 0x0012
		"19": "PortDS_LogMinDelayReqInterval",  // 0x0013
		"20": "PortDS_LogAnnounceInterval",     // 0x0014
		"21": "PortDS_AnnounceReceiptTimeout",  // 0x0015
		"22": "PortDS_LogSyncInterval",         // 0x0016
		"23": "PortDS_DelayMechanism",          // 0x0017
		"24": "PortDS_LogMinPdelayReqInterval", // 0x0018
		"25": "PortDS_VersionNumber",           // 0x0019
		"26": "PortDS_MinorVersionNumber",      // 0x001A
		"27": "PortDS_DelayAsymmetry",          // 0x001B
		"28": "PortDS_PortEnable",              // 0x001C
	}
)

func NWTTPMICDecodeCapabilityInfo(pmic models.PortManagementContainer) models.PortManagementContainer {
	logger.UtilLog.Tracef("NW-TT PMIC Decode ready")

	// NW-TT shows capability to TSCTSF
	logger.UtilLog.Info("Deal with Port management capability information")

	container_len := int(pmic.PortManCont[1])<<8 + int(pmic.PortManCont[2]) + 3
	var capability_list []uint16
	for i := 3; i < container_len; i = i + 2 {
		key := uint16(pmic.PortManCont[i])<<8 + uint16(pmic.PortManCont[i+1])
		value, exist := nwtt_capability[strconv.FormatUint(uint64(key), 10)]

		// if find in capability represent TSCTSF wants it
		if exist {
			logger.UtilLog.Infof("TSCTSF wants [%s] capability", value)
			capability_list = append(capability_list, key)
		} else {
			logger.UtilLog.Info("TSCTSF do not want this capability")
		}
	}

	if len(capability_list) != 0 {
		pmic.PortManCont = PMICCreation(capability_list)
	}
	return pmic
}

func DSTTPMICDecodeCapabilityInfo(pmic models.PortManagementContainer) models.PortManagementContainer {
	logger.UtilLog.Tracef("DS-TT PMIC Decode ready")

	// DS-TT shows capability to TSCTSF
	logger.UtilLog.Info("Deal with Port management capability information")

	container_len := int(pmic.PortManCont[1])<<8 + int(pmic.PortManCont[2]) + 3
	var capability_list []uint16
	for i := 3; i < container_len; i = i + 2 {
		key := uint16(pmic.PortManCont[i])<<8 + uint16(pmic.PortManCont[i+1])
		value, exist := dstt_capability[strconv.FormatUint(uint64(key), 10)]

		// if find in capability represent TSCTSF wants it
		if exist {
			logger.UtilLog.Infof("TSCTSF wants [%s] capability", value)
			capability_list = append(capability_list, key)
		} else {
			logger.UtilLog.Info("TSCTSF do not want this capability")
		}
	}

	if len(capability_list) != 0 {
		pmic.PortManCont = PMICCreation(capability_list)
	}
	return pmic
}

func TTPortResponse(pmic models.PortManagementContainer) uint64 {
	logger.UtilLog.Infof("PMIC Decode ready :[%X]", pmic.PortManCont)
	if pmic.PortManCont[0] == 2 {
		logger.UtilLog.Info("Deal with MANAGE PORT COMPLETE Message")

		capability := bytes.IndexByte(pmic.PortManCont, byte(0x70))
		if capability != -1 {
			logger.UtilLog.Info("Handle port management capability")
		}
		status := bytes.IndexByte(pmic.PortManCont, byte(0x71))
		if status != -1 {
			logger.UtilLog.Info("Handle port status")
			return DecodePortStatus(pmic, status)
		}
		update_result := bytes.IndexByte(pmic.PortManCont, byte(0x72))
		if update_result != -1 {
			logger.UtilLog.Info("Handle port update result")
			return DecodePortUpdatedStatus(pmic, update_result)

		}
	} else if pmic.PortManCont[0] == 3 {
		logger.UtilLog.Info("Deal with PORT MANAGEMENT NOTIFY Message")
	} else if pmic.PortManCont[0] == 5 {
		logger.UtilLog.Info("Deal with PORT MANAGEMENT NOTIFY COMPLETE Message")
	}
	return ^uint64(0)
}

// TODO: parameter detailed
func DecodePortStatus(pmic models.PortManagementContainer, index int) uint64 {
	var txPropagationDelay uint32
	var traffic_class_num int
	var traffic_class_table []uint8
	var supportedPTPInstanceTypes []uint8
	var supportedTransportTypes []uint8
	var supportedDelayMechanisms []uint8
	var ptpGrandmasterCapable bool
	var gptpGrandmasterCapable bool
	var supportedPTPProfiles []uint8
	var numberOfSupportedPTPInstances uint16
	var ptpInstanceList []uint8

	//portstatus_len := int(pmic.PortManCont[index+1])<<8 + int(pmic.PortManCont[index+2]) + 3
	success_read_num := int(pmic.PortManCont[index+3])
	i := index + 4
	for num := 0; num < success_read_num; num += 1 {
		parameter_name := uint16(pmic.PortManCont[i])<<8 + uint16(pmic.PortManCont[i+1])
		parameter_len := int(pmic.PortManCont[i+2])<<8 + int(pmic.PortManCont[i+3])
		switch parameter_name {
		case 1: //0001H
			i = i + 4
			data := binary.BigEndian.Uint64(pmic.PortManCont[i : i+parameter_len])
			txPropagationDelay = uint32(data / uint64(math.Pow(2, 16)))
			logger.UtilLog.Trace("txPropagationDelay = ", txPropagationDelay)
			i = i + parameter_len
		case 2: //0002H
			i = i + 4
			traffic_class_num = int(pmic.PortManCont[i+2])
			traffic_class_table = pmic.PortManCont[i : i+parameter_len]
			logger.UtilLog.Trace("traffic_class_table = ", traffic_class_table)
			i = i + parameter_len
		case 226: //0x00E2
			i = i + 4
			supportedPTPInstanceTypes = pmic.PortManCont[i : i+parameter_len]
			logger.UtilLog.Traceln("Supported PTP instance types = ", supportedPTPInstanceTypes)
			i = i + parameter_len

		case 227: //0x00E3
			i = i + 4
			supportedTransportTypes = pmic.PortManCont[i : i+parameter_len]
			logger.UtilLog.Traceln("Supported transport types = ", supportedTransportTypes)
			i = i + parameter_len

		case 228: //0x00E4
			i = i + 4
			supportedDelayMechanisms = pmic.PortManCont[i : i+parameter_len]
			logger.UtilLog.Traceln("Supported delay mechanisms = ", supportedDelayMechanisms)
			i = i + parameter_len

		case 229: //0x00E5
			i = i + 4
			data := int(pmic.PortManCont[i])
			if data == 0 {
				ptpGrandmasterCapable = false
			} else {
				ptpGrandmasterCapable = true

			}
			logger.UtilLog.Traceln("PTP grandmaster capable = ", ptpGrandmasterCapable)
			i = i + parameter_len

		case 230: //0x00E6
			i = i + 4
			data := int(pmic.PortManCont[i])
			if data == 0 {
				gptpGrandmasterCapable = false
			} else {
				gptpGrandmasterCapable = true

			}
			logger.UtilLog.Traceln("gPTP grandmaster capable = ", gptpGrandmasterCapable)
			i = i + parameter_len

		case 231: //0x00E7
			i = i + 4
			supportedPTPProfiles = pmic.PortManCont[i : i+parameter_len]
			logger.UtilLog.Traceln("Supported PTP profiles = ", supportedPTPProfiles)
			i = i + parameter_len

		case 232: //0x00E8
			i = i + 4
			numberOfSupportedPTPInstances = binary.BigEndian.Uint16(pmic.PortManCont[i : i+parameter_len])
			logger.UtilLog.Traceln("Number of supported PTP instances = ", numberOfSupportedPTPInstances)
			i = i + parameter_len

		case 233: //0x00E9
			i = i + 4
			ptpInstanceList = pmic.PortManCont[i : i+parameter_len]
			logger.UtilLog.Traceln("PTP instance list = ", ptpInstanceList)
			i = i + parameter_len
		default:
			logger.UtilLog.Info("Skip, do not need it temporarily")
			i = i + 4 + parameter_len
		}
	}
	tsctsf_self := tsctsf_context.GetSelf()
	for bridge_id, Bridge_info := range tsctsf_self.Bridges {
		nwtt_port_info, exist := Bridge_info.Nwtt_ports[pmic.PortNum]
		if exist {
			if txPropagationDelay != 0 && nwtt_port_info.TxPropagationDelay != txPropagationDelay {
				nwtt_port_info.TxPropagationDelay = txPropagationDelay
			}
			if traffic_class_num != 0 && nwtt_port_info.Traffic_class_num != traffic_class_num {
				nwtt_port_info.Traffic_class_table = traffic_class_table
				nwtt_port_info.Traffic_class_num = traffic_class_num
			}
			if supportedPTPInstanceTypes != nil {
				nwtt_port_info.SupportedPTPInstanceTypes = supportedPTPInstanceTypes
			}
			if supportedTransportTypes != nil {
				nwtt_port_info.SupportedTransportTypes = supportedTransportTypes
			}
			if supportedDelayMechanisms != nil {
				nwtt_port_info.SupportedDelayMechanisms = supportedDelayMechanisms
			}
			nwtt_port_info.PTPGrandmasterCapable = ptpGrandmasterCapable
			nwtt_port_info.GPTPGrandmasterCapable = gptpGrandmasterCapable

			if supportedPTPProfiles != nil {
				nwtt_port_info.SupportedPTPProfiles = supportedPTPProfiles
			}
			if numberOfSupportedPTPInstances != 0 {
				nwtt_port_info.NumberOfSupportedPTPInstances = numberOfSupportedPTPInstances
			}
			if ptpInstanceList != nil {
				nwtt_port_info.PTPInstanceList = ptpInstanceList
			}
			nwtt_port_info.Update = true
			Bridge_info.Nwtt_ports[pmic.PortNum] = nwtt_port_info
			tsctsf_self.Bridges[bridge_id] = Bridge_info
			return bridge_id
		}
	}

	for bridge_id, Bridge_info := range tsctsf_self.Bridges {
		dstt_port_info, exist := Bridge_info.Dstt_ports[pmic.PortNum]
		if exist {
			if txPropagationDelay != 0 && dstt_port_info.TxPropagationDelay != txPropagationDelay {
				dstt_port_info.TxPropagationDelay = txPropagationDelay
			}
			if traffic_class_num != 0 && dstt_port_info.Traffic_class_num != traffic_class_num {
				dstt_port_info.Traffic_class_table = traffic_class_table
				dstt_port_info.Traffic_class_num = traffic_class_num
			}
			if supportedPTPInstanceTypes != nil {
				dstt_port_info.SupportedPTPInstanceTypes = supportedPTPInstanceTypes
			}
			if supportedTransportTypes != nil {
				dstt_port_info.SupportedTransportTypes = supportedTransportTypes
			}
			if supportedDelayMechanisms != nil {
				dstt_port_info.SupportedDelayMechanisms = supportedDelayMechanisms
			}
			dstt_port_info.PTPGrandmasterCapable = ptpGrandmasterCapable
			dstt_port_info.GPTPGrandmasterCapable = gptpGrandmasterCapable

			if supportedPTPProfiles != nil {
				dstt_port_info.SupportedPTPProfiles = supportedPTPProfiles
			}
			if numberOfSupportedPTPInstances != 0 {
				dstt_port_info.NumberOfSupportedPTPInstances = numberOfSupportedPTPInstances
			}
			if ptpInstanceList != nil {
				dstt_port_info.PTPInstanceList = ptpInstanceList
			}
			dstt_port_info.Update = true
			Bridge_info.Dstt_ports[pmic.PortNum] = dstt_port_info
			tsctsf_self.Bridges[bridge_id] = Bridge_info
			return bridge_id
		}
	}
	return ^uint64(0)
}

// TODO: other parameter
func DecodePortUpdatedStatus(pmic models.PortManagementContainer, index int) uint64 {
	var ptpInstanceList []uint8

	success_set_num := int(pmic.PortManCont[index+3])
	i := index + 4
	for num := 0; num < success_set_num; num += 1 {
		parameter_name := uint16(pmic.PortManCont[i])<<8 + uint16(pmic.PortManCont[i+1])
		parameter_len := int(pmic.PortManCont[i+2])<<8 + int(pmic.PortManCont[i+3])
		switch parameter_name {
		case 233: //ptp instance list
			i = i + 4
			ptpInstanceList = pmic.PortManCont[i : i+parameter_len]
			logger.UtilLog.Traceln("PTP instance list = ", ptpInstanceList)
			i = i + parameter_len

		default:
			logger.UtilLog.Info("Skip, do not need it temporarily")
			i = i + 4 + parameter_len
		}
	}
	tsctsf_self := tsctsf_context.GetSelf()

	for bridge_id, Bridge_info := range tsctsf_self.Bridges {
		nwtt_port_info, exist := Bridge_info.Nwtt_ports[pmic.PortNum]
		if exist {
			if ptpInstanceList != nil {
				nwtt_port_info.PTPInstanceList = ptpInstanceList
			}
			nwtt_port_info.Update = true
			Bridge_info.Nwtt_ports[pmic.PortNum] = nwtt_port_info
			tsctsf_self.Bridges[bridge_id] = Bridge_info
			return bridge_id
		}
	}

	for bridge_id, Bridge_info := range tsctsf_self.Bridges {
		dstt_port_info, exist := Bridge_info.Dstt_ports[pmic.PortNum]
		if exist {
			if ptpInstanceList != nil {
				dstt_port_info.PTPInstanceList = ptpInstanceList
			}
			dstt_port_info.Update = true
			Bridge_info.Dstt_ports[pmic.PortNum] = dstt_port_info
			tsctsf_self.Bridges[bridge_id] = Bridge_info
			return bridge_id
		}
	}
	return ^uint64(0)
}

// TS24.519 5.2.1 Network-requested Port management procedure PMIC creation on TSCTSF
func PMICCreation(capability_list []uint16) []byte {
	var pmic []byte
	list_len := len(capability_list)

	// 9.1 port management service message type
	pmic = append(pmic, 0b00000001)

	// 9.2 port management list
	// iei (nothing)

	// pmic length : 2 bytes
	len := list_len * 3
	Len := IntToBytes(len)
	pmic = append(pmic, Len[0], Len[1])

	// read capability parameters TSCTSF needed from DS-TT/NW-TT
	for i := 0; i < list_len; i = i + 1 {
		pmic = append(pmic, 0b00000010)
		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, capability_list[i])
		pmic = append(pmic, b[0], b[1])
	}

	return pmic
}

func IntToBytes(n int) []byte {
	data := uint16(n)
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}

func BytesTo32Uint(b []byte) uint32 {
	bytebuf := bytes.NewBuffer(b)
	var value uint32
	binary.Read(bytebuf, binary.BigEndian, &value)
	return value
}
