package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// ─── estimateTextTokens ─────────────────────────────────────────────

func TestEstimateTextTokens_Empty(t *testing.T) {
	require.Equal(t, 0, estimateTextTokens(""))
}

func TestEstimateTextTokens_ShortText(t *testing.T) {
	// "hello" = 5 bytes, 5/4 = 1
	require.Equal(t, 1, estimateTextTokens("hello"))
}

func TestEstimateTextTokens_LongText(t *testing.T) {
	text := strings.Repeat("a", 4096)
	require.Equal(t, 1024, estimateTextTokens(text))
}

// ─── estimateSystemTokenCount ────────────────────────────────────────

func TestEstimateSystemTokenCount_StringSystem(t *testing.T) {
	data := map[string]any{
		"system": strings.Repeat("x", 400),
	}
	require.Equal(t, 100, estimateSystemTokenCount(data))
}

func TestEstimateSystemTokenCount_BlockArraySystem(t *testing.T) {
	data := map[string]any{
		"system": []any{
			map[string]any{"type": "text", "text": strings.Repeat("x", 200)},
			map[string]any{"type": "text", "text": strings.Repeat("y", 200)},
		},
	}
	require.Equal(t, 100, estimateSystemTokenCount(data))
}

func TestEstimateSystemTokenCount_NoSystem(t *testing.T) {
	data := map[string]any{}
	require.Equal(t, 0, estimateSystemTokenCount(data))
}

// ─── estimateMessagesTokenCount ──────────────────────────────────────

func TestEstimateMessagesTokenCount_StringContent(t *testing.T) {
	messages := []any{
		map[string]any{"role": "user", "content": strings.Repeat("a", 400)},
		map[string]any{"role": "assistant", "content": strings.Repeat("b", 800)},
	}
	// upToIdx=0: 400/4 = 100
	require.Equal(t, 100, estimateMessagesTokenCount(messages, 0))
	// upToIdx=1: (400+800)/4 = 300
	require.Equal(t, 300, estimateMessagesTokenCount(messages, 1))
}

func TestEstimateMessagesTokenCount_BlockContent(t *testing.T) {
	messages := []any{
		map[string]any{
			"role": "user",
			"content": []any{
				map[string]any{"type": "text", "text": strings.Repeat("x", 800)},
			},
		},
	}
	require.Equal(t, 200, estimateMessagesTokenCount(messages, 0))
}

// ─── injectAutoPromptCache: 基本场景 ─────────────────────────────────

func TestInjectAutoPromptCache_SystemStringShort_NoInjection(t *testing.T) {
	// system 太短（< 1024 tokens），只有一条 user 消息也太短 → 不注入
	input := map[string]any{
		"model":  "claude-sonnet-4-20250514",
		"system": "Be helpful.",
		"messages": []any{
			map[string]any{"role": "user", "content": "Hi"},
		},
	}
	body, _ := json.Marshal(input)
	result := injectAutoPromptCache(body)

	var output map[string]any
	require.NoError(t, json.Unmarshal(result, &output))

	// system 应保持原样（字符串，不转换为块数组）
	_, isString := output["system"].(string)
	require.True(t, isString, "short system should remain a string, not be converted")
}

func TestInjectAutoPromptCache_SystemStringLong_InjectsSystem(t *testing.T) {
	// system 足够长（>= 1024 tokens = 4096 chars）→ 注入 system
	longSystem := strings.Repeat("You are a helpful assistant. ", 200) // ~5600 chars
	input := map[string]any{
		"model":  "claude-sonnet-4-20250514",
		"system": longSystem,
		"messages": []any{
			map[string]any{"role": "user", "content": "Hi"},
		},
	}
	body, _ := json.Marshal(input)
	result := injectAutoPromptCache(body)

	var output map[string]any
	require.NoError(t, json.Unmarshal(result, &output))

	// system 应转换为块数组并带有 cache_control
	systemArr, ok := output["system"].([]any)
	require.True(t, ok, "long system string should be converted to block array")
	require.Len(t, systemArr, 1)
	block := systemArr[0].(map[string]any)
	require.Equal(t, "text", block["type"])
	cc, exists := block["cache_control"]
	require.True(t, exists, "system block should have cache_control")
	require.Equal(t, "ephemeral", cc.(map[string]any)["type"])
}

