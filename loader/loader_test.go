package loader

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/lyft/goruntime/mocks"
	"github.com/stretchr/testify/require"
)

func makeFileInDir(assert *require.Assertions, path string, text string) {
	err := os.MkdirAll(filepath.Dir(path), os.ModeDir|os.ModePerm)
	assert.NoError(err)

	err = ioutil.WriteFile(path, []byte(text), os.ModePerm)
	assert.NoError(err)
}

func TestNilRuntime(t *testing.T) {
	assert := require.New(t)

	loader := New("", "", &mocks.Scope{})
	snapshot := loader.Snapshot()
	assert.Equal("", snapshot.Get("foo"))
	assert.Equal(uint64(100), snapshot.GetInteger("bar", 100))
	assert.True(snapshot.FeatureEnabled("baz", 100))
	assert.False(snapshot.FeatureEnabled("blah", 0))
}

func TestRuntime(t *testing.T) {
	assert := require.New(t)

	// Setup base test directory.
	tempDir, err := ioutil.TempDir("", "runtime_test")
	assert.NoError(err)
	defer os.RemoveAll(tempDir)

	// Make test files for first runtime snapshot.
	makeFileInDir(assert, tempDir+"/testdir1/app/file1", "hello")
	makeFileInDir(assert, tempDir+"/testdir1/app/dir/file2", "world")
	makeFileInDir(assert, tempDir+"/testdir1/app/dir2/file3", "\n 34  ")
	err = ioutil.WriteFile(tempDir+"/testdir1/app/dir2/bad_perms", []byte("blah"), 0)
	assert.NoError(err)
	err = os.Symlink(tempDir+"/testdir1", tempDir+"/current")
	assert.NoError(err)

	loader := New(tempDir+"/current", "app", &mocks.Scope{})
	runtime_update := make(chan int)
	loader.AddUpdateCallback(runtime_update)

	snapshot := loader.Snapshot()
	assert.Equal("", snapshot.Get("foo"))
	assert.Equal(uint64(5), snapshot.GetInteger("foo", 5))
	assert.Equal("hello", snapshot.Get("file1"))
	assert.Equal(uint64(6), snapshot.GetInteger("file1", 6))
	assert.Equal("world", snapshot.Get("dir.file2"))
	assert.Equal(uint64(7), snapshot.GetInteger("dir.file2", 7))
	assert.Equal(uint64(34), snapshot.GetInteger("dir2.file3", 100))

	keys := snapshot.Keys()
	sort.Strings(keys)
	assert.True(reflect.DeepEqual([]string{"dir.file2", "dir2.file3", "file1"}, keys))

	// Make test files for second runtime snapshot.
	makeFileInDir(assert, tempDir+"/testdir2/app/file1", "hello2")
	makeFileInDir(assert, tempDir+"/testdir2/app/dir/file2", "world2")
	makeFileInDir(assert, tempDir+"/testdir2/app/dir2/file3", "100")
	err = os.Symlink(tempDir+"/testdir2", tempDir+"/current_new")
	assert.NoError(err)
	err = os.Rename(tempDir+"/current_new", tempDir+"/current")
	assert.NoError(err)

	<-runtime_update

	snapshot = loader.Snapshot()
	assert.Equal("", snapshot.Get("foo"))
	assert.Equal("hello2", snapshot.Get("file1"))
	assert.Equal("world2", snapshot.Get("dir.file2"))
	assert.Equal(uint64(100), snapshot.GetInteger("dir2.file3", 0))
	assert.True(snapshot.FeatureEnabled("dir2.file3", 0))

	keys = snapshot.Keys()
	sort.Strings(keys)
	assert.True(reflect.DeepEqual([]string{"dir.file2", "dir2.file3", "file1"}, keys))
}
