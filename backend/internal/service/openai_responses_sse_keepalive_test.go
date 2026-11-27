package service

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pengbin9472/ggbond/internal/config"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type responseHeaderKeepaliveUpstream struct {
	httpUpstreamRecorder
	recorder      *openAIResponseFlushRecorder
	beforeHeaders string
	headerDelay   time.Duration
}

func (u *responseHeaderKeepaliveUpstream) Do(req *http.Request, proxyURL string, accountID int64, concurrency int) (*http.Response, error) {
	if u.headerDelay > 0 {
		time.Sleep(u.headerDelay)
	} else {
		select {
		case <-u.recorder.flushEvents:
		case <-time.After(3 * time.Second):
			return nil, errors.New("no downstream keepalive while waiting for upstream headers")
		}
	}
	u.beforeHeaders, _ = u.recorder.snapshot()
	return u.httpUpstreamRecorder.Do(req, proxyURL, accountID, concurrency)
}

func TestOpenAIResponsesHeaderKeepaliveForward(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, failure := range []bool{false, true} {
			name := "native"
			if passthrough {
				name = "passthrough"
			}
			if failure {
				name += "/error"
			} else {
				name += "/success"
			}
			t.Run(name, func(t *testing.T) {
				recorder := newOpenAIResponseFlushRecorder()
				c, _ := gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
					Body: io.NopCloser(strings.NewReader(
						"data: {\"type\":\"response.output_text.delta\",\"delta\":\"hello\"}\n\n" +
							"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_test\",\"usage\":{\"input_tokens\":7,\"output_tokens\":3,\"total_tokens\":10}}}\n\n")),
				}
				if failure {
					resp.StatusCode = http.StatusBadRequest
					resp.Header.Set("Content-Type", "application/json")
					resp.Body = io.NopCloser(strings.NewReader(`{"error":{"type":"invalid_request_error","code":"bad_input","param":"input","message":"Invalid input"}}`))
				}
				upstream := &responseHeaderKeepaliveUpstream{httpUpstreamRecorder: httpUpstreamRecorder{resp: resp}, recorder: recorder}
				svc := &OpenAIGatewayService{
					cfg: &config.Config{Gateway: config.GatewayConfig{
						StreamKeepaliveInterval: 1,
						MaxLineSize:             defaultMaxLineSize,
					}},
					httpUpstream: upstream,
				}
				account := &Account{
					ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Concurrency: 1,
					Status: StatusActive, Schedulable: true,
					Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://upstream.example"},
					Extra:       map[string]any{"openai_passthrough": passthrough},
				}
				result, err := svc.Forward(c.Request.Context(), c, account, []byte(`{"model":"gpt-5.2","stream":true,"input":"hello"}`))
				require.Equal(t, "", stripKeepaliveComments(upstream.beforeHeaders), "no model output before upstream headers")
				require.Contains(t, upstream.beforeHeaders, ": keepalive\n\n")
				body, _ := recorder.snapshot()
				require.Equal(t, http.StatusOK, recorder.status)
				require.Equal(t, "text/event-stream", recorder.Header().Get("Content-Type"))
				require.Equal(t, "no", recorder.Header().Get("X-Accel-Buffering"))
				if failure {
					require.Error(t, err)
					events := parseCompactBridgeSSE(t, stripKeepaliveComments(body))
					require.Len(t, events, 1)
					require.Equal(t, "response.failed", events[0][0])
					if !passthrough {
						require.Equal(t, "bad_input", gjson.Get(events[0][1], "response.error.code").String())
						require.Equal(t, "input", gjson.Get(events[0][1], "response.error.param").String())
					}
				} else {
					require.NoError(t, err)
					require.NotNil(t, result)
					require.Equal(t, 7, result.Usage.InputTokens)
					require.Equal(t, 3, result.Usage.OutputTokens)
					require.NotNil(t, result.FirstTokenMs)
					require.GreaterOrEqual(t, *result.FirstTokenMs, 1000)
					require.Contains(t, body, "response.completed")
					require.NotContains(t, body, "response.failed")
				}
			})
		}
	}
}

