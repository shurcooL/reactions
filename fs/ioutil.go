package fs

import (
	"encoding/json"
	"errors"

	"github.com/shurcooL/webdavfs/vfsutil"
	"golang.org/x/net/webdav"
)

// jsonEncodeFile encodes v into file at path, overwriting or creating it.
func jsonEncodeFile(fs webdav.FileSystem, path string, v interface{}) error {
	f, err := vfsutil.Create(fs, path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(v)
}

// jsonDecodeFile decodes contents of file at path into v.
func jsonDecodeFile(fs webdav.FileSystem, path string, v interface{}) error {
	f, err := vfsutil.Open(fs, path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}

var errIsDir = errors.New("is a directory")

// jsonDecodeFileNotDir decodes contents of file at path into v.
// Before decoding, it checks if file is a directory, returns errIsDir if so.
func jsonDecodeFileNotDir(fs webdav.FileSystem, path string, v interface{}) error {
	f, err := vfsutil.Open(fs, path)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return errIsDir
	}
	return json.NewDecoder(f).Decode(v)
}
