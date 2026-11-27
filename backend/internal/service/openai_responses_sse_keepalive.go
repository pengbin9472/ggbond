package service

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// doOpenAIResponsesUpstream keeps streaming Responses clients alive even while
// the upstream has not sent HTTP headers. Body readers take over after Do returns.
// A heartbeat is transport traffic, not model output: it must not stop failover
// or affect token usage and first-token timing.
func (s *OpenAIGatewayService) doOpenAIResponsesUpstream(c *gin.Context, request *http.Request, proxyURL string, account *Account, clientStream bool) (*http.Response, error) {
	if clientStream && GetOpenAIClientTransport(c) != OpenAIClientTransportWS && s.cfg != nil {
		stop := startOpenAISSEKeepalive(c, time.Duration(s.cfg.Gateway.StreamKeepaliveInterval)*time.Second)
		defer stop()
	}
	return s.doOpenAIUpstream(request, proxyURL, account)
}

// writeOpenAIJSONError preserves HTTP error responses until a heartbeat commits
// SSE. After that point Responses requires a terminal event instead of raw JSON.
func writeOpenAIJSONError(c *gin.Context, statusCode int, payload gin.H) {
	if !StopOpenAICompactSSEKeepaliveCommitted(c) {
		c.JSON(statusCode, payload)
		return
	}
	body, _ := json.Marshal(payload)
	writeOpenAIResponsesSSEFailure(c, statusCode, body)
}

func writeOpenAIResponsesSSEFailure(c *gin.Context, statusCode int, body []byte) {
	message := sanitizeUpstreamErrorMessage(strings.TrimSpace(extractUpstreamErrorMessage(body)))
	if message == "" {
		message = "Upstream request failed with HTTP " + strconv.Itoa(statusCode)
	}
	errorPayload := map[string]any{"message": message}
	for _, field := range []string{"type", "code", "param"} {
		if value := strings.TrimSpace(gjson.GetBytes(body, "error."+field).String()); value != "" {
			errorPayload[field] = value
		}
	}
	if _, ok := errorPayload["code"]; !ok {
		errorPayload["code"] = "upstream_error"
		if errType, ok := errorPayload["type"]; ok {
			errorPayload["code"] = errType
		}
	}
	writeOpenAIResponsesSSEFailurePayload(c, statusCode, errorPayload)
}
