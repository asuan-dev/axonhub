package anthropic

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/llm"
	"github.com/looplj/axonhub/llm/httpclient"
	openai "github.com/looplj/axonhub/llm/transformer/openai"
	"github.com/looplj/axonhub/llm/transformer/openai/responses"
)

func TestAnthropicDocumentContentDiagnosesLossyDowngradeToChatAndResponses(t *testing.T) {
	body := []byte(`{
		"model": "claude-3-sonnet-20240229",
		"max_tokens": 1024,
		"messages": [{
			"role": "user",
			"content": [
				{"type": "text", "text": "summarize"},
				{
					"type": "document",
					"source": {
						"type": "base64",
						"media_type": "application/pdf",
						"data": "JVBERi0xLjQ="
					}
				}
			]
		}]
	}`)
	inbound := NewInboundTransformer()

	requireHasLossy := func(t *testing.T, req *llm.Request, target llm.APIFormat) {
		t.Helper()
		found := false
		for _, d := range llm.LossyDowngrades(req) {
			if d.SourceProtocol == llm.APIFormatAnthropicMessage &&
				d.SourceField == "content[].type=document" &&
				d.TargetProtocol == target &&
				d.Reason == llm.LossyDowngradeReasonNoEquivalentSemantics {
				found = true
				break
			}
		}
		require.Truef(t, found, "missing document LossyDowngrade -> %s: %#v", target, llm.LossyDowngrades(req))
	}

	respLLMReq, err := inbound.TransformRequest(t.Context(), &httpclient.Request{
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    body,
	})
	require.NoError(t, err)
	require.Equal(t, "document", respLLMReq.Messages[0].Content.MultipleContent[1].Type)
	respOutbound, err := responses.NewOutboundTransformer("https://api.openai.com", "test-key")
	require.NoError(t, err)
	respReq, err := respOutbound.TransformRequest(t.Context(), respLLMReq)
	require.NoError(t, err)
	require.NotContains(t, string(respReq.Body), `"type":"document"`)
	require.NotContains(t, string(respReq.Body), "JVBERi0xLjQ=")
	requireHasLossy(t, respLLMReq, llm.APIFormatOpenAIResponse)

	chatLLMReq, err := inbound.TransformRequest(t.Context(), &httpclient.Request{
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    body,
	})
	require.NoError(t, err)
	chatOutbound, err := openai.NewOutboundTransformer("https://api.openai.com", "test-key")
	require.NoError(t, err)
	chatReq, err := chatOutbound.TransformRequest(t.Context(), chatLLMReq)
	require.NoError(t, err)
	require.NotContains(t, string(chatReq.Body), `"type":"document"`)
	require.NotContains(t, string(chatReq.Body), "JVBERi0xLjQ=")
	requireHasLossy(t, chatLLMReq, llm.APIFormatOpenAIChatCompletion)
}
