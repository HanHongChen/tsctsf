package util

import (
	"bytes"
	"encoding/binary"
	"net"

	tsctsf_context "github.com/HanHongChen/tsctsf/internal/context"

	"github.com/HanHongChen/bitbucket-openapi/models"
	"github.com/HanHongChen/tsctsf/internal/logger"
)

var (
	umic_capability = map[string]string{
		"1":   "User plane node Address",             // 0x0001
		"3":   "User plane node ID",                  // 0x0003
		"4":   "NW-TT port numbers",                  // 0x0004
		"116": "Supported PTP instance types",        //0x0074
		"117": "Supported transport types",           //0x0075
		"118": "Supported delay mechanisms",          //0x0076
		"119": "PTP grandmaster capable",             //0x0077
		"120": "gPTP grandmaster capable",            //0x0078
		"121": "Supported PTP profiles",              //0x0079
		"122": "Number of supported PTP instances",   //0x007A
		"123": "DSTTPortTimeSynchronizationInfoList", //0x007B
		"124": "PTPInstanceSpecification",            //0x007C

	}
)

func UpNodeResponse(umic models.BridgeManagementContainer) uint64 {
	logger.UtilLog.Infof("UMIC Decode ready :[%x]", umic.BridgeManCont)
	if umic.BridgeManCont[0] == 2 {
		logger.UtilLog.Info("Deal with MANAGE UpNode COMPLETE Message")

		capability := bytes.IndexByte(umic.BridgeManCont, byte(0x70))
		if capability != -1 {
			logger.UtilLog.Info("Handle UpNode management capability")
		}
		status := bytes.IndexByte(umic.BridgeManCont, byte(0x71))
		if status != -1 {
			logger.UtilLog.Info("Handle UpNode status")
			return DecodeUpNodeStatus(umic, status)
		}
		update_result := bytes.IndexByte(umic.BridgeManCont, 0x48)
		if update_result != -1 {
			logger.UtilLog.Info("Handle UpNode update result")
			return DecodeUpNodeUpdatedStatus(umic, update_result)
		}
	} else if umic.BridgeManCont[0] == 3 {
		logger.UtilLog.Info("Deal with UpNode MANAGEMENT NOTIFY Message")
	} else if umic.BridgeManCont[0] == 5 {
		logger.UtilLog.Info("Deal with UpNode MANAGEMENT NOTIFY COMPLETE Message")
	}
	return ^uint64(0)
}

