package gosysutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mktemp returns a new temp dir or exits with a fatal error
func mktempdir(t *testing.T) string {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := ioutil.TempDir(pwd, ".test-")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestFsStat(t *testing.T) {
	assert := assert.New(t)
	pwd, err := os.Getwd()
	assert.Nil(err)
	stat, err := FsStatFromPath(pwd)
	assert.Nil(err)
	fmt.Printf("\nSome stats obtained:\nTotal %d, \nFree %d\n\n", stat.Total, stat.Free)
}

func TestFileFalloc(t *testing.T) {
	var sz int64 = 100000 // bytes
	assert := assert.New(t)
	dir := mktempdir(t)
	defer os.RemoveAll(dir)
	fn := path.Join(dir, "reserved")
	if err := FileFallocate(fn, sz, 0664, true); err != nil {
		t.Fatal(err)
	}
	// assert.Nil(err)
	// fmt.Println(err.Error())
	fi, err := os.Stat(fn)
	assert.Nil(err)
	fmt.Printf("File size: specified: %d, actual: %d\n", sz, fi.Size())
	assert.Equal(sz, fi.Size())
}

func TestBindMountDidMount(t *testing.T) {
	assert := assert.New(t)

	dirtgt := mktempdir(t)
	defer os.RemoveAll(dirtgt)
	dirsrc := mktempdir(t)
	defer os.RemoveAll(dirsrc)

	fn := filepath.Join(dirsrc, "somefile.txt")
	testcont := []byte("Some content")

	err := os.WriteFile(fn, testcont, 0644)
	assert.Nil(err)
	if err := MountBind(dirsrc, dirtgt); err != nil {
		t.Fatal(err)
	}
	txt, err := os.ReadFile(fn)
	assert.Nil(err)
	assert.Equal(testcont, txt)

	err = Unmount(dirtgt)
	assert.Nil(err)
}

func TestBindMountAll(t *testing.T) {
	assert := assert.New(t)

	dirtgt := mktempdir(t)
	defer os.RemoveAll(dirtgt)
	dirsrc := mktempdir(t)
	defer os.RemoveAll(dirsrc)

	dirs := []string{"somedir", "anotherdir", "dir3", "dirfour"}
	// define and make source dirs
	fullpaths := make([]string, len(dirs))
	for i, s := range dirs { // prepend source root to dir names
		fullpaths[i] = filepath.Join(dirsrc, s)
		if err := os.Mkdir(fullpaths[i], 0700); err != nil {
			t.Fatal(err)
		}
	}

	// do the mount
	if err := MountBindAll(append(fullpaths, dirtgt)...); err != nil {
		t.Fatal(err)
	}

	// write dummy files in each source dir and check whether they are indeed accessible from the mount
	for i, d := range dirs {
		wcont := []byte(fmt.Sprintf("Some content in dir %s", d))
		fn := fmt.Sprintf("somefile%d.txt", i+1)
		if err := (os.WriteFile(filepath.Join(dirsrc, d, fn), wcont, 0666)); err != nil {
			t.Fatal(err)
		}
		rcont, err := os.ReadFile(filepath.Join(dirtgt, d, fn))
		if err != nil {
			t.Fatal(err)
		}
		assert.EqualValues(wcont, rcont)
	}

	// first unmount one of the dirs, to test if UmountAll correctly ignores "not mounted" errors
	if err := Unmount(filepath.Join(dirtgt, dirs[len(dirs)-1])); err != nil {
		t.Fatal(err)
	}

	if err := UmountAll(dirtgt); err != nil {
		t.Fatal(err)
	}
}
