package consumer

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/HanHongChen/bitbucket-openapi/models"
	tsctsf_models "github.com/HanHongChen/openapi-tsctsf/models"
	tsctsf_context "github.com/HanHongChen/tsctsf/internal/context"
	"github.com/HanHongChen/tsctsf/internal/logger"
	"github.com/HanHongChen/tsctsf/util"
	"github.com/free5gc/openapi"
)

type PduSessionDetNet struct {
	UeIpv4Addr  *tsctsf_models.IpAddr
	Dnn         string
	Snssai      tsctsf_models.Snssai
	FlowDesc    tsctsf_models.FlowInfo
	Periodicity int32
}

type PostTSCAppSessionsResponse struct {
	Location                 string
	TscAppSessionContextData *tsctsf_models.TscAppSessionContextData
}

func HandleDetNetAppSessionCreate(new_detnet PduSessionDetNet) (*PostTSCAppSessionsResponse, error) {
	logger.PolicyAuthLog.Logger.Infoln("Handle DetNet App Session Context Creation")

	tscEvent := tsctsf_models.EventsSubscReqData{
		Events: []tsctsf_models.TscEvent{
			tsctsf_models.TscEvent_FAILED_RESOURCES_ALLOCATION,
			tsctsf_models.TscEvent_SUCCESSFUL_RESOURCES_ALLOCATION,
		},
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
			EvSubsc:  &tscEvent,
			NotifUri: "https://127.0.0.7:8000",
			// TsnPortManContDstt: new_bridge.TsnPortManContDstt,
			// TsnPortManContNwtt: new_bridge.TsnPortManContNwtts,
			MedComponents: map[string]tsctsf_models.MediaComponent{
				"1": medComp,
			},
		},
	}

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
		return nil, err
	} else if httpResponse == nil {
		logger.PolicyAuthLog.Warnln("send Individual App Session context create Failed[httpResponse is nil]")
		return nil, nil
	}
	defer func() {
		if rspCloseErr := httpResponse.Body.Close(); rspCloseErr != nil {
			logger.PolicyAuthLog.Errorf(
				"PolicyAuthorizationPostAppSessions response body cannot close: %+v",
				rspCloseErr)
		}
	}()

	bodyBytes, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		logger.PolicyAuthLog.Errorf("PolicyAuthorizationPostAppSessions response body cannot read: %+v",
			err,
		)
		return nil, err
	}
	if httpResponse.StatusCode != http.StatusCreated && httpResponse.StatusCode != http.StatusNoContent {
		logger.PolicyAuthLog.Warnln("send Individual App Session context create Failed [%v]",
			httpResponse.StatusCode)
	} else {
		logger.PolicyAuthLog.Infof("send Individual App Session context create successfully")
	}

	// store the appSession ID with DNN/S-NSSAI
	var Loc *url.URL
	var appSessID string
	var resp PostTSCAppSessionsResponse
	if err := openapi.Deserialize(&resp, bodyBytes, httpResponse.Header.Get("Content-Type")); err != nil {
		logger.PolicyAuthLog.Warnf("Failed to deserialize response: %s", err.Error())
		return nil, err
	}

	resp.Location = httpResponse.Header.Get("Location")
	appSessID = util.Split_appSessionId(Loc)
	tsctsf_self := tsctsf_context.GetSelf()
	// dnnSnssai := new_detnet.Dnn + string(new_detnet.Snssai.Sst) + new_detnet.Snssai.Sd
	// TODO : the conditions to match for notifying the event within the "eventFilters" attribute;
	_, exist := tsctsf_self.AppSessionIdPool.Load(new_detnet.UeIpv4Addr)
	if !exist {
		logger.PolicyAuthLog.Infof("Store New AF-session ID :[%d] with DNN/S-NSSAI :[%s]", appSessID, new_detnet.UeIpv4Addr)
		tsctsf_self.AppSessionIdPool.Store(new_detnet.UeIpv4Addr, appSessID)

	}
	return &resp, nil
}

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
