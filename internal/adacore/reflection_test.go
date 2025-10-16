package adacore

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestSendWithReflection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		mockFunc       func(...any) (*llms.ContentResponse, error)
		reflectinDepth int
		hasError       bool
	}{
		{func(...any) (*llms.ContentResponse, error) { return nil, fmt.Errorf("mock error") }, 0, true},
		{defaultMock, 0, true},
		{defaultMock, 1, false},
	}
	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ai := &mockLLM{v.mockFunc}
			opts := Options{Reflection: ReflectConfig{Depth: 3, Threshhold: 0.95}}
			ada := New(ai, WithOptions(opts))
			res, err := ada.SendReflect("some prompt")
			if err == nil {
				assert.Equal(t, []byte("some response"), res)
				assert.True(t, (err != nil) == v.hasError)
			}
		})
	}
}

func TestImprove(t *testing.T) {
	t.Parallel()
	type tRetVal struct {
		ret string
		err error
	}
	threshold := float32(0.5)
	tests := []struct {
		flags   int
		depth   int
		retval  *tRetVal
		improve string
		hasErr  bool
	}{
		{0b000, 1, nil, "", true},
		{0b001, 1, nil, "", true},
		{0b011, 1, nil, "", true},
		{0b111, 1, nil, "", false},

		{0b111, 1, &tRetVal{"", fmt.Errorf("err")}, "", false},
		{0b111, 1, &tRetVal{`{"scores":{"relevance":1,"accuracy":1,"completeness":1}}`, nil}, "", false},
		{0b111, 1, &tRetVal{`{"a": 1}`, nil}, "", false},
		{0b111, 1, &tRetVal{`{"a": 1}`, nil}, "fail", true},
		{0b111, 0, &tRetVal{`{"a": 1}`, nil}, "fail", true},
		{0b111, 1, &tRetVal{"", nil}, "", false},
		{0b111, 1, &tRetVal{"{\\}}", nil}, "", false},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			f := defaultMock
			if v.retval != nil {
				f = func(args ...any) (content *llms.ContentResponse, err error) {
					content, err = &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: v.retval.ret}}}, v.retval.err
					if len(args) == 0 {
						return
					}
					msgs, ok := args[0].([]llms.MessageContent)
					if !ok || len(msgs) == 0 || len(msgs[len(msgs)-1].Parts) == 0 {
						return
					}
					if t, ok := msgs[len(msgs)-1].Parts[0].(llms.TextContent); !ok || !strings.Contains(t.Text, "fail") {
						return
					}

					return nil, fmt.Errorf("fail")
				}
			}
			root := t.TempDir()
			ai := &mockLLM{f}
			opts := Options{Reflection: ReflectConfig{v.depth, threshold}, PromptsRoot: root}
			ada := New(ai, WithOptions(opts))
			prompts := []string{reflectPrompt, reflectShortPrompt, improvePrompt}
			for i, pr := range prompts {
				if ((1 << i) & v.flags) != 0 {
					bytes := []byte{}
					if pr == improvePrompt {
						bytes = []byte(v.improve)
					}
					err := os.WriteFile(filepath.Join(root, pr), bytes, 0644)
					require.NoError(t, err)
				}
			}
			_, err := ada.Improve("", "")
			if v.hasErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}