func TestInjectAutoPromptCache_SystemBlockArray_InjectsLastTextBlock(t *testing.T) {
	longText := strings.Repeat("x", 5000)
	input := map[string]any{
		"model": "claude-sonnet-4-20250514",
		"system": []any{
			map[string]any{"type": "text", "text": "First block"},
			map[string]any{"type": "text", "text": longText},
		},
		"messages": []any{
			map[string]any{"role": "user", "content": "Hi"},
		},
	}
	body, _ := json.Marshal(input)
	result := injectAutoPromptCache(body)

	var output map[string]any
	require.NoError(t, json.Unmarshal(result, &output))

	systemArr := output["system"].([]any)
	// 第一个 text 块不应有 cache_control
	firstBlock := systemArr[0].(map[string]any)
	_, exists := firstBlock["cache_control"]
	require.False(t, exists, "first text block should NOT have cache_control")

	// 最后一个 text 块应有 cache_control
	lastBlock := systemArr[1].(map[string]any)
	cc, exists := lastBlock["cache_control"]
	require.True(t, exists, "last text block should have cache_control")
	require.Equal(t, "ephemeral", cc.(map[string]any)["type"])
}

// ─── injectAutoPromptCache: messages 断点位置 ────────────────────────

func TestInjectAutoPromptCache_MultiTurn_MarksBeforeLastUser(t *testing.T) {
	// 多轮对话：断点应在最后一条 user 消息之前（assistant 消息上）
	longSystem := strings.Repeat("system prompt text. ", 300)
	input := map[string]any{
		"model":  "claude-sonnet-4-20250514",
		"system": longSystem,
		"messages": []any{
			map[string]any{"role": "user", "content": "First question"},
			map[string]any{"role": "assistant", "content": "First answer with some detail."},
			map[string]any{"role": "user", "content": "Second question"},
			map[string]any{"role": "assistant", "content": "Second answer with more detail."},
			map[string]any{"role": "user", "content": "Third question (new)"},
		},
	}
	body, _ := json.Marshal(input)
	result := injectAutoPromptCache(body)

	var output map[string]any
	require.NoError(t, json.Unmarshal(result, &output))

	msgs := output["messages"].([]any)

	// messages[3]（倒数第二条 assistant 消息）应有 cache_control
	assistantMsg := msgs[3].(map[string]any)
	// content 应被转换为块数组
	contentArr, ok := assistantMsg["content"].([]any)
	require.True(t, ok, "assistant content should be converted to block array")
	lastBlock := contentArr[len(contentArr)-1].(map[string]any)
	cc, exists := lastBlock["cache_control"]
	require.True(t, exists, "message before last user should have cache_control")
	require.Equal(t, "ephemeral", cc.(map[string]any)["type"])

	// messages[4]（最后一条 user 消息）不应有 cache_control
	lastUserMsg := msgs[4].(map[string]any)
	switch content := lastUserMsg["content"].(type) {
	case string:
		// 纯字符串 → 没有 cache_control，OK
	case []any:
		for _, item := range content {
			block := item.(map[string]any)
			_, exists := block["cache_control"]
			require.False(t, exists, "last user message should NOT have cache_control")
		}
	}
}

func TestInjectAutoPromptCache_SingleMessage_MarksIt(t *testing.T) {
	// 首轮对话（只有一条 user 消息）：仍标记以覆盖重试场景
	longSystem := strings.Repeat("detailed system prompt. ", 300)
	input := map[string]any{
		"model":  "claude-sonnet-4-20250514",
		"system": longSystem,
		"messages": []any{
			map[string]any{"role": "user", "content": "Hello world"},
		},
	}
	body, _ := json.Marshal(input)
	result := injectAutoPromptCache(body)

	var output map[string]any
	require.NoError(t, json.Unmarshal(result, &output))

	msgs := output["messages"].([]any)
	userMsg := msgs[0].(map[string]any)
	contentArr, ok := userMsg["content"].([]any)
	require.True(t, ok, "single user content should be converted to block array")
	block := contentArr[0].(map[string]any)
	_, exists := block["cache_control"]
	require.True(t, exists, "single user message should have cache_control for retry coverage")
}

