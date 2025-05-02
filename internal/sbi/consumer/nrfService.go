package consumer

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/HanHongChen/openapi-tsctsf/models"
	"github.com/HanHongChen/openapi-tsctsf/nrf/NFDiscovery"
	"github.com/HanHongChen/openapi-tsctsf/nrf/NFManagement"
	tsctsf_context "github.com/HanHongChen/tsctsf/internal/context"
	"github.com/HanHongChen/tsctsf/internal/logger"
	"github.com/HanHongChen/tsctsf/internal/util"
)

type nnrfService struct {
	consumer *Consumer

	nfMngmntMu sync.RWMutex
	nfDiscMu   sync.RWMutex

	nfMngmntClients map[string]*NFManagement.APIClient
	nfDiscClients   map[string]*NFDiscovery.APIClient
}

func (s *nnrfService) getNFManagementClient(uri string) *NFManagement.APIClient {
	if uri == "" {
		return nil
	}
	s.nfMngmntMu.RLock()
	client, ok := s.nfMngmntClients[uri]
	if ok {
		defer s.nfMngmntMu.RUnlock()
		return client
	}

	configuration := NFManagement.NewConfiguration()
	configuration.SetBasePath(uri)
	client = NFManagement.NewAPIClient(configuration)

	s.nfMngmntMu.RUnlock()
	s.nfMngmntMu.Lock()
	defer s.nfMngmntMu.Unlock()
	s.nfMngmntClients[uri] = client
	return client
}

func (s *nnrfService) getNFDiscClient(uri string) *NFDiscovery.APIClient {
	if uri == "" {
		return nil
	}
	s.nfDiscMu.RLock()
	client, ok := s.nfDiscClients[uri]
	if ok {
		defer s.nfDiscMu.RUnlock()
		return client
	}

	configuration := NFDiscovery.NewConfiguration()
	configuration.SetBasePath(uri)
	client = NFDiscovery.NewAPIClient(configuration)

	s.nfDiscMu.RUnlock()
	s.nfDiscMu.Lock()
	defer s.nfDiscMu.Unlock()
	s.nfDiscClients[uri] = client
	return client
}

func (s *nnrfService) SendSearchNFInstances(
	nrfUri string, targetNfType, requestNfType models.NrfNfManagementNfType, param NFDiscovery.SearchNFInstancesRequest) (
	*models.SearchResult, error,
) {
	// Set client and set url
	client := s.getNFDiscClient(nrfUri)

	ctx, _, err := s.consumer.Context().GetTokenCtx(models.ServiceName_NNRF_DISC, models.NrfNfManagementNfType_NRF)
	if err != nil {
		return nil, err
	}
	param.TargetNfType = &targetNfType
	param.RequesterNfType = &requestNfType
	res, err := client.NFInstancesStoreApi.SearchNFInstances(ctx, &param)
	if err != nil {
		logger.ConsumerLog.Errorf("SearchNFInstances failed: %+v", err)
		return nil, err
	}

	result := res.SearchResult

	return &result, nil
}

func (s *nnrfService) SendNFInstancesUDR(nrfUri, id string) string {
	targetNfType := models.NrfNfManagementNfType_UDR
	requestNfType := models.NrfNfManagementNfType_TSCTSF
	localVarOptionals := NFDiscovery.SearchNFInstancesRequest{
		// 	DataSet: optional.NewInterface(models.DataSetId_SUBSCRIPTION),
	}

	result, err := s.SendSearchNFInstances(nrfUri, targetNfType, requestNfType, localVarOptionals)
	if err != nil {
		logger.ConsumerLog.Error(err.Error())
		return ""
	}
	for _, profile := range result.NfInstances {
		if uri := util.SearchNFServiceUri(profile, models.ServiceName_NUDR_DR, models.NfServiceStatus_REGISTERED); uri != "" {
			return uri
		}
	}
	return ""
}

func (s *nnrfService) SendNFInstancesBSF(nrfUri string) string {
	targetNfType := models.NrfNfManagementNfType_BSF
	requestNfType := models.NrfNfManagementNfType_TSCTSF
	localVarOptionals := NFDiscovery.SearchNFInstancesRequest{}

	result, err := s.SendSearchNFInstances(nrfUri, targetNfType, requestNfType, localVarOptionals)
	if err != nil {
		logger.ConsumerLog.Error(err.Error())
		return ""
	}
	for _, profile := range result.NfInstances {
		if uri := util.SearchNFServiceUri(profile, models.ServiceName_NBSF_MANAGEMENT,
			models.NfServiceStatus_REGISTERED); uri != "" {
			return uri
		}
	}
	return ""
}

