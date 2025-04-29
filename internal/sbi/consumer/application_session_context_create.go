package consumer

import (
	"context"
	"net/http"
	"time"

	// "github.com/HanHongChen/openapi-tsctsf/models"
	//github.com/HanHongChen/bitbucket-openapi/models
	tsctsf_models "github.com/HanHongChen/openapi-tsctsf/models"
	tsctsf_context "github.com/HanHongChen/tsctsf/internal/context"
	"github.com/HanHongChen/tsctsf/internal/logger"
	"github.com/HanHongChen/tsctsf/util"
)

type PduSessionDetNet struct {
	UeIpv4Addr  string
	Dnn         string
	Snssai      tsctsf_models.Snssai
	FlowDesc    tsctsf_models.FlowInfo
	Periodicity int32
}

type PostTSCAppSessionsResponse struct {
	Location                 string
	TscAppSessionContextData tsctsf_models.TscAppSessionContextData
}

func ConvertTscEventsToAfSubscriptions(tscEvents []tsctsf_models.TscEvent) []tsctsf_models.AfEventSubscription {
	var afSubs []tsctsf_models.AfEventSubscription

	for _, e := range tscEvents {
		var afEvent tsctsf_models.PcfPolicyAuthorizationAfEvent
		switch e {
		case tsctsf_models.TscEvent_FAILED_RESOURCES_ALLOCATION:
			afEvent = tsctsf_models.PcfPolicyAuthorizationAfEvent_FAILED_RESOURCES_ALLOCATION
		case tsctsf_models.TscEvent_SUCCESSFUL_RESOURCES_ALLOCATION:
			afEvent = tsctsf_models.PcfPolicyAuthorizationAfEvent_SUCCESSFUL_RESOURCES_ALLOCATION
		default:
			// Skip unsupported events
			logger.PolicyAuthLog.Warnf("Unsupported TSC event [%s] for PCF AFEventSubscription", e)
			continue
		}
		afSubs = append(afSubs, tsctsf_models.AfEventSubscription{
			Event: afEvent,
		})
	}
	return afSubs
}