func TestInjectAutoPromptCache_TwoMessages_MarksFirstUserInsteadOfLastUser(t *testing.T) {
	// 两条消息 [user, assistant]... 这种情况 last user 是 index 0，不是典型的请求形式
	// 但如果出现，lastUserIdx=0 且 len > 1，不标记
	longSystem := strings.Repeat("system. ", 600)
	input := map[string]any{
		"model":  "claude-sonnet-4-20250514",
		"system": longSystem,
		"messages": []any{
			map[string]any{"role": "user", "content": "Q1"},
			map[string]any{"role": "assistant", "content": "A1"},
		},
	}
	body, _ := json.Marshal(input)
	result := injectAutoPromptCache(body)

	var output map[string]any
	require.NoError(t, json.Unmarshal(result, &output))

	msgs := output["messages"].([]any)
	// 不应标记任何 messages（因为 lastUserIdx=0 且 len>1）
	for i, m := range msgs {
		msg := m.(map[string]any)
		switch content := msg["content"].(type) {
		case []any:
			for _, item := range content {
				block := item.(map[string]any)
				_, exists := block["cache_control"]
				require.False(t, exists, "message[%d] should NOT have cache_control in this edge case", i)
			}
		case string:
			// 纯字符串 → 没被修改 → OK
		}
	}
}

func TestInjectAutoPromptCache_ThreeMessages_MarksAssistant(t *testing.T) {
	// 标准 3 条消息：[user, assistant, user]
	// 断点应在 assistant（index 1）上
	longSystem := strings.Repeat("prompt. ", 600)
	input := map[string]any{
		"model":  "claude-sonnet-4-20250514",
		"system": longSystem,
		"messages": []any{
			map[string]any{"role": "user", "content": "Q1"},
			map[string]any{
				"role": "assistant",
				"content": []any{
					map[string]any{"type": "text", "text": "A1 with details"},
				},
			},
			map[string]any{"role": "user", "content": "Q2"},
		},
	}
	body, _ := json.Marshal(input)
	result := injectAutoPromptCache(body)

	var output map[string]any
	require.NoError(t, json.Unmarshal(result, &output))

	msgs := output["messages"].([]any)
	// messages[1]（assistant）应有 cache_control
	asstMsg := msgs[1].(map[string]any)
	contentArr := asstMsg["content"].([]any)
	block := contentArr[0].(map[string]any)
	_, exists := block["cache_control"]
	require.True(t, exists, "assistant message (index 1) should have cache_control")

	// messages[2]（最后 user）不应有
	lastUserMsg := msgs[2].(map[string]any)
	_, isString := lastUserMsg["content"].(string)
	require.True(t, isString, "last user content should remain unchanged string")
}

// ─── injectAutoPromptCache: 最小 token 数检查 ────────────────────────

func TestInjectAutoPromptCache_BelowMinTokens_NoInjection(t *testing.T) {
	// 所有内容合计不足 1024 tokens → 不注入任何标记
	input := map[string]any{
		"model":  "claude-sonnet-4-20250514",
		"system": "Short system prompt.",
		"messages": []any{
			map[string]any{"role": "user", "content": "Short question."},
			map[string]any{"role": "assistant", "content": "Short answer."},
			map[string]any{"role": "user", "content": "Another short question."},
		},
	}
	body, _ := json.Marshal(input)
	result := injectAutoPromptCache(body)

	// 输入输出应完全一致（无注入）
	var inputData, outputData map[string]any
	json.Unmarshal(body, &inputData)
	json.Unmarshal(result, &outputData)

	// 检查 system 没有被转换
	_, isString := outputData["system"].(string)
	require.True(t, isString, "system should remain string when below threshold")

	// 检查 messages 中没有 cache_control
	msgs := outputData["messages"].([]any)
	for i, m := range msgs {
		msg := m.(map[string]any)
		switch content := msg["content"].(type) {
		case []any:
			for _, item := range content {
				if block, ok := item.(map[string]any); ok {
					_, exists := block["cache_control"]
					require.False(t, exists, "message[%d] should NOT have cache_control below threshold", i)
				}
			}
		}
	}
}

