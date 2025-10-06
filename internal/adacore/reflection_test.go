package adacore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

func TestSendWithReflection(t *testing.T) {
	origDataPath := userDataPath
	t.Cleanup(func() {
		userDataPath = origDataPath
	})

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