func HandleDetNetAppSessionCreate(new_detnet PduSessionDetNet) (resp PostTSCAppSessionsResponse, err error) {
	logger.PolicyAuthLog.Logger.Infoln("Handle DetNet App Session Context Creation")

	tscEvent := []tsctsf_models.TscEvent{
		tsctsf_models.TscEvent_FAILED_RESOURCES_ALLOCATION,
		tsctsf_models.TscEvent_SUCCESSFUL_RESOURCES_ALLOCATION,
	}
	afSubs := ConvertTscEventsToAfSubscriptions(tscEvent)
	evSubsc := &tsctsf_models.PcfPolicyAuthorizationEventsSubscReqData{
		Events: afSubs,
	}

	ulTimeStr := "2022-07-25 15:14:50"
	layout := "2006-01-02 15:04:05"
	parsedUlTime, err := time.Parse(layout, ulTimeStr)
	dlTimeStr := "2022-07-25 15:18:50"
	parsedDlTime, err := time.Parse(layout, dlTimeStr)

	medComp := tsctsf_models.MediaComponent{
		AfAppId:  "vr",
		MedCompN: 1,
		TsnQos: &tsctsf_models.TsnQosContainer{
			TscPackDelay:    10,
			MaxTscBurstSize: 4096,
			TscPrioLevel:    1,
		},
		TscaiInputUl: &tsctsf_models.TscaiInputContainer{
			Periodicity:      1,
			BurstArrivalTime: &parsedUlTime,
		},
		TscaiInputDl: &tsctsf_models.TscaiInputContainer{
			Periodicity:      1,
			BurstArrivalTime: &parsedDlTime,
		},
	}
	// 4.2.2.25 Provisioning of TSC UMI and PMI
	PostAppSessionsData := tsctsf_models.AppSessionContext{
		AscReqData: &tsctsf_models.AppSessionContextReqData{
			SuppFeat: "28",
			UeIpv4:   new_detnet.UeIpv4Addr,
			EvSubsc:  evSubsc,
			NotifUri: "https://127.0.0.56:8000",
			// TsnPortManContDstt: new_bridge.TsnPortManContDstt,
			// TsnPortManContNwtt: new_bridge.TsnPortManContNwtts,
			MedComponents: map[string]tsctsf_models.MediaComponent{
				"1": medComp,
			},
		},
	}

	logger.PolicyAuthLog.Infoln("start to send Individual App Session context creation request")
	client := util.GetNpcfPolicyAuthorizationClient()
	appSessContext, httpResponse, err := client.ApplicationSessionsCollectionApi.PostAppSessions(
		context.Background(), PostAppSessionsData,
	)

	if err != nil {
		if httpResponse != nil {
			logger.PolicyAuthLog.Warnf("send Individual App Session context create Error[%s]", httpResponse.Status)
		} else {
			logger.PolicyAuthLog.Warnf("send Individual App Session context create Failed[%s]", err.Error())
		}
		return resp, err
	} else if httpResponse == nil {
		logger.PolicyAuthLog.Warnln("send Individual App Session context create Failed[httpResponse is nil]")
		return resp, nil
	}

	if httpResponse.StatusCode != http.StatusCreated && httpResponse.StatusCode != http.StatusNoContent {
		logger.PolicyAuthLog.Warnln("send Individual App Session context create Failed [%v]",
			httpResponse.StatusCode)
	} else {
		logger.PolicyAuthLog.Infof("send Individual App Session context create successfully")
	}

	// store the appSession ID with DNN/S-NSSAI
	var appSessID string
	logger.PolicyAuthLog.Errorf("AscReqData = [%+v], AscRespData = [%+v], EvsNotif = [%+v]",
		appSessContext.AscReqData, appSessContext.AscRespData, appSessContext.EvsNotif)

	resp.Location = httpResponse.Header.Get("Location")
	appSessID = util.Split_appSessionId_str(httpResponse.Header.Get("Location"))
	tsctsf_self := tsctsf_context.GetSelf()
	// dnnSnssai := new_detnet.Dnn + string(new_detnet.Snssai.Sst) + new_detnet.Snssai.Sd
	// TODO : the conditions to match for notifying the event within the "eventFilters" attribute;
	_, exist := tsctsf_self.AppSessionIdPool.Load(new_detnet.UeIpv4Addr)
	if !exist {
		logger.PolicyAuthLog.Infof("Store New AF-session ID :[%d] with DNN/S-NSSAI :[%s]", appSessID, new_detnet.UeIpv4Addr)
		tsctsf_self.AppSessionIdPool.Store(new_detnet.UeIpv4Addr, appSessID)

	}

	return resp, nil
}

// Npcf_PolicyAuthorization_create service 4.2.2.2
// func HandleAppSessionCreate(new_bridge tsctsf_models.PduSessionTsnBridge) {
// 	logger.PolicyAuthLog.Infoln("Handle App Session context Creation")

// 	TsnPortManContDstt := util.DSTTPMICDecodeCapabilityInfo(*new_bridge.TsnPortManContDstt)
// 	new_bridge.TsnPortManContDstt = &TsnPortManContDstt
// 	// for i := 0; i < len(new_bridge.TsnPortManContNwtts); i += 1 {
// 	// 	new_bridge.TsnPortManContNwtts[i].PortManCont = TsnPortManContDstt.PortManCont
// 	// }
// 	logger.PolicyAuthLog.Debugf("create PMIC for dstt to read : [%x]", new_bridge.TsnPortManContDstt)

// 	// 4.2.2.31 Subscription to TSN related events
// 	af := tsctsf_models.AfEventSubscription{
// 		Event: tsctsf_models.AfEvent_TSN_BRIDGE_INFO,
// 	}
// 	events := tsctsf_models.EventsSubscReqData{
// 		Events: []tsctsf_models.AfEventSubscription{
// 			af,
// 		},
// 	}