func TestInjectAutoPromptCache_SystemShortButMessagesMeetThreshold(t *testing.T) {
	// system 太短不缓存，但 system + messages 累计达到 1024 → 标记 messages 中的断点
	input := map[string]any{
		"model":  "claude-sonnet-4-20250514",
		"system": "Be helpful.", // 很短
		"messages": []any{
			map[string]any{"role": "user", "content": strings.Repeat("detailed question. ", 100)},
			map[string]any{"role": "assistant", "content": strings.Repeat("detailed answer. ", 200)},
			map[string]any{"role": "user", "content": "Follow up question"},
		},
	}
	body, _ := json.Marshal(input)
	result := injectAutoPromptCache(body)

	var output map[string]any
	require.NoError(t, json.Unmarshal(result, &output))

	// system 不应被标记（太短）
	_, isString := output["system"].(string)
	require.True(t, isString, "short system should remain string")

	// messages[1]（assistant）应有 cache_control（累计 token 数达标）
	msgs := output["messages"].([]any)
	asstMsg := msgs[1].(map[string]any)
	contentArr, ok := asstMsg["content"].([]any)
	require.True(t, ok)
	block := contentArr[0].(map[string]any)
	_, exists := block["cache_control"]
	require.True(t, exists, "messages breakpoint should be injected when cumulative tokens meet threshold")
}

func TestInjectAutoPromptCache_HaikuModel_HigherThreshold(t *testing.T) {
	// Haiku 模型的最低门槛是 2048 tokens
	// 内容在 1024-2048 之间 → Sonnet 会注入，但 Haiku 不应注入
	mediumSystem := strings.Repeat("x", 5000) // ~1250 tokens，超过 1024 但不到 2048

	inputHaiku := map[string]any{
		"model":  "claude-3-5-haiku-20241022",
		"system": mediumSystem,
		"messages": []any{
			map[string]any{"role": "user", "content": "Hello"},
		},
	}
	bodyHaiku, _ := json.Marshal(inputHaiku)
	resultHaiku := injectAutoPromptCache(bodyHaiku)

	var outputHaiku map[string]any
	require.NoError(t, json.Unmarshal(resultHaiku, &outputHaiku))

	// Haiku：1250 tokens < 2048 → 不注入
	_, isString := outputHaiku["system"].(string)
	require.True(t, isString, "haiku model with medium system should NOT inject (below 2048 threshold)")

	// 对比：同样内容用 Sonnet → 应注入
	inputSonnet := map[string]any{
		"model":  "claude-sonnet-4-20250514",
		"system": mediumSystem,
		"messages": []any{
			map[string]any{"role": "user", "content": "Hello"},
		},
	}
	bodySonnet, _ := json.Marshal(inputSonnet)
	resultSonnet := injectAutoPromptCache(bodySonnet)

	var outputSonnet map[string]any
	require.NoError(t, json.Unmarshal(resultSonnet, &outputSonnet))

	// Sonnet：1250 tokens >= 1024 → 注入
	systemArr, ok := outputSonnet["system"].([]any)
	require.True(t, ok, "sonnet model with medium system should inject")
	block := systemArr[0].(map[string]any)
	_, exists := block["cache_control"]
	require.True(t, exists)
}

// ─── injectAutoPromptCache: 边界情况 ─────────────────────────────────

func TestInjectAutoPromptCache_ExistingCacheControl_NotDuplicated(t *testing.T) {
	longSystem := strings.Repeat("x", 5000)
	input := map[string]any{
		"model": "claude-sonnet-4-20250514",
		"system": []any{
			map[string]any{
				"type":          "text",
				"text":          longSystem,
				"cache_control": map[string]any{"type": "ephemeral"},
			},
		},
		"messages": []any{
			map[string]any{"role": "user", "content": "Q1"},
			map[string]any{
				"role": "assistant",
				"content": []any{
					map[string]any{
						"type":          "text",
						"text":          "A1",
						"cache_control": map[string]any{"type": "ephemeral"},
					},
				},
			},
			map[string]any{"role": "user", "content": "Q2"},
		},
	}
	body, _ := json.Marshal(input)
	result := injectAutoPromptCache(body)

	var output map[string]any
	require.NoError(t, json.Unmarshal(result, &output))

	// 应保持不变（已有 cache_control 的块不重复注入）
	count := countCacheControlBlocks(output)
	require.Equal(t, 2, count, "existing cache_control should not be duplicated")
}