func TestOpenAIResponsesHeaderKeepalivePreservesNonStreamingAndFastErrors(t *testing.T) {
	for _, tc := range []struct {
		name     string
		stream   bool
		interval int
		delay    time.Duration
	}{
		{"non_streaming", false, 1, 1100 * time.Millisecond},
		{"disabled", true, 0, 1100 * time.Millisecond},
		{"fast_error", true, 10, time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := newOpenAIResponseFlushRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			svc := &OpenAIGatewayService{
				cfg: &config.Config{Gateway: config.GatewayConfig{StreamKeepaliveInterval: tc.interval}},
				httpUpstream: &responseHeaderKeepaliveUpstream{
					recorder: recorder, headerDelay: tc.delay,
					httpUpstreamRecorder: httpUpstreamRecorder{resp: &http.Response{
						StatusCode: http.StatusBadRequest, Header: http.Header{},
						Body: io.NopCloser(strings.NewReader(`{"error":{"message":"Invalid input"}}`)),
					}},
				},
			}
			resp, err := svc.doOpenAIResponsesUpstream(c, c.Request, "", &Account{ID: 1}, tc.stream)
			require.NoError(t, err)
			defer resp.Body.Close()
			writeOpenAIJSONError(c, http.StatusBadRequest, gin.H{"error": gin.H{"message": "Invalid input"}})
			require.Equal(t, http.StatusBadRequest, recorder.status)
			body, _ := recorder.snapshot()
			require.JSONEq(t, `{"error":{"message":"Invalid input"}}`, body)
		})
	}
}

func TestOpenAIResponsesKeepaliveRestartsPreserveFailoverAccounting(t *testing.T) {
	c, recorder := newPassthroughKeepaliveTestContext(t)
	originalWriter := c.Writer
	for range 3 {
		stop := startOpenAISSEKeepalive(c, time.Hour)
		value, _ := c.Get(openAICompactSSEKeepaliveKey)
		require.True(t, value.(*openAICompactSSEKeepalive).beat())
		stop()
		require.Same(t, originalWriter, c.Writer)
		require.True(t, StopOpenAICompactSSEKeepaliveCommitted(c))
		require.Equal(t, -1, OpenAICompactKeepaliveAdjustedWrittenSize(c))
		require.False(t, openAIStreamClientOutputStarted(c, false))
	}
	require.Equal(t, 3, strings.Count(recorder.Body.String(), ": keepalive\n\n"))
	_, err := c.Writer.Write([]byte("data: real\n\n"))
	require.NoError(t, err)
	require.Equal(t, len("data: real\n\n"), OpenAICompactKeepaliveAdjustedWrittenSize(c))
}

func TestOpenAIResponsesKeepaliveReusesActiveCompactWriter(t *testing.T) {
	c, recorder := newCompactBridgeTestContext(t, true)
	stopOuter := StartOpenAICompactSSEKeepalive(c, time.Hour)
	defer stopOuter()
	originalWrapper := c.Writer
	stopInner := startOpenAISSEKeepalive(c, time.Hour)
	stopInner()
	require.Same(t, originalWrapper, c.Writer)
	value, _ := c.Get(openAICompactSSEKeepaliveKey)
	require.True(t, value.(*openAICompactSSEKeepalive).beat(), "inner stop must leave outer heartbeat running")
	require.True(t, StopOpenAICompactSSEKeepaliveCommitted(c))
	require.Contains(t, recorder.Body.String(), ": keepalive\n\n")
}

