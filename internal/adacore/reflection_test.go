package adacore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

func TestSendWithReflection(t *testing.T) {
	origDataPath, origMockFun := userDataPath, generateContentMock
	t.Cleanup(func() {
		userDataPath = origDataPath
		generateContentMock = origMockFun
	})

	tests := []struct {
		mockFunc       func() (*llms.ContentResponse, error)
		reflectinDepth int
		hasError       bool
	}{
		{func() (*llms.ContentResponse, error) { return nil, fmt.Errorf("mock error") }, 0, true},
		{nil, 0, true},
		{nil, 1, false},
	}
	for _, v := range tests {
		if v.mockFunc != nil {
			generateContentMock = v.mockFunc
		} else {
			generateContentMock = origMockFun
		}
		ai := &mockLLM{}
		cfg := &Config{ReflectionDepth: 0}
		ada := New(ai, cfg)
		res, err := ada.SendWithReflection("some prompt")
		if err == nil {
			assert.Equal(t, []byte("some response"), res)
			assert.True(t, (err != nil) == v.hasError)
		}
	}
}
