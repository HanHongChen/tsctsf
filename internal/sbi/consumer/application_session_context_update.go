package consumer

import (
	"context"
	"net/http"

	"github.com/HanHongChen/bitbucket-openapi/models"
	"github.com/HanHongChen/tsctsf/internal/logger"
	"github.com/HanHongChen/tsctsf/util"
)

func AppSessionUpdate_PMIC(portManCont models.PortManagementContainer,
	appSessionId string, dstt_or_nwtt bool) {
	var appSessionContextUpdateData models.AppSessionContextUpdateData

	// 4.2.2.31 Subscription to TSN related events
	af := models.AfEventSubscription{
		Event: models.AfEvent_TSN_BRIDGE_INFO,
	}
	events := models.EventsSubscReqDataRm{
		Events: []models.AfEventSubscription{
			af,
		},
	}

	if dstt_or_nwtt {
		logger.ConsumerLog.Info("send nwtt pmic")
		appSessionContextUpdateData = models.AppSessionContextUpdateData{
			EvSubsc: &events,
			TsnPortManContNwtts: []models.PortManagementContainer{
				portManCont,
			},
		}
	} else {
		logger.ConsumerLog.Info("send dstt pmic")
		appSessionContextUpdateData = models.AppSessionContextUpdateData{
			EvSubsc:            &events,
			TsnPortManContDstt: &portManCont,
		}
	}
	HandleAppSessionUpdate(appSessionId, appSessionContextUpdateData)
}

func AppSessionUpdate_UMIC(bridgeManCont models.BridgeManagementContainer,
	appSessionId string) {
	var appSessionContextUpdateData models.AppSessionContextUpdateData

	// 4.2.2.31 Subscription to TSN related events
	af := models.AfEventSubscription{
		Event: models.AfEvent_TSN_BRIDGE_INFO,
	}
	events := models.EventsSubscReqDataRm{
		Events: []models.AfEventSubscription{
			af,
		},
	}
	logger.ConsumerLog.Info("send nwtt umic")

	appSessionContextUpdateData = models.AppSessionContextUpdateData{
		EvSubsc:          &events,
		TsnBridgeManCont: &bridgeManCont,
	}

	HandleAppSessionUpdate(appSessionId, appSessionContextUpdateData)
}

func HandleAppSessionUpdate(appSessionId string,
	appSessionContextUpdateData models.AppSessionContextUpdateData) {
	logger.PolicyAuthLog.Infoln("Handle App Session context Update")

	// send Post request to PCF
	logger.PolicyAuthLog.Infoln("start to send Individual App Session context Update request")
	client := util.GetNpcfPolicyAuthorizationClient()
	_, httpResponse, err := client.IndividualApplicationSessionContextDocumentApi.ModAppSession(
		context.Background(), appSessionId, appSessionContextUpdateData)
	if err != nil {
		if httpResponse != nil {
			logger.PolicyAuthLog.Warnf("send Individual App Session context Update Error[%s]", httpResponse.Status)
		} else {
			logger.PolicyAuthLog.Warnf("send Individual App Session context Update Failed[%s]", err.Error())
		}

	} else if httpResponse == nil {
		logger.PolicyAuthLog.Warnln("send Individual App Session context Update Failed[httpResponse is nil]")
	}
	defer func() {
		if rspCloseErr := httpResponse.Body.Close(); rspCloseErr != nil {
			logger.PolicyAuthLog.Errorf(
				"PolicyAuthorizationModAppSession response body cannot close: %+v",
				rspCloseErr)
		}
	}()
	if httpResponse.StatusCode != http.StatusOK && httpResponse.StatusCode != http.StatusNoContent {
		logger.PolicyAuthLog.Warnln("send Individual App Session context Update Failed ", httpResponse.StatusCode)
	} else {
		logger.PolicyAuthLog.Infof("send Individual App Session context Update successfully")
	}

}