func TestOpenAIResponsesHeaderKeepaliveThenAccountFailover(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		name := "native"
		if passthrough {
			name = "passthrough"
		}
		t.Run(name, func(t *testing.T) {
			recorder := newOpenAIResponseFlushRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			svc := &OpenAIGatewayService{
				cfg: &config.Config{Gateway: config.GatewayConfig{StreamKeepaliveInterval: 1, MaxLineSize: defaultMaxLineSize}},
				httpUpstream: &responseHeaderKeepaliveUpstream{
					recorder: recorder,
					httpUpstreamRecorder: httpUpstreamRecorder{resp: &http.Response{
						StatusCode: http.StatusServiceUnavailable, Header: http.Header{},
						Body: io.NopCloser(strings.NewReader(`{"error":{"message":"Temporarily unavailable"}}`)),
					}},
				},
			}
			account := &Account{
				ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Concurrency: 1,
				Status: StatusActive, Schedulable: true,
				Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://upstream.example"},
				Extra:       map[string]any{"openai_passthrough": passthrough},
			}
			body := []byte(`{"model":"gpt-5.2","stream":true,"input":"hello"}`)
			_, err := svc.Forward(c.Request.Context(), c, account, body)
			var failoverErr *UpstreamFailoverError
			require.ErrorAs(t, err, &failoverErr)
			require.Equal(t, -1, OpenAICompactKeepaliveAdjustedWrittenSize(c), "heartbeats must not prevent account switching")

			svc.httpUpstream = &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"text/event-stream"}},
				Body: io.NopCloser(strings.NewReader("data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_second\",\"usage\":{\"input_tokens\":2,\"output_tokens\":1}}}\n\n")),
			}}
			account.ID = 2
			result, err := svc.Forward(c.Request.Context(), c, account, body)
			require.NoError(t, err)
			require.Equal(t, 2, result.Usage.InputTokens)
			got, _ := recorder.snapshot()
			require.Contains(t, got, ": keepalive\n\n")
			require.Contains(t, got, "resp_second")
			require.NotContains(t, got, "response.failed")
			require.NotContains(t, got, "Temporarily unavailable")
		})
	}
}

func TestOpenAIResponsesHeaderKeepaliveDoesNotResetFirstOutputTimeout(t *testing.T) {
	c, recorder := newPassthroughKeepaliveTestContext(t)
	upstream := &blockingOpenAIResponseHeaderUpstream{canceled: make(chan struct{})}
	// The header mock's fallback is 1.5s. A 1s first-output timeout must still
	// cancel it even though a heartbeat has already committed the response.
	stop := startOpenAISSEKeepalive(c, time.Hour)
	defer stop()
	value, _ := c.Get(openAICompactSSEKeepaliveKey)
	require.True(t, value.(*openAICompactSSEKeepalive).beat())
	svc := &OpenAIGatewayService{
		cfg: &config.Config{Gateway: config.GatewayConfig{
			OpenAIFirstOutputTimeoutSeconds: 1, StreamKeepaliveInterval: 1, MaxLineSize: defaultMaxLineSize,
		}},
		httpUpstream: upstream,
	}
	account := &Account{
		ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Concurrency: 1,
		Status: StatusActive, Schedulable: true,
		Credentials: map[string]any{"access_token": "test-token", "chatgpt_account_id": "test-account"},
	}
	_, err := svc.Forward(c.Request.Context(), c, account, []byte(`{"model":"gpt-5.5","stream":true,"reasoning":{"effort":"low"},"input":"hello"}`))
	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.Equal(t, http.StatusGatewayTimeout, failoverErr.StatusCode)
	require.Contains(t, string(failoverErr.ResponseBody), "first_output_timeout")
	require.True(t, failoverErr.SafeToFailoverAfterWrite)
	require.True(t, StopOpenAICompactSSEKeepaliveCommitted(c))
	require.Equal(t, -1, OpenAICompactKeepaliveAdjustedWrittenSize(c))
	require.Contains(t, recorder.Body.String(), ": keepalive\n\n")
	select {
	case <-upstream.canceled:
	default:
		t.Fatal("first-output timeout did not cancel the upstream header wait")
	}
}
