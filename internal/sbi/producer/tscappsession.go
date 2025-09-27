package producer

import (
	"net/http"

	"github.com/HanHongChen/openapi-tsctsf/models"
	"github.com/HanHongChen/tsctsf/internal/logger"
	"github.com/HanHongChen/tsctsf/internal/sbi/consumer"
	"github.com/free5gc/util/httpwrapper"
)

func HandleCreateTSCAppSessions(request *httpwrapper.Request) *httpwrapper.Response {
	logger.TSCAppSessLog.Infoln("Handle Create TSC App Sessions")

	tscAppSession := request.Body.(models.TscAppSessionContextData)

	response, locationHeader, problemDetails := TSCAppSessionsCreateProcedure(tscAppSession)
	header := http.Header{
		"Location": {locationHeader},
	}
	if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	} else {
		return httpwrapper.NewResponse(http.StatusCreated, header, response)
	}
}
func TSCAppSessionsCreateProcedure(tscAppSession models.TscAppSessionContextData) (
	models.TscAppSessionContextData,
	string,
	*models.ProblemDetails) {

	logger.TSCAppSessLog.Infof("tscAppSession")
	ueAddr := tscAppSession.UeIpAddr.Ipv4Addr
	afIdentifier := tscAppSession.AfId
	flowDescription := tscAppSession.FlowInfo
	qosParam := tscAppSession.QosReference

	logger.TSCAppSessLog.Infof("UeIpAddr: %v", ueAddr)
	logger.TSCAppSessLog.Infof("AfId: %v", afIdentifier)
	logger.TSCAppSessLog.Infof("FlowInfo: %v", flowDescription)
	logger.TSCAppSessLog.Infof("QosReference: %v", qosParam)

	// Calculate PDB

	// Time domain

	// send to pcf
	// tsctsf_self := tsctsf_context.GetSelf()
	// // appSessID, exist := tsctsf_self.AppSessionIdPool.Load(ueAddr)
	// _, exist := tsctsf_self.AppSessionIdPool.Load(ueAddr)
	// if !exist {
	// 	logger.TSCAppSessLog.Infof("No session found for the given ueAddr. Create new AF session.")
	var newDetNet consumer.PduSessionDetNet
	newDetNet.Dnn = tscAppSession.Dnn
	// newDetNet.Snssai = *tscAppSession.Snssai
	newDetNet.UeIpv4Addr = ueAddr
	// 	resp, err := consumer.HandleDetNetAppSessionCreate(newDetNet)
	// 	return
	// }
	logger.TSCAppSessLog.Infof("Create new AF session.")
	resp, err := consumer.HandleDetNetAppSessionCreate(newDetNet)
	tscAppSessionContextData := resp.TscAppSessionContextData
	if err != nil {
		logger.TSCAppSessLog.Warningf("Producer called TSCAppSessionsCreateProcedure error [%+v].", err)
		problemDetails := &models.ProblemDetails{
			Status: http.StatusInternalServerError,
			Cause:  "UNSPECIFIED",
		}
		return tscAppSessionContextData, "", problemDetails
	}
	logger.TSCAppSessLog.Warningln("TSCAppSessionsCreateProcedure successfully")
	return tscAppSessionContextData, resp.Location, nil
}
