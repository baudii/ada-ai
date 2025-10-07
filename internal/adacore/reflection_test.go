package adacore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

func TestSendWithReflection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		mockFunc       func() (*llms.ContentResponse, error)
		reflectinDepth int
		hasError       bool
	}{
		{func() (*llms.ContentResponse, error) { return nil, fmt.Errorf("mock error") }, 0, true},
		{defaultMock, 0, true},
		{defaultMock, 1, false},
	}
	for _, v := range tests {
		ai := &mockLLM{v.mockFunc}
		cfg := &Config{Reflection: ReflectConfig{Depth: 3, Threshhold: 0.95}}
		ada := New(ai, cfg)
		res, err := ada.SendReflect("some prompt")
		if err == nil {
			assert.Equal(t, []byte("some response"), res)
			assert.True(t, (err != nil) == v.hasError)
		}
	}
}

// func TestImprove(t *testing.T) {
// 	ai := &mockLLM{}
// 	cfg := &Config{Reflection: ReflectConfig{1, 0.5}}
// 	root := t.TempDir()
// 	ada := New(ai, cfg, WithRoot(root))

// }

func TestAvg(t *testing.T) {
	tests := []struct {
		r reflection
		e float32
	}{
		{reflection{Scores: score{1, 1, 1}, Suggestions: []string{}}, 1},
		{reflection{Scores: score{0.2, 0.9, 0.4}, Suggestions: []string{}}, 0.5},
		{reflection{Scores: score{0, 0, 0}, Suggestions: []string{}}, 0},
	}

	for _, v := range tests {
		assert.Equal(t, v.e, v.r.avg())
	}
}

func TestLoadTemplates(t *testing.T) {

}
