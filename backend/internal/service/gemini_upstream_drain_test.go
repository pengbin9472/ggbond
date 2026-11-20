//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/pengbin9472/ggbond/internal/config"
	pkghttputil "github.com/pengbin9472/ggbond/internal/pkg/httputil"
	"github.com/stretchr/testify/require"
)

func TestGeminiUpstreamRequestContextDrainsAfterClientCancellation(t *testing.T) {
	clientCtx, cancelClient := context.WithCancel(context.Background())
	defer cancelClient()
	svc := &GeminiMessagesCompatService{
		cfg: &config.Config{Gateway: config.GatewayConfig{StreamDataIntervalTimeout: 1}},
	}

	requestCtx := svc.upstreamRequestContext(clientCtx)
	attemptCtx, release := pkghttputil.NewUpstreamAttemptContext(requestCtx, time.Second)
	defer release()
	cancelClient()

	require.NoError(t, requestCtx.Err(), "上游请求 context 应脱离客户端取消")
	select {
	case <-attemptCtx.Done():
		t.Fatalf("上游 attempt 应保留配置的排水时间，提前取消：%v", attemptCtx.Err())
	case <-time.After(25 * time.Millisecond):
	}

	select {
	case <-attemptCtx.Done():
		require.ErrorIs(t, attemptCtx.Err(), context.Canceled)
	case <-time.After(2 * time.Second):
		t.Fatal("上游 attempt 未在配置的排水时间后取消")
	}
}

func TestGeminiForwardNative_NonStreamingUsageSurvivesCanceledClientContext(t *testing.T) {
	svc, upstream := newGeminiVertexAPIKeyTestService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c, _ := newGeminiNativeTestContext(t)

	result, err := svc.ForwardNative(ctx, c, geminiVertexAPIKeyTestAccount(""), "gemini-3-flash-preview", "generateContent", false,
		[]byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`))

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 10, result.Usage.InputTokens)
	require.Equal(t, 5, result.Usage.OutputTokens)
	require.NotNil(t, upstream.lastReq)
	require.NoError(t, upstream.lastReq.Context().Err())
	// 请求带有 drain policy；真实 HTTPUpstream 会据此在客户端断开后有界读取。
	attemptCtx, release := pkghttputil.NewUpstreamAttemptContext(upstream.lastReq.Context(), time.Second)
	defer release()
	require.NoError(t, attemptCtx.Err())

}
