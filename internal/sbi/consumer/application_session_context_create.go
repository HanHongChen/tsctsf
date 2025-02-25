package consumer

import (
	"context"
	"net/http"
	"net/url"

	"github.com/HanHongChen/bitbucket-openapi/models"
	tsctsf_context "github.com/yashihh/tsctsf/internal/context"
	"github.com/yashihh/tsctsf/internal/logger"
	"github.com/yashihh/tsctsf/util"
)

// Npcf_PolicyAuthorization_create service 4.2.2.2
func HandleAppSessionCreate(new_bridge models.PduSessionTsnBridge) {
	logger.PolicyAuthLog.Infoln("Handle App Session context Creation")

	TsnPortManContDstt := util.DSTTPMICDecodeCapabilityInfo(*new_bridge.TsnPortManContDstt)
	new_bridge.TsnPortManContDstt = &TsnPortManContDstt
	// for i := 0; i < len(new_bridge.TsnPortManContNwtts); i += 1 {
	// 	new_bridge.TsnPortManContNwtts[i].PortManCont = TsnPortManContDstt.PortManCont
	// }
	logger.PolicyAuthLog.Debugf("create PMIC for dstt to read : [%x]", new_bridge.TsnPortManContDstt)

	// 4.2.2.31 Subscription to TSN related events
	af := models.AfEventSubscription{
		Event: models.AfEvent_TSN_BRIDGE_INFO,
	}
	events := models.EventsSubscReqData{
		Events: []models.AfEventSubscription{
			af,
		},
	}

	// TODO: add survival time
	// 4.2.2.24 Provisioning of TSCAI input info and Qos related data
	medComp := models.MediaComponent{
		AfAppId:  "edge",
		MedCompN: 0,
		TscQos: &models.TsnQosContainer{
			TscPackDelay:    10,
			MaxTscBurstSize: 4096,
			TscPrioLevel:    1,
		},
		TscaiInputUl: &models.TscaiInputContainer{
			Periodicity:      1,
			BurstArrivalTime: "2022-07-25 15:14:50",
		},
		TscaiInputDl: &models.TscaiInputContainer{
			Periodicity:      1,
			BurstArrivalTime: "2022-07-25 15:18:50",
		},
	}

	// 4.2.2.25 Provisioning of TSC UMI and PMI
	PostAppSessionsData := models.AppSessionContext{
		AscReqData: &models.AppSessionContextReqData{
			SuppFeat:           "20",
			UeIpv4:             new_bridge.UeIpv4Addr,
			EvSubsc:            &events,
			NotifUri:           "https://127.0.0.7:8000",
			TsnPortManContDstt: new_bridge.TsnPortManContDstt,
			// TsnPortManContNwtt: new_bridge.TsnPortManContNwtts,
			MedComponents: map[string]models.MediaComponent{
				"Tsn": medComp,
			},
		},
	}
	// send Post request to PCF
	logger.PolicyAuthLog.Infoln("start to send Individual App Session context creation request")
	client := util.GetNpcfPolicyAuthorizationClient()
	_, httpResponse, err := client.ApplicationSessionsCollectionApi.PostAppSessions(
		context.Background(), PostAppSessionsData,
	)
	if err != nil {
		if httpResponse != nil {
			logger.PolicyAuthLog.Warnf("send Individual App Session context create Error[%s]", httpResponse.Status)
		} else {
			logger.PolicyAuthLog.Warnf("send Individual App Session context create Failed[%s]", err.Error())
		}

	} else if httpResponse == nil {
		logger.PolicyAuthLog.Warnln("send Individual App Session context create Failed[httpResponse is nil]")
	}
	defer func() {
		if rspCloseErr := httpResponse.Body.Close(); rspCloseErr != nil {
			logger.PolicyAuthLog.Errorf(
				"PolicyAuthorizationPostAppSessions response body cannot close: %+v",
				rspCloseErr)
		}
	}()
	if httpResponse.StatusCode != http.StatusCreated && httpResponse.StatusCode != http.StatusNoContent {
		logger.PolicyAuthLog.Warnln("send Individual App Session context create Failed [%v]",
			httpResponse.StatusCode)
	} else {
		logger.PolicyAuthLog.Infof("send Individual App Session context create successfully")
	}

	// store the appSession ID with DNN/S-NSSAI
	var Loc *url.URL
	var appSessID string
	Loc, _ = httpResponse.Location()
	appSessID = util.Split_appSessionId(Loc)
	tsctsf_self := tsctsf_context.GetSelf()
	dnnSnssai := new_bridge.Dnn + string(new_bridge.Snssai.Sst) + new_bridge.Snssai.Sd
	// TODO : the conditions to match for notifying the event within the "eventFilters" attribute;
	_, exist := tsctsf_self.AppSessionIdPool.Load(dnnSnssai)
	if !exist {
		logger.PolicyAuthLog.Infof("Store New AF-session ID :[%d] with DNN/S-NSSAI :[%s]", appSessID, dnnSnssai)
		tsctsf_self.AppSessionIdPool.Store(dnnSnssai, appSessID)

	}

	//test bridge management api
	//configuration := test_api(new_bridge.TsnPortManContDstt)
	//logger.BridgeInfoManagementlog.Infof("receive TSN Bridge Configuration from CNC : %#v", configuration)
}