// 	// TODO: add survival time
// 	// 4.2.2.24 Provisioning of TSCAI input info and Qos related data
// 	medComp := models.MediaComponent{
// 		AfAppId:  "edge",
// 		MedCompN: 0,
// 		TscQos: &models.TsnQosContainer{
// 			TscPackDelay:    10,
// 			MaxTscBurstSize: 4096,
// 			TscPrioLevel:    1,
// 		},
// 		TscaiInputUl: &models.TscaiInputContainer{
// 			Periodicity:      1,
// 			BurstArrivalTime: "2022-07-25 15:14:50",
// 		},
// 		TscaiInputDl: &models.TscaiInputContainer{
// 			Periodicity:      1,
// 			BurstArrivalTime: "2022-07-25 15:18:50",
// 		},
// 	}

// 	// 4.2.2.25 Provisioning of TSC UMI and PMI
// 	PostAppSessionsData := models.AppSessionContext{
// 		AscReqData: &models.AppSessionContextReqData{
// 			SuppFeat:           "20",
// 			UeIpv4:             new_bridge.UeIpv4Addr,
// 			EvSubsc:            &events,
// 			NotifUri:           "https://127.0.0.7:8000",
// 			TsnPortManContDstt: new_bridge.TsnPortManContDstt,
// 			// TsnPortManContNwtt: new_bridge.TsnPortManContNwtts,
// 			MedComponents: map[string]models.MediaComponent{
// 				"Tsn": medComp,
// 			},
// 		},
// 	}
// 	// send Post request to PCF
// 	logger.PolicyAuthLog.Infoln("start to send Individual App Session context creation request")
// 	client := util.GetNpcfPolicyAuthorizationClient()
// 	_, httpResponse, err := client.ApplicationSessionsCollectionApi.PostAppSessions(
// 		context.Background(), PostAppSessionsData,
// 	)
// 	if err != nil {
// 		if httpResponse != nil {
// 			logger.PolicyAuthLog.Warnf("send Individual App Session context create Error[%s]", httpResponse.Status)
// 		} else {
// 			logger.PolicyAuthLog.Warnf("send Individual App Session context create Failed[%s]", err.Error())
// 		}

// 	} else if httpResponse == nil {
// 		logger.PolicyAuthLog.Warnln("send Individual App Session context create Failed[httpResponse is nil]")
// 	}
// 	defer func() {
// 		if rspCloseErr := httpResponse.Body.Close(); rspCloseErr != nil {
// 			logger.PolicyAuthLog.Errorf(
// 				"PolicyAuthorizationPostAppSessions response body cannot close: %+v",
// 				rspCloseErr)
// 		}
// 	}()
// 	if httpResponse.StatusCode != http.StatusCreated && httpResponse.StatusCode != http.StatusNoContent {
// 		logger.PolicyAuthLog.Warnln("send Individual App Session context create Failed [%v]",
// 			httpResponse.StatusCode)
// 	} else {
// 		logger.PolicyAuthLog.Infof("send Individual App Session context create successfully")
// 	}

// 	// store the appSession ID with DNN/S-NSSAI
// 	var Loc *url.URL
// 	var appSessID string
// 	Loc, _ = httpResponse.Location()
// 	appSessID = util.Split_appSessionId(Loc)
// 	tsctsf_self := tsctsf_context.GetSelf()
// 	dnnSnssai := new_bridge.Dnn + string(new_bridge.Snssai.Sst) + new_bridge.Snssai.Sd
// 	// TODO : the conditions to match for notifying the event within the "eventFilters" attribute;
// 	_, exist := tsctsf_self.AppSessionIdPool.Load(dnnSnssai)
// 	if !exist {
// 		logger.PolicyAuthLog.Infof("Store New AF-session ID :[%d] with DNN/S-NSSAI :[%s]", appSessID, dnnSnssai)
// 		tsctsf_self.AppSessionIdPool.Store(dnnSnssai, appSessID)

// 	}

// 	//test bridge management api
// 	//configuration := test_api(new_bridge.TsnPortManContDstt)
// 	//logger.BridgeInfoManagementlog.Infof("receive TSN Bridge Configuration from CNC : %#v", configuration)
// }
