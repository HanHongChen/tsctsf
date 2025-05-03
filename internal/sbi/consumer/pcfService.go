package consumer

import (
	"context"
	"errors"
	"sync"

	"github.com/HanHongChen/openapi-tsctsf/pcf/PolicyAuthorization"
	"github.com/HanHongChen/tsctsf/internal/logger"
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

func (s *npcfService) DetNetAppSessionCreate(uri string, req *PolicyAuthorization.PostAppSessionsRequest) (
	resp *PolicyAuthorization.PostAppSessionsResponse, err error) {
	logger.PolicyAuthLog.Logger.Infoln("DetNetAppSessionCreate start")

	if uri == "" {
		return resp, errors.New("Npcf client uri is nil")
	}

	client := s.getPolicyAuthorizationClient(uri)
	if client == nil {
		return resp, errors.New("Npcf client is nil")
	}

	logger.PolicyAuthLog.Infoln("start to send Individual App Session context creation request")
	resp, err = client.ApplicationSessionsCollectionApi.PostAppSessions(
		context.Background(), req,
	)

	if err == nil {
		logger.PolicyAuthLog.Infof("send Individual App Session context create successfully")
		return resp, nil
	}

	return resp, err
}