func DecodeUpNodeStatus(umic models.BridgeManagementContainer, index int) uint64 {
	var upNodeAddr net.HardwareAddr
	var upNodeID uint64
	// var nwttPortNumber
	var supportedPTPInstanceTypes []uint8
	var supportedTransportTypes []uint8
	var supportedDelayMechanisms []uint8
	var ptpGrandmasterCapable bool
	var gptpGrandmasterCapable bool
	var supportedPTPProfiles []uint8
	var numberOfSupportedPTPInstances uint16
	var ptpInstanceList []uint8
	tsctsf_self := tsctsf_context.GetSelf()

	success_read_num := int(umic.BridgeManCont[index+3])
	i := index + 4
	for num := 0; num < success_read_num; num += 1 {
		parameter_name := uint16(umic.BridgeManCont[i])<<8 + uint16(umic.BridgeManCont[i+1])
		parameter_len := int(umic.BridgeManCont[i+2])<<8 + int(umic.BridgeManCont[i+3])
		switch parameter_name {
		case 1: // 0x0001
			i = i + 4
			upNodeAddr, _ = net.ParseMAC(string(umic.BridgeManCont[i : i+parameter_len]))
			logger.UtilLog.Traceln("User Plane Node Addr = ", upNodeAddr)
			i = i + parameter_len
		case 3: // 0x0003
			i = i + 4
			upNodeID = binary.BigEndian.Uint64(umic.BridgeManCont[i : i+parameter_len])
			logger.UtilLog.Traceln("User Plane Node ID = ", upNodeID)
			i = i + parameter_len
		case 4: // 0x0004
			i = i + 4
			logger.UtilLog.Traceln("[TODO] NW-TT port numbers = ")
			i = i + parameter_len
		case 116: //0x0074
			i = i + 4
			supportedPTPInstanceTypes = umic.BridgeManCont[i : i+parameter_len]
			logger.UtilLog.Traceln("Supported PTP instance types = ", supportedPTPInstanceTypes)
			i = i + parameter_len

		case 117: //0x0075
			i = i + 4
			supportedTransportTypes = umic.BridgeManCont[i : i+parameter_len]
			logger.UtilLog.Traceln("Supported transport types = ", supportedTransportTypes)
			i = i + parameter_len

		case 118: //0x0076
			i = i + 4
			supportedDelayMechanisms = umic.BridgeManCont[i : i+parameter_len]
			logger.UtilLog.Traceln("Supported delay mechanisms = ", supportedDelayMechanisms)
			i = i + parameter_len

		case 119: //0x0077
			i = i + 4
			data := int(umic.BridgeManCont[i])
			if data == 0 {
				ptpGrandmasterCapable = false
			} else {
				ptpGrandmasterCapable = true

			}
			logger.UtilLog.Traceln("PTP grandmaster capable = ", ptpGrandmasterCapable)
			i = i + parameter_len

		case 120: //0x0078
			i = i + 4
			data := int(umic.BridgeManCont[i])
			if data == 0 {
				gptpGrandmasterCapable = false
			} else {
				gptpGrandmasterCapable = true

			}
			logger.UtilLog.Traceln("gPTP grandmaster capable = ", gptpGrandmasterCapable)
			i = i + parameter_len

		case 121: //0x0079
			i = i + 4
			supportedPTPProfiles = umic.BridgeManCont[i : i+parameter_len]
			logger.UtilLog.Traceln("Supported PTP profiles = ", supportedPTPProfiles)
			i = i + parameter_len

		case 122: //0x007A
			i = i + 4
			numberOfSupportedPTPInstances = binary.BigEndian.Uint16(umic.BridgeManCont[i : i+parameter_len])
			logger.UtilLog.Traceln("Number of supported PTP instances = ", numberOfSupportedPTPInstances)
			i = i + parameter_len

		case 123: //0x007B
			i = i + 4
			ptpInstanceList = umic.BridgeManCont[i : i+parameter_len]
			logger.UtilLog.Traceln("[TODO] DS-TT port time synchronization information list = ")
			i = i + parameter_len
		case 124: //0x007c
			i = i + 4
			ptpInstanceList = umic.BridgeManCont[i : i+parameter_len]
			logger.UtilLog.Traceln("[TODO] PTP instance specification = ")
			i = i + parameter_len
		default:
			logger.UtilLog.Info("Skip, do not need it temporarily")
			i = i + 4 + parameter_len
		}
	}

	var newupNodeInfo tsctsf_context.UpNode_info
	newupNodeInfo.Update = true

	if supportedPTPInstanceTypes != nil {
		newupNodeInfo.SupportedPTPInstanceTypes = supportedPTPInstanceTypes
	}
	if supportedTransportTypes != nil {
		newupNodeInfo.SupportedTransportTypes = supportedTransportTypes
	}
	if supportedDelayMechanisms != nil {
		newupNodeInfo.SupportedDelayMechanisms = supportedDelayMechanisms
	}
	newupNodeInfo.PTPGrandmasterCapable = ptpGrandmasterCapable
	newupNodeInfo.GPTPGrandmasterCapable = gptpGrandmasterCapable

	if supportedPTPProfiles != nil {
		newupNodeInfo.SupportedPTPProfiles = supportedPTPProfiles
	}
	if numberOfSupportedPTPInstances != 0 {
		newupNodeInfo.NumberOfSupportedPTPInstances = numberOfSupportedPTPInstances
	}
	if ptpInstanceList != nil {
		newupNodeInfo.PTPInstanceList = ptpInstanceList
	}
	update_bridge_info := tsctsf_self.Bridges[upNodeID]
	update_bridge_info.UpNode_info = newupNodeInfo
	tsctsf_self.Bridges[upNodeID] = update_bridge_info
	return upNodeID
}

// TODO: other parameter
func DecodeUpNodeUpdatedStatus(umic models.BridgeManagementContainer, index int) uint64 {
	var ptpInstanceList []uint8
	var ptpID uint16
	var dsttInfoList []uint8
	var dsttPortNum uint32
	success_set_num := int(umic.BridgeManCont[index+3])
	i := index + 4
	for num := 0; num < success_set_num; num += 1 {
		parameter_name := uint16(umic.BridgeManCont[i])<<8 + uint16(umic.BridgeManCont[i+1])
		parameter_len := int(umic.BridgeManCont[i+2])
		switch parameter_name {
		case 123:
			i = i + 3
			dsttInfoList = umic.BridgeManCont[i : i+parameter_len]
			dsttPortNum = uint32(binary.BigEndian.Uint16(umic.BridgeManCont[i+2 : i+4]))
			logger.UtilLog.Traceln("DSTT Port Number", dsttPortNum, ",DSTT Port Time Sync Info list = ", dsttInfoList)
			i = i + parameter_len
		case 124: //ptp instance list
			i = i + 3
			ptpInstanceList = umic.BridgeManCont[i : i+parameter_len]
			ptpID = binary.BigEndian.Uint16(umic.BridgeManCont[i+2 : i+4])
			logger.UtilLog.Traceln("PTP instance list = ", ptpInstanceList)
			i = i + parameter_len

		default:
			logger.UtilLog.Infof("Skip, do not need [%x] it temporarily", parameter_name)
			i = i + 3 + parameter_len
		}
	}
	tsctsf_self := tsctsf_context.GetSelf()

	for bridge_id, Bridge_info := range tsctsf_self.Bridges {
		updateUpNode := Bridge_info.UpNode_info
		if updateUpNode.PTPInstanceId == ptpID {
			if ptpInstanceList != nil {
				updateUpNode.PTPInstanceList = ptpInstanceList
			}
			updateUpNode.Update = true
			Bridge_info.UpNode_info = updateUpNode
			tsctsf_self.Bridges[bridge_id] = Bridge_info
			return bridge_id
		}
	}
	return ^uint64(0)
}
