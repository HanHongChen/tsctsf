package processor

import (
	"net/http"
	"time"

	"github.com/HanHongChen/openapi-tsctsf/models"
	"github.com/HanHongChen/openapi-tsctsf/pcf/PolicyAuthorization"
	tsctsf_context "github.com/HanHongChen/tsctsf/internal/context"
	"github.com/HanHongChen/tsctsf/internal/logger"
	"github.com/HanHongChen/tsctsf/util"
	"github.com/gin-gonic/gin"
)

func ConvertTscEventsToAfSubscriptions(tscEvents []models.TscEvent) []models.AfEventSubscription {
	var afSubs []models.AfEventSubscription

	for _, e := range tscEvents {
		var afEvent models.PcfPolicyAuthorizationAfEvent
		switch e {
		case models.TscEvent_FAILED_RESOURCES_ALLOCATION:
			afEvent = models.PcfPolicyAuthorizationAfEvent_FAILED_RESOURCES_ALLOCATION
		case models.TscEvent_SUCCESSFUL_RESOURCES_ALLOCATION:
			afEvent = models.PcfPolicyAuthorizationAfEvent_SUCCESSFUL_RESOURCES_ALLOCATION
		default:
			// Skip unsupported events
			logger.PolicyAuthLog.Warnf("Unsupported TSC event [%s] for PCF AFEventSubscription", e)
			continue
		}
		afSubs = append(afSubs, models.AfEventSubscription{
			Event: afEvent,
		})
	}
	return afSubs
}

// func (p *Processor) getDefaultPcfUri(context *tsctsf_context.TSCTSFContext) string {
// 	context
// }

func (p *Processor) HandlePostTSCAppSession(c *gin.Context, tscAppSessContext models.TscAppSessionContextData) {
	logger.TSCAppSessLog.Info("Handle TSC App Session Post")

	if tscAppSessContext.UeIpAddr.Ipv4Addr == "" {
		logger.TSCAppSessLog.Warning("PostTSCAppSession Ip addr is nil")
		return
	} else if tscAppSessContext.AfId == "" {
		logger.TSCAppSessLog.Warning("PostTSCAppSession AfId is nil")
		return
	} else if tscAppSessContext.FlowInfo == nil {
		logger.TSCAppSessLog.Warning("PostTSCAppSession FlowInfo is nil")
		return
	} else if tscAppSessContext.QosReference == "" {
		logger.TSCAppSessLog.Warning("PostTSCAppSession QosReference is nil")
		return
	}

	logger.TSCAppSessLog.Infof("UeIpAddr: %v", tscAppSessContext.UeIpAddr.Ipv4Addr)
	logger.TSCAppSessLog.Infof("AfId: %v", tscAppSessContext.AfId)
	logger.TSCAppSessLog.Infof("FlowInfo: %v", tscAppSessContext.FlowInfo)
	logger.TSCAppSessLog.Infof("QosReference: %v", tscAppSessContext.QosReference)

	tscEvent := []models.TscEvent{
		models.TscEvent_FAILED_RESOURCES_ALLOCATION,
		models.TscEvent_SUCCESSFUL_RESOURCES_ALLOCATION,
	}

	afSubs := ConvertTscEventsToAfSubscriptions(tscEvent)
	evSubsc := &models.PcfPolicyAuthorizationEventsSubscReqData{
		Events: afSubs,
	}

	ulTimeStr := "2022-07-25 15:14:50"
	layout := "2006-01-02 15:04:05"
	parsedUlTime, err := time.Parse(layout, ulTimeStr)
	if err != nil {
		logger.TSCAppSessLog.Warnf("PostTSCAppSession parseUlTime fail [%v]", err)
	}
	dlTimeStr := "2022-07-25 15:18:50"
	parsedDlTime, err := time.Parse(layout, dlTimeStr)
	if err != nil {
		logger.TSCAppSessLog.Warnf("PostTSCAppSession parseDlTime fail [%v]", err)
	}

	medComp := models.MediaComponent{
		AfAppId:  tscAppSessContext.AfId,
		MedCompN: 1,
		TsnQos: &models.TsnQosContainer{
			TscPackDelay:    10,
			MaxTscBurstSize: 4096,
			TscPrioLevel:    1,
		},
		TscaiInputUl: &models.TscaiInputContainer{
			Periodicity:      1,
			BurstArrivalTime: &parsedUlTime,
		},
		TscaiInputDl: &models.TscaiInputContainer{
			Periodicity:      1,
			BurstArrivalTime: &parsedDlTime,
		},
	}

	appSessContext := &models.AppSessionContext{
		AscReqData: &models.AppSessionContextReqData{
			SuppFeat: "28",
			UeIpv4:   tscAppSessContext.UeIpAddr.Ipv4Addr,
			EvSubsc:  evSubsc,
			NotifUri: "https://127.0.0.56:8000",
			// TsnPortManContDstt: new_bridge.TsnPortManContDstt,
			// TsnPortManContNwtt: new_bridge.TsnPortManContNwtts,
			MedComponents: map[string]models.MediaComponent{
				"1": medComp,
			},
		},
	}

	req := &PolicyAuthorization.PostAppSessionsRequest{
		AppSessionContext: appSessContext,
	}

	//TODO: change get PCF uri way
	resp, err := p.Consumer().DetNetAppSessionCreate("http://127.0.0.7:8000", req)

	if err == nil {
		// store the appSession ID with DNN/S-NSSAI
		var appSessID string

		appSessID = util.Split_appSessionId_str(resp.Location)
		tsctsf_self := tsctsf_context.GetSelf()
		// dnnSnssai := new_detnet.Dnn + string(new_detnet.Snssai.Sst) + new_detnet.Snssai.Sd
		// TODO : the conditions to match for notifying the event within the "eventFilters" attribute;
		_, exist := tsctsf_self.AppSessionIdPool.Load(resp.AppSessionContext.AscReqData.UeIpv4)
		if !exist {
			logger.TSCAppSessLog.Infof("Store New AF-session ID :[%d] with DNN/S-NSSAI :[%s]", appSessID,
				resp.AppSessionContext.AscReqData.UeIpv4)
			tsctsf_self.AppSessionIdPool.Store(resp.AppSessionContext.AscReqData.UeIpv4, appSessID)
		}

		tsctsfResp := models.TscAppSessionContextData{
			AfId:         tscAppSessContext.AfId,
			UeIpAddr:     tscAppSessContext.UeIpAddr,
			FlowInfo:     tscAppSessContext.FlowInfo,
			QosReference: tscAppSessContext.QosReference,
			AppId:        tscAppSessContext.AppId, // parsed from PCF Location
			//TODO: SPEC said it need notifyUri
		}

		c.JSON(http.StatusCreated, tsctsfResp)
	}

	logger.TSCAppSessLog.Warnf("PostTSCAppSession error [%+v]", err.Error())
	c.JSON(http.StatusInternalServerError, err)
}