func TestInjectAutoPromptCache_MaxQuotaReached_NoInjection(t *testing.T) {
	longSystem := strings.Repeat("x", 5000)
	input := map[string]any{
		"model": "claude-sonnet-4-20250514",
		"system": []any{
			map[string]any{
				"type":          "text",
				"text":          longSystem,
				"cache_control": map[string]any{"type": "ephemeral"},
			},
		},
		"messages": []any{
			map[string]any{
				"role": "user",
				"content": []any{
					map[string]any{"type": "text", "text": "Q1", "cache_control": map[string]any{"type": "ephemeral"}},
				},
			},
			map[string]any{
				"role": "assistant",
				"content": []any{
					map[string]any{"type": "text", "text": "A1", "cache_control": map[string]any{"type": "ephemeral"}},
				},
			},
			map[string]any{
				"role": "user",
				"content": []any{
					map[string]any{"type": "text", "text": "Q2", "cache_control": map[string]any{"type": "ephemeral"}},
				},
			},
		},
	}
	body, _ := json.Marshal(input)
	result := injectAutoPromptCache(body)

	var output map[string]any
	require.NoError(t, json.Unmarshal(result, &output))

	count := countCacheControlBlocks(output)
	require.Equal(t, 4, count, "should not inject beyond max quota of 4")
}

func TestInjectAutoPromptCache_InvalidJSON_ReturnsOriginal(t *testing.T) {
	body := []byte(`{invalid json}`)
	result := injectAutoPromptCache(body)
	require.Equal(t, body, result)
}

func TestInjectAutoPromptCache_NoMessages_OnlySystemInjected(t *testing.T) {
	longSystem := strings.Repeat("x", 5000)
	input := map[string]any{
		"model":    "claude-sonnet-4-20250514",
		"system":   longSystem,
		"messages": []any{},
	}
	body, _ := json.Marshal(input)
	result := injectAutoPromptCache(body)

	var output map[string]any
	require.NoError(t, json.Unmarshal(result, &output))

	// system 应被注入
	systemArr, ok := output["system"].([]any)
	require.True(t, ok)
	block := systemArr[0].(map[string]any)
	_, exists := block["cache_control"]
	require.True(t, exists, "system should be injected even with empty messages")
}

func TestInjectAutoPromptCache_SystemWithThinkingBlock_SkipsThinking(t *testing.T) {
	longText := strings.Repeat("x", 5000)
	input := map[string]any{
		"model": "claude-sonnet-4-20250514",
		"system": []any{
			map[string]any{"type": "text", "text": longText},
			map[string]any{"type": "thinking", "text": "internal thinking"},
		},
		"messages": []any{
			map[string]any{"role": "user", "content": "Hi"},
		},
	}
	body, _ := json.Marshal(input)
	result := injectAutoPromptCache(body)

	var output map[string]any
	require.NoError(t, json.Unmarshal(result, &output))

	systemArr := output["system"].([]any)
	// text 块应有 cache_control
	textBlock := systemArr[0].(map[string]any)
	_, exists := textBlock["cache_control"]
	require.True(t, exists, "text block should have cache_control")

	// thinking 块不应有 cache_control
	thinkingBlock := systemArr[1].(map[string]any)
	_, exists = thinkingBlock["cache_control"]
	require.False(t, exists, "thinking block should NOT have cache_control")
}

// ─── injectAutoPromptCache: 实际多轮对话场景 ─────────────────────────

