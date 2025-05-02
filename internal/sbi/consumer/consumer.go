package consumer

import (
	"context"

	"github.com/HanHongChen/openapi-tsctsf/nrf/NFDiscovery"
	"github.com/HanHongChen/openapi-tsctsf/nrf/NFManagement"
	"github.com/HanHongChen/openapi-tsctsf/pcf/PolicyAuthorization"
	tsctsf_context "github.com/HanHongChen/tsctsf/internal/context"
	"github.com/HanHongChen/tsctsf/pkg/factory"
)

type tsctsf interface {
	Config() *factory.Config
	Context() *tsctsf_context.TSCTSFContext
	CancelContext() context.Context
}

type Consumer struct {
	tsctsf

	*nnrfService
	*npcfService
}

func NewConsumer(tsctsf tsctsf) (*Consumer, error) {
	c := &Consumer{
		tsctsf: tsctsf,
	}

	c.nnrfService = &nnrfService{
		consumer:        c,
		nfMngmntClients: make(map[string]*NFManagement.APIClient),
		nfDiscClients:   make(map[string]*NFDiscovery.APIClient),
	}

	c.npcfService = &npcfService{
		consumer:        c,
		nfPolAuthClient: make(map[string]*PolicyAuthorization.APIClient),
	}

	return c, nil
}
