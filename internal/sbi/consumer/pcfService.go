package consumer

import (
	"context"
	"errors"
	"sync"
	"time"

	tsctsf_models "github.com/HanHongChen/openapi-tsctsf/models"
	"github.com/HanHongChen/openapi-tsctsf/pcf/PolicyAuthorization"
	tsctsf_context "github.com/HanHongChen/tsctsf/internal/context"
	"github.com/HanHongChen/tsctsf/internal/logger"
	"github.com/HanHongChen/tsctsf/util"
)

type npcfService struct {
	consumer        *Consumer
	nfPolAuthMu     sync.RWMutex
	nfPolAuthClient map[string]*PolicyAuthorization.APIClient
}

func (s *npcfService) getPolicyAuthorizationClient(uri string) *PolicyAuthorization.APIClient {
	if uri == "" {
		return nil
	}
	s.nfPolAuthMu.RLock()
	client, ok := s.nfPolAuthClient[uri]
	if ok {
		s.nfPolAuthMu.RUnlock()
		return client
	}

	configuration := PolicyAuthorization.NewConfiguration()
	configuration.SetBasePath(uri)
	client = PolicyAuthorization.NewAPIClient(configuration)

	s.nfPolAuthMu.RUnlock()
	s.nfPolAuthMu.Lock()
	defer s.nfPolAuthMu.Unlock()
	s.nfPolAuthClient[uri] = client
	return client
}

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

func (s *npcfService) DetNetAppSessionCreate(uri string, new_detnet PduSessionDetNet) (resp PostTSCAppSessionsResponse, err error) {
	logger.PolicyAuthLog.Logger.Infoln("Handle DetNet App Session Context Creation")

	if uri == "" {
		return resp, errors.New("Npcf client uri is nil")
	}

	client := s.getPolicyAuthorizationClient(uri)
	if client == nil {
		return resp, errors.New("Npcf client is nil")
	}

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

	req := PolicyAuthorization.PostAppSessionsRequest{
		AppSessionContext: &PostAppSessionsData,
	}

	logger.PolicyAuthLog.Infoln("start to send Individual App Session context creation request")
	appSessContext, err := client.ApplicationSessionsCollectionApi.PostAppSessions(
		context.Background(), &req,
	)

	if err == nil {
		logger.PolicyAuthLog.Infof("send Individual App Session context create successfully")
		//TODO: move this part into processor
		// store the appSession ID with DNN/S-NSSAI
		var appSessID string

		resp.Location = appSessContext.Location
		appSessID = util.Split_appSessionId_str(resp.Location)
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

	return resp, err
}
