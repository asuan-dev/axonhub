package responses

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestConvertInputToMessages_MergesReasoningBetweenFunctionCallAndOutput(t *testing.T) {
	items := []Item{
		{Type: "function_call", CallID: "call_1dfb1152-5e5c-4aa1-ad30-8aeb7d62d670", Name: "exec_command", Arguments: `{"cmd":"echo hi"}`},
		{
			Type: "reasoning",
			Summary: []ReasoningSummary{{
				Type: "summary_text",
				Text: "...\n",
			}},
		},
		{Type: "function_call_output", CallID: "call_1dfb1152-5e5c-4aa1-ad30-8aeb7d62d670", Output: &Input{Text: lo.ToPtr("tool result")}},
	}

	messages, err := convertInputToMessages(&Input{Items: items})
	require.NoError(t, err)
	require.Len(t, messages, 2)

	require.Equal(t, "assistant", messages[0].Role)
	require.Len(t, messages[0].ToolCalls, 1)
	require.Equal(t, "call_1dfb1152-5e5c-4aa1-ad30-8aeb7d62d670", messages[0].ToolCalls[0].ID)
	require.Equal(t, "exec_command", messages[0].ToolCalls[0].Function.Name)
	require.NotNil(t, messages[0].ReasoningContent)
	require.Equal(t, "...\n", *messages[0].ReasoningContent)

	require.Equal(t, "tool", messages[1].Role)
	require.NotNil(t, messages[1].ToolCallID)
	require.Equal(t, "call_1dfb1152-5e5c-4aa1-ad30-8aeb7d62d670", *messages[1].ToolCallID)
	require.NotNil(t, messages[1].Content.Content)
	require.Equal(t, "tool result", *messages[1].Content.Content)
}

func TestConvertInputToMessages_MergesReasoningBetweenCustomToolCallAndOutput(t *testing.T) {
	inputText := "SELECT 1"
	items := []Item{
		{Type: "custom_tool_call", CallID: "call_custom_1", Name: "exec", Input: &inputText},
		{
			Type: "reasoning",
			Summary: []ReasoningSummary{{
				Type: "summary_text",
				Text: "checking result shape",
			}},
		},
		{Type: "custom_tool_call_output", CallID: "call_custom_1", Output: &Input{Text: lo.ToPtr("ok")}},
	}

	messages, err := convertInputToMessages(&Input{Items: items})
	require.NoError(t, err)
	require.Len(t, messages, 2)
	require.Equal(t, "assistant", messages[0].Role)
	require.Len(t, messages[0].ToolCalls, 1)
	require.Equal(t, "call_custom_1", messages[0].ToolCalls[0].ID)
	require.NotNil(t, messages[0].ReasoningContent)
	require.Equal(t, "checking result shape", *messages[0].ReasoningContent)
	require.Equal(t, "tool", messages[1].Role)
	require.Equal(t, "call_custom_1", *messages[1].ToolCallID)
}
