package core

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/uget/uget/core/api"
)

// File denotes a remote file object
//
// For any given File, the order of method calls must be:
//     1. `Err()` - if this returns `nil`, continue with checking the file's availability:
//     2. `Offline()` - and if this also returns false, the file is valid and available.
// If `Err()` returns an error, `Offline()` and all non-URL methods will panic.
// Same for when `Offline()` returns `true`.
type File interface {
	api.File

	// Err returns the error associated with this file, if there is any.
	//
	// Read the note on call order in the interface description..
	Err() error

	// Offline returns whether this file is offline.
	//
	// Read the note on call order in the interface description.
	Offline() bool

	// LengthUnknown returns whether this file's length is known
	// e.g. HEAD request without Content-Length.
	LengthUnknown() bool

	// ID returns the identifier for this file (sha256-sum of the URL string)
	ID() string

	// OriginalURL returns the original URL (passed to Client) that ultimately yielded this File.
	OriginalURL() *url.URL

	// done callback when this file is done downloading.
	// also ensures File is not implemented outside this package.
	done()

	// SeqNum is the order in which the file was popped from the queue
	SeqNum() int

	setPopOrder(int)
}

var _ File = &onlineFile{}
var _ File = &offlineFile{}
var _ File = &erroredFile{}

func online(f api.File, orig *url.URL, done func()) File {
	return &onlineFile{file{f, orig, -1}, done}
}

func offline(orig, curr *url.URL) File {
	return &offlineFile{file{nil, orig, -1}, curr}
}

func errored(orig, curr *url.URL, err error) File {
	return &erroredFile{file{nil, orig, -1}, curr, err}
}

type file struct {
	api.File
	original *url.URL
	popOrder int
}

func (f file) OriginalURL() *url.URL { return f.original }
func (f file) ID() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(f.URL().String())))
}
func (f *file) setPopOrder(o int)     { f.popOrder = o }
func (f *file) SeqNum() int           { logrus.Debugf("file#seqNum: %v", f.popOrder); return f.popOrder }

func (f file) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{
		"id":       f.ID(),
		"provider": f.Provider().Name(),
		"name":     f.Name(),
		"url":      f.URL().String(),
		"size":     f.Size(),
	}
	if cks, algo, _ := f.Checksum(); cks != nil {
		data["checksum"] = map[string]string{"algo": strings.ToLower(algo), "sum": fmt.Sprintf("%x", cks)}
	}
	return json.Marshal(data)
}

type onlineFile struct {
	file
	donecb func()
}

func (f *onlineFile) Err() error          { return nil }
func (f *onlineFile) Offline() bool       { return false }
func (f *onlineFile) LengthUnknown() bool { return f.Size() == api.FileSizeUnknown }
func (f *onlineFile) done()               { f.donecb() }

type offlineFile struct {
	file
	u *url.URL
}

func (f *offlineFile) Err() error          { return nil }
func (f *offlineFile) Offline() bool       { return true }
func (f *offlineFile) LengthUnknown() bool { panic("LengthUnknown() on offline file") }
func (f *offlineFile) done()               { panic("done() on offline file") }
func (f *offlineFile) URL() *url.URL       { return f.u }

type erroredFile struct {
	file
	u   *url.URL
	err error
}

func (f *erroredFile) Err() error          { return f.err }
func (f *erroredFile) Offline() bool       { panic("Offline() on errored file") }
func (f *erroredFile) LengthUnknown() bool { panic("LengthUnknown() on errored file") }
func (f *erroredFile) done()               { panic("done() on errored file") }
func (f *erroredFile) URL() *url.URL       { return f.u }
