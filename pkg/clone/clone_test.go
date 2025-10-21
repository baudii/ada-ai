package clone_test

import (
	"testing"

	"github.com/baudii/ada-ai/pkg/clone"
	"github.com/stretchr/testify/assert"
)

func TestDeepCopyMap(t *testing.T) {
	t.Parallel()
	s1 := struct{ s string }{"hello"}
	s2 := struct{ s any }{"world"}
	var (
		nilSl  []int          = nil
		nilPtr *int           = nil
		nilMap map[string]any = nil
		nilIfc any
	)
	m1 := map[string]any{
		"a": "bbb",
		"b": map[string]any{"c": 123, "s": nil},
		"d": []int{1, 2, 3},
		"e": nil,
		"f": &s1,
		"g": nilSl,
		"h": nilMap,
		"i": nilPtr,
		"j": nilIfc,
	}

	m3 := clone.DeepCopyMap(nil)
	m2 := clone.DeepCopyMap(m1)
	m1["a"] = "zzz"
	m1["d"].([]int)[0] = 100
	delete(m1, "b")
	m1["e"] = "hello"
	m1["f"] = &s2

	assert.Nil(t, m3)
	assert.Equal(t, map[string]any{"a": "zzz", "d": []int{100, 2, 3}, "e": "hello", "f": &s2, "g": nilSl, "h": nilMap, "i": nilPtr, "j": nilIfc}, m1)
	assert.Equal(t, map[string]any{
		"a": "bbb",
		"b": map[string]any{"c": 123, "s": nil},
		"d": []int{1, 2, 3},
		"e": nil,
		"f": &s1,
		"g": nilSl,
		"h": nilMap,
		"i": nilPtr,
		"j": nilIfc,
	}, m2)
}
