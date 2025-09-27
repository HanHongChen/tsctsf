package producer

import (
	"fmt"
	"math"
	"strconv"

	//github.com/HanHongChen/bitbucket-openapi/models
	tsctsf_context "github.com/HanHongChen/tsctsf/internal/context"
	"github.com/HanHongChen/tsctsf/internal/logger"
	"github.com/HanHongChen/tsctsf/pkg/factory"
)

const (
	DSTT bool = false
	NWTT bool = true
)

// Creates a new configuration resource to activate time synchronization service.
// func HandleCreateIndividualTimeSynchronizationExposureConfiguration(request *httpwrapper.Request) *httpwrapper.Response {
// 	logger.TimeSyncCfgLog.Infoln("Handle Create Individual Time Synchronization Exposure Configuration")

// 	newTimeSyncExpoxeCfg := request.Body.(models.TimeSyncExposureConfig)
// 	subscriptionID := request.Params["subscriptionId"]
// 	url, problemDetails := TimeSyncExpoCfgCreateProcedure(newTimeSyncExpoxeCfg, subscriptionID)
// 	header := http.Header{
// 		"Location": {url},
// 	}
// 	if problemDetails != nil {
// 		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
// 	} else {
// 		return httpwrapper.NewResponse(http.StatusCreated, header, newTimeSyncExpoxeCfg)
// 	}
// }

// func TimeSyncExpoCfgCreateProcedure(timeSyncExpoCfg models.TimeSyncExposureConfig, subscriptionID string) (string, *models.ProblemDetails) {
// 	logger.TimeSyncCfgLog.Infof("PTP instance type : %v", timeSyncExpoCfg.ReqPtpIns.InstanceType)
// 	logger.TimeSyncCfgLog.Infof("Transport protocol : %v", timeSyncExpoCfg.ReqPtpIns.Protocol)
// 	logger.TimeSyncCfgLog.Infof("PTP Profile : %v", timeSyncExpoCfg.ReqPtpIns.PtpProfile)
// 	logger.TimeSyncCfgLog.Infof("Grandmaster enabled : %v", timeSyncExpoCfg.GmEnable)
// 	// logger.TimeSyncCfgLog.Infof("Grandmaster priority :")
// 	logger.TimeSyncCfgLog.Infof("Time Domain : %v", timeSyncExpoCfg.TimeDom)
// 	logger.TimeSyncCfgLog.Infof("UE identity (for a DS-TT port) : %v", timeSyncExpoCfg.ReqPtpIns.PortConfigs[0].Supi)

// 	for i, subscription := range factory.TsctsfConfig.Subscriptions {
// 		if subscription.SubscriptionId == subscriptionID {
// 			// create a new configuration
// 			newConfigID, err := getUnusedConfigID()
// 			if err != nil {
// 				logger.TimeSyncCfgLog.Warnf(err.Error())

// 				problemDetails := models.ProblemDetails{
// 					Title:  "Unsupported request resources",
// 					Status: http.StatusNotFound,
// 					Detail: err.Error(),
// 				}
// 				return "", &problemDetails
// 			}
// 			if subscription.ConfigurationId == "" && subscription.SubscriptionCfg == nil {
// 				factory.TsctsfConfig.Subscriptions[i].ConfigurationId = newConfigID
// 				factory.TsctsfConfig.Subscriptions[i].SubscriptionCfg = &timeSyncExpoCfg
// 			}
// 			appsessID, exist := getAppSessIDBySubscID(subscriptionID)
// 			if !exist {
// 				//  TODO : problem msg
// 				problemDetails := &models.ProblemDetails{
// 					Status: http.StatusBadRequest,
// 					Cause:  "No App Session exist",
// 				}
// 				return "", problemDetails
// 			}
// 			ptpInstanceID := AssigedPTPInstanceID(timeSyncExpoCfg.UpNodeId)
// 			umic := util.CreatePTPInstanceListForUMIC(timeSyncExpoCfg, ptpInstanceID)
// 			// pmic := util.CreatePTPInstanceListForPMIC(timeSyncExpoCfg, ptpInstanceID, DSTT)

// 			// consumer.AppSessionUpdate_PMIC(pmic, appsessID, DSTT) // DSTT
// 			consumer.AppSessionUpdate_UMIC(umic, appsessID)

// 			// logger.TimeSyncCfgLog.Debugf("ConfigID[%s] with SubscriptionCfg: %+v", factory.TsctsfConfig.Subscriptions[i].ConfigurationId, factory.TsctsfConfig.Subscriptions[i].SubscriptionCfg)
// 			resourceUri := fmt.Sprintf("ntsctsf-time-sync/v1/subscriptions/%s/configuration/%s", subscriptionID, newConfigID)
// 			return resourceUri, nil
// 		}
// 	}
// 	logger.TimeSyncCfgLog.Warnf("Update Subscription data of Subscription ID[%s] fail.", subscriptionID)

// 	//  TODO : problem msg
// 	problemDetails := &models.ProblemDetails{
// 		Status: http.StatusBadRequest,
// 		Cause:  "Malformed request syntax",
// 	}
// 	return "", problemDetails

// }

func getUnusedConfigID() (string, error) {
	var idx uint32 = 1
	factory.TsctsfConfig.RLock()
	defer factory.TsctsfConfig.RUnlock()

	usedConfigIDs := make(map[uint32]bool)

	// collect exist ConfigID
	for _, subscription := range factory.TsctsfConfig.Subscriptions {
		if subscription.ConfigurationId != "" {
			tempID, err := strconv.Atoi(subscription.ConfigurationId)
			if err != nil {
				return "", err
			}
			usedConfigIDs[uint32(tempID)] = true
		}
	}

	for idx = 1; idx <= math.MaxUint32; idx++ {
		if !usedConfigIDs[idx] {
			return strconv.Itoa(int(idx)), nil
		}
	}
	return "", fmt.Errorf("No available Config ID")
}

func getAppSessIDBySubscID(subscID string) (string, bool) {
	tsctsf_self := tsctsf_context.GetSelf()

	for appSessID, id := range tsctsf_self.SubscripSession {
		if id == subscID {
			return appSessID, true
		}
	}
	return "", false
}

/*
Assigning a new PTP Instance ID and indicating it to the NW-TT in "PTP instance specification" in UMIC and PMIC(s) for each NW
TT port that is part of the PTP instance.
*/
func AssigedPTPInstanceID(upNodeId uint64) uint16 {
	tsctsf_self := tsctsf_context.GetSelf()
	// [TODO]: gen PTP instance ID
	ptpID := uint16(12345)

	updateBridge := tsctsf_self.Bridges[upNodeId]

	for port := range updateBridge.Nwtt_ports {
		port_info := updateBridge.Nwtt_ports[port]
		port_info.PTPInstanceId = ptpID
		tsctsf_self.Bridges[upNodeId].Nwtt_ports[port] = port_info
	}
	for port := range updateBridge.Dstt_ports {
		port_info := updateBridge.Dstt_ports[port]
		port_info.PTPInstanceId = ptpID
		tsctsf_self.Bridges[upNodeId].Dstt_ports[port] = port_info
	}
	updateBridge.UpNode_info.PTPInstanceId = ptpID
	tsctsf_self.Bridges[upNodeId] = updateBridge
	logger.TimeSyncCfgLog.Infoln("Assigning a new PTP Instance ID :", ptpID)
	return ptpID
}