func TestInjectAutoPromptCache_RealWorldMultiTurn_CacheHitScenario(t *testing.T) {
	// 模拟真实场景：验证连续请求间缓存断点的稳定性
	longSystem := strings.Repeat("You are a coding assistant with deep expertise. ", 100)

	// 请求 N：3 轮对话
	requestN := map[string]any{
		"model":  "claude-sonnet-4-20250514",
		"system": longSystem,
		"messages": []any{
			map[string]any{"role": "user", "content": "Write a function to sort an array"},
			map[string]any{"role": "assistant", "content": strings.Repeat("Here's the implementation. ", 50)},
			map[string]any{"role": "user", "content": "Can you optimize it?"},
			map[string]any{"role": "assistant", "content": strings.Repeat("Here's the optimized version. ", 50)},
			map[string]any{"role": "user", "content": "Add error handling"},
		},
	}
	bodyN, _ := json.Marshal(requestN)
	resultN := injectAutoPromptCache(bodyN)

	var outputN map[string]any
	require.NoError(t, json.Unmarshal(resultN, &outputN))

	msgsN := outputN["messages"].([]any)
	// 断点应在 messages[3]（第二条 assistant，最后一条 user 之前）
	asstMsg := msgsN[3].(map[string]any)
	contentArr, ok := asstMsg["content"].([]any)
	require.True(t, ok)
	block := contentArr[0].(map[string]any)
	_, exists := block["cache_control"]
	require.True(t, exists, "request N should mark message[3] (assistant before last user)")

	// 最后一条 user 消息不应被标记
	lastUser := msgsN[4].(map[string]any)
	_, isString := lastUser["content"].(string)
	require.True(t, isString, "last user message should not be modified")

	// 请求 N+1：增加了一轮对话
	requestN1 := map[string]any{
		"model":  "claude-sonnet-4-20250514",
		"system": longSystem,
		"messages": []any{
			map[string]any{"role": "user", "content": "Write a function to sort an array"},
			map[string]any{"role": "assistant", "content": strings.Repeat("Here's the implementation. ", 50)},
			map[string]any{"role": "user", "content": "Can you optimize it?"},
			map[string]any{"role": "assistant", "content": strings.Repeat("Here's the optimized version. ", 50)},
			map[string]any{"role": "user", "content": "Add error handling"},
			map[string]any{"role": "assistant", "content": strings.Repeat("Here's the version with error handling. ", 50)},
			map[string]any{"role": "user", "content": "Write tests for it"},
		},
	}
	bodyN1, _ := json.Marshal(requestN1)
	resultN1 := injectAutoPromptCache(bodyN1)

	var outputN1 map[string]any
	require.NoError(t, json.Unmarshal(resultN1, &outputN1))

	msgsN1 := outputN1["messages"].([]any)
	// 断点应在 messages[5]（第三条 assistant，最后一条 user 之前）
	asstMsgN1 := msgsN1[5].(map[string]any)
	contentArrN1, ok := asstMsgN1["content"].([]any)
	require.True(t, ok)
	blockN1 := contentArrN1[0].(map[string]any)
	_, existsN1 := blockN1["cache_control"]
	require.True(t, existsN1, "request N+1 should mark message[5] (assistant before last user)")
}

// ─── injectCacheControlOnMessage ─────────────────────────────────────

func TestInjectCacheControlOnMessage_StringContent(t *testing.T) {
	msg := map[string]any{
		"role":    "assistant",
		"content": "Hello world",
	}
	count := 0
	injectCacheControlOnMessage(msg, &count, 2)

	require.Equal(t, 1, count)
	contentArr, ok := msg["content"].([]any)
	require.True(t, ok)
	block := contentArr[0].(map[string]any)
	require.Equal(t, "Hello world", block["text"])
	cc := block["cache_control"].(map[string]string)
	require.Equal(t, "ephemeral", cc["type"])
}

func TestInjectCacheControlOnMessage_BlockArrayContent(t *testing.T) {
	msg := map[string]any{
		"role": "assistant",
		"content": []any{
			map[string]any{"type": "text", "text": "First"},
			map[string]any{"type": "text", "text": "Second"},
		},
	}
	count := 0
	injectCacheControlOnMessage(msg, &count, 2)

	require.Equal(t, 1, count)
	contentArr := msg["content"].([]any)
	// 第一个块不应有
	first := contentArr[0].(map[string]any)
	_, exists := first["cache_control"]
	require.False(t, exists)
	// 最后一个块应有
	last := contentArr[1].(map[string]any)
	_, exists = last["cache_control"]
	require.True(t, exists)
}

func TestInjectCacheControlOnMessage_AlreadyAtMaxInject(t *testing.T) {
	msg := map[string]any{
		"role":    "assistant",
		"content": "Hello",
	}
	count := 2
	injectCacheControlOnMessage(msg, &count, 2)

	// 不应注入（已达上限）
	require.Equal(t, 2, count)
	_, isString := msg["content"].(string)
	require.True(t, isString, "content should not be modified when at max inject limit")
}

func TestInjectCacheControlOnMessage_NotAMap(t *testing.T) {
	count := 0
	injectCacheControlOnMessage("not a map", &count, 2)
	require.Equal(t, 0, count)
}
