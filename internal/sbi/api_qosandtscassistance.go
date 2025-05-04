package sbi

import (
	"net/http"

	"github.com/HanHongChen/openapi-tsctsf"
	"github.com/HanHongChen/openapi-tsctsf/models"
	"github.com/HanHongChen/tsctsf/internal/logger"
	"github.com/gin-gonic/gin"
)

func (s *Server) getQoSAnsTscRoutes() []Route {
	return []Route{
		{
			Name:    "Index",
			Method:  http.MethodGet,
			Pattern: "/",
			APIFunc: s.Index,
		},
		{
			Name:    "PostTSCAppSessions",
			Method:  http.MethodPost,
			Pattern: "/tsc-app-sessions",
			APIFunc: s.HTTPPostTSCAppSessions,
		},
		{
			Name:    "GetTSCAppSession",
			Method:  http.MethodGet,
			Pattern: "/tsc-app-sessions/:appSessionId",
			APIFunc: s.HTTPGetTSCAppSession,
		},
		{
			Name:    "ModAppSession",
			Method:  http.MethodPatch,
			Pattern: "/tsc-app-sessions/:appSessionId",
			APIFunc: s.HTTPModAppSession,
		},
		{
			Name:    "DeleteTSCAppSession",
			Method:  http.MethodDelete,
			Pattern: "/tsc-app-sessions/:appSessionId/delete",
			APIFunc: s.HTTPDeleteTSCAppSession,
		},
		{
			Name:    "putEventsSubsc",
			Method:  http.MethodPut,
			Pattern: "/tsc-app-sessions/:appSessionId/events-subscription",
			APIFunc: s.HTTPPutEventsSubsc,
		},
		{
			Name:    "DeleteEventsSubsc",
			Method:  http.MethodDelete,
			Pattern: "/tsc-app-sessions/:appSessionId/events-subscription",
			APIFunc: s.HTTPDeleteEventsSubsc,
		},
	}
}

// Index is the index handler.
func (s *Server) Index(c *gin.Context) {
	c.JSON(http.StatusOK, "Hello World!")
}

func (s *Server) HTTPPostTSCAppSessions(c *gin.Context) {
	var tscAppSessionContext models.TscAppSessionContextData
	// step 1: retrieve http request body
	requestBody, err := c.GetRawData()
	if err != nil {
		problemDetail := models.ProblemDetails{
			Title:  "System failure",
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
			Cause:  "SYSTEM_FAILURE",
		}
		logger.TSCAppSessLog.Errorf("Get Request Body error: %+v", err)
		c.JSON(http.StatusInternalServerError, problemDetail)
		return
	}

	// step 2: convert requestBody to openapi models
	err = openapi.Deserialize(&tscAppSessionContext, requestBody, "application/json")
	if err != nil {
		problemDetail := "[Request Body] " + err.Error()
		rsp := models.ProblemDetails{
			Title:  "Malformed request syntax",
			Status: http.StatusBadRequest,
			Detail: problemDetail,
		}
		logger.TSCAppSessLog.Errorln(problemDetail)
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	s.Processor().HandlePostTSCAppSession(c, tscAppSessionContext)
}

func (s *Server) HTTPGetTSCAppSession(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{})
}

func (s *Server) HTTPModAppSession(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{})
}

func (s *Server) HTTPDeleteTSCAppSession(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{})
}

func (s *Server) HTTPPutEventsSubsc(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{})
}

func (s *Server) HTTPDeleteEventsSubsc(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{})
}