func (s *nnrfService) SendNFInstancesAMF(nrfUri string, guami models.Guami, serviceName models.ServiceName) string {
	targetNfType := models.NrfNfManagementNfType_AMF
	requestNfType := models.NrfNfManagementNfType_TSCTSF

	localVarOptionals := NFDiscovery.SearchNFInstancesRequest{
		Guami: &guami,
	}

	result, err := s.SendSearchNFInstances(nrfUri, targetNfType, requestNfType, localVarOptionals)
	if err != nil {
		logger.ConsumerLog.Error(err.Error())
		return ""
	}
	for _, profile := range result.NfInstances {
		return util.SearchNFServiceUri(profile, serviceName, models.NfServiceStatus_REGISTERED)
	}
	return ""
}

// management
func (s *nnrfService) BuildNFInstance(
	context *tsctsf_context.TSCTSFContext,
) (profile models.NrfNfManagementNfProfile, err error) {
	profile.NfInstanceId = context.NfId
	profile.NfType = models.NrfNfManagementNfType_TSCTSF
	profile.NfStatus = models.NrfNfManagementNfStatus_REGISTERED
	profile.Ipv4Addresses = append(profile.Ipv4Addresses, context.RegisterIPv4)
	services := []models.NrfNfManagementNfService{}
	for _, nfService := range context.NfService {
		services = append(services, nfService)
	}
	if len(services) > 0 {
		profile.NfServices = services
	}

	profile.TsctsfInfoList = map[string]models.TsctsfInfo{
		"tsctsf-1": {
			SNssaiInfoList: map[string]models.SnssaiTsctsfInfoItem{
				`{"sst":1,"sd":"010203"}`: {
					SNssai: &models.ExtSnssai{
						Sst: 1,
						Sd:  "010203",
					},
					DnnInfoList: []models.DnnTsctsfInfoItem{
						{
							Dnn: "internet",
						},
					},
				},
			},
		},
	}

	if context.Locality != "" {
		profile.Locality = context.Locality
	}
	return profile, nil
}

func (s *nnrfService) SendRegisterNFInstance(ctx context.Context) (
	resouceNrfUri string, retrieveNfInstanceID string, err error,
) {
	// Set client and set url
	tsctsfContext := s.consumer.Context()

	client := s.getNFManagementClient(tsctsfContext.NrfUri)
	nfProfile, err := s.BuildNFInstance(tsctsfContext)
	if err != nil {
		return "", "",
			errors.Wrap(err, "RegisterNFInstance buildNfProfile()")
	}

	var res *NFManagement.RegisterNFInstanceResponse

	finish := false
	for !finish {
		select {
		case <-ctx.Done():
			return "", "", fmt.Errorf("RegisterNFInstance context done")
		default:
			req := &NFManagement.RegisterNFInstanceRequest{
				NfInstanceID:             &tsctsfContext.NfId,
				NrfNfManagementNfProfile: &nfProfile,
			}
			res, err = client.NFInstanceIDDocumentApi.RegisterNFInstance(ctx, req)
			if err != nil || res == nil {
				logger.ConsumerLog.Errorf("TSCTSF register to NRF Error[%v]", err)
				time.Sleep(2 * time.Second)
				continue
			}

			if res.Location == "" {
				// NFUpdate
				finish = true
			} else {
				// NFRegister
				resourceUri := res.Location
				resouceNrfUri = resourceUri[:strings.Index(resourceUri, "/nnrf-nfm/")]
				retrieveNfInstanceID = resourceUri[strings.LastIndex(resourceUri, "/")+1:]

				// oauth2 := false
				// if nf.CustomInfo != nil {
				// 	v, ok := nf.CustomInfo["oauth2"].(bool)
				// 	if ok {
				// 		oauth2 = v
				// 		logger.MainLog.Infoln("OAuth2 setting receive from NRF:", oauth2)
				// 	}
				// }
				// tsctsf_context.GetSelf().OAuth2Required = oauth2
				// if oauth2 && tsctsf_context.GetSelf().NrfCertPem == "" {
				// 	logger.CfgLog.Error("OAuth2 enable but no nrfCertPem provided in config.")
				// }

				finish = true
			}
		}
	}

	return resouceNrfUri, retrieveNfInstanceID, err
}

func (s *nnrfService) SendDeregisterNFInstance() (problemDetails *models.ProblemDetails, err error) {
	logger.ConsumerLog.Infof("Send Deregister NFInstance")

	ctx, pd, err := tsctsf_context.GetSelf().GetTokenCtx(models.ServiceName_NNRF_NFM, models.NrfNfManagementNfType_NRF)
	if err != nil {
		return pd, err
	}

	tsctsfContext := s.consumer.Context()
	client := s.getNFManagementClient(tsctsfContext.NrfUri)
	request := &NFManagement.DeregisterNFInstanceRequest{
		NfInstanceID: &tsctsfContext.NfId,
	}

	_, err = client.NFInstanceIDDocumentApi.DeregisterNFInstance(ctx, request)

	return problemDetails, err
}
