package project

import (
	"fmt"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNav(t *testing.T) {
	t.Parallel()
	lproj, err := New(t.TempDir())
	require.NoError(t, err)
	err = lproj.AddNav("file1", "txt", []byte("content1"))
	require.NoError(t, err)
	content, err := lproj.NavContent("file1")
	require.NoError(t, err)
	assert.Equal(t, []byte("content1"), content)
	_, err = lproj.NavContent("file2.txt")
	require.Error(t, err)
}

func TestTryLoadNav(t *testing.T) {
	t.Parallel()
	lproj, err := New(t.TempDir())
	require.NoError(t, err)
	fn := "file1"
	cont := []byte("{}")
	err = lproj.AddNav(fn, "txt", cont)
	require.NoError(t, err)
	err = lproj.materializeNav()
	require.NoError(t, err)
	var res []byte
	ok := lproj.TryLoadNav(fmt.Sprintf("%v.%v", fn, "txt"), &res)
	require.True(t, ok)
	assert.Equal(t, cont, res)
	ok = lproj.TryLoadNav("file2.txt", &res)
	require.False(t, ok)
}

func TestAddNav(t *testing.T) {
	t.Parallel()
	tests := []struct {
		key      string
		ext      string
		content  []byte
		expected nav
		err      string
	}{
		{"file1", "txt", []byte("content1"), nav{content: []byte("content1")}, ""},
		{"structure", "json", []byte(`{"a":"b"}`), nav{content: []byte(`{"a":"b"}`)}, ""},
		{"structure", "json", []byte(`invalid json`), nav{}, "parse structure content"},
	}
	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			lproj, err := New(t.TempDir())
			require.NoError(t, err)
			err = lproj.AddNav(v.key, v.ext, v.content)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else if assert.NoError(t, err) {
				nav, ok := lproj.navs[v.key]
				require.True(t, ok)
				assert.Equal(t, v.expected.content, nav.content)
				assert.Equal(t, filepath.Join(lproj.navPath, fmt.Sprintf("%v.%v", v.key, v.ext)), nav.filepath)
			}
		})
	}
}
