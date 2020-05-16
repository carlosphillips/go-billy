package test

import (
	"io/ioutil"
	"os"
	"runtime"

	. "github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/util"
	. "gopkg.in/check.v1"
)

// LinkSuite is a convenient test suite to validate any implementation of
// billy.Link
type LinkSuite struct {
	FS interface {
		Basic
		Dir
		Link
	}
}

func (s *LinkSuite) TestLink(c *C) {
	if runtime.GOOS == "plan9" {
		c.Skip("skipping on Plan 9; links are not supported")
	}
	err := util.WriteFile(s.FS, "file", nil, 0644)
	c.Assert(err, IsNil)

	err = s.FS.Link("file", "link")
	c.Assert(err, IsNil)
}

func (s *LinkSuite) TestLinkNested(c *C) {
	if runtime.GOOS == "plan9" {
		c.Skip("skipping on Plan 9; links are not supported")
	}
	err := util.WriteFile(s.FS, "file", []byte("hello world!"), 0644)
	c.Assert(err, IsNil)

	err = s.FS.Link("file", "linkA")
	c.Assert(err, IsNil)

	err = s.FS.Link("linkA", "linkB")
	c.Assert(err, IsNil)

	fi, err := s.FS.Stat("linkB")
	c.Assert(err, IsNil)
	c.Assert(fi.Name(), Equals, "linkB")
	c.Assert(fi.Size(), Equals, int64(12))
}

func (s *LinkSuite) TestLinkWithNonExistentdTarget(c *C) {
	if runtime.GOOS == "plan9" {
		c.Skip("skipping on Plan 9; links are not supported")
	}
	err := s.FS.Link("file", "link")
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (s *LinkSuite) TestLinkWithExistingLink(c *C) {
	if runtime.GOOS == "plan9" {
		c.Skip("skipping on Plan 9; links are not supported")
	}
	err := util.WriteFile(s.FS, "link", nil, 0644)
	c.Assert(err, IsNil)

	err = s.FS.Link("file", "link")
	c.Assert(err, Not(IsNil))
}

func (s *LinkSuite) TestOpenWithLinkToAbsolutePath(c *C) {
	if runtime.GOOS == "plan9" {
		c.Skip("skipping on Plan 9; links are not supported")
	}
	err := util.WriteFile(s.FS, "dir/file", []byte("foo"), 0644)
	c.Assert(err, IsNil)

	err = s.FS.Link("/dir/file", "dir/link")
	c.Assert(err, IsNil)

	f, err := s.FS.Open("dir/link")
	c.Assert(err, IsNil)

	all, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	c.Assert(string(all), Equals, "foo")
	c.Assert(f.Close(), IsNil)
}

func (s *LinkSuite) TestLinkProperlyNamed(c *C) {
	if runtime.GOOS == "plan9" {
		c.Skip("skipping on Plan 9; links are not supported")
	}

	err := util.WriteFile(s.FS, "dir/file", nil, 0644)
	c.Assert(err, IsNil)

	err = s.FS.Link("dir/file", "link")
	c.Assert(err, IsNil)

	f, err := s.FS.Open("link")
	c.Assert(err, IsNil)
	c.Assert(f.Name(), Equals, "link")

	fis, err := s.FS.ReadDir("/")
	c.Assert(err, IsNil)
	c.Assert(len(fis), Equals, 2)
	c.Assert("link", IsIn, []interface{}{fis[0].Name(), fis[1].Name()})
}

func (s *LinkSuite) TestRenameWithLink(c *C) {
	if runtime.GOOS == "plan9" {
		c.Skip("skipping on Plan 9; links are not supported")
	}

	err := util.WriteFile(s.FS, "dir/file", []byte("foo"), 0644)
	c.Assert(err, IsNil)

	err = s.FS.Link("dir/file", "link")
	c.Assert(err, IsNil)

	err = s.FS.Rename("link", "newlink")
	c.Assert(err, IsNil)

	f, err := s.FS.Open("newlink")
	c.Assert(err, IsNil)
	all, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	c.Assert(string(all), Equals, "foo")
	c.Assert(f.Name(), Equals, "newlink")
	c.Assert(f.Close(), IsNil)
}

func (s *LinkSuite) TestRenameTargetWithLink(c *C) {
	if runtime.GOOS == "plan9" {
		c.Skip("skipping on Plan 9; links are not supported")
	}

	err := util.WriteFile(s.FS, "dir/file", []byte("foo"), 0644)
	c.Assert(err, IsNil)

	err = s.FS.Link("dir/file", "link")
	c.Assert(err, IsNil)

	err = s.FS.Rename("dir/file", "dif/newfile")
	c.Assert(err, IsNil)

	f, err := s.FS.Open("link")
	c.Assert(err, IsNil)

	all, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	c.Assert(string(all), Equals, "foo")
	c.Assert(f.Close(), IsNil)

	fi, err := s.FS.Stat("link")
	c.Assert(err, IsNil)
	c.Assert(fi.Name(), Equals, "link")
}

func (s *LinkSuite) TestRemoveLinkTarget(c *C) {
	if runtime.GOOS == "plan9" {
		c.Skip("skipping on Plan 9; links are not supported")
	}
	err := util.WriteFile(s.FS, "file", []byte("foo"), 0644)
	c.Assert(err, IsNil)

	err = s.FS.Link("file", "link")
	c.Assert(err, IsNil)

	err = s.FS.Remove("file")
	c.Assert(err, IsNil)

	f, err := s.FS.Open("link")
	c.Assert(err, IsNil)

	all, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	c.Assert(string(all), Equals, "foo")
	c.Assert(f.Close(), IsNil)
}

func (s *LinkSuite) TestRemoveLink(c *C) {
	if runtime.GOOS == "plan9" {
		c.Skip("skipping on Plan 9; links are not supported")
	}
	err := util.WriteFile(s.FS, "file", []byte("foo"), 0644)
	c.Assert(err, IsNil)

	err = s.FS.Link("file", "link")
	c.Assert(err, IsNil)

	err = s.FS.Remove("link")
	c.Assert(err, IsNil)

	_, err = s.FS.Open("link")
	c.Assert(os.IsNotExist(err), Equals, true)

	f, err := s.FS.Open("file")
	c.Assert(err, IsNil)

	all, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	c.Assert(string(all), Equals, "foo")
	c.Assert(f.Close(), IsNil)
}

func (s *LinkSuite) TestWriteToLink(c *C) {
	if runtime.GOOS == "plan9" {
		c.Skip("skipping on Plan 9; links are not supported")
	}

	err := util.WriteFile(s.FS, "dir/file", []byte("foo"), 0644)
	c.Assert(err, IsNil)

	err = s.FS.Link("dir/file", "link")
	c.Assert(err, IsNil)

	err = util.WriteFile(s.FS, "link", []byte("bar"), 0644)
	c.Assert(err, IsNil)

	f, err := s.FS.Open("dir/file")
	c.Assert(err, IsNil)

	all, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	c.Assert(string(all), Equals, "bar")
	c.Assert(f.Close(), IsNil)
}

func (s *LinkSuite) TestWriteToTarget(c *C) {
	if runtime.GOOS == "plan9" {
		c.Skip("skipping on Plan 9; links are not supported")
	}

	err := util.WriteFile(s.FS, "dir/file", []byte("foo"), 0644)
	c.Assert(err, IsNil)

	err = s.FS.Link("dir/file", "link")
	c.Assert(err, IsNil)

	err = util.WriteFile(s.FS, "dir/file", []byte("bar"), 0644)
	c.Assert(err, IsNil)

	f, err := s.FS.Open("link")
	c.Assert(err, IsNil)

	all, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	c.Assert(string(all), Equals, "bar")
	c.Assert(f.Close(), IsNil)
}

type isInChecker struct {
	*CheckerInfo
}

var IsIn Checker = &isInChecker{
	&CheckerInfo{Name: "IsIn", Params: []string{"Expected", "In"}},
}

func (checker *isInChecker) Check(params []interface{}, names []string) (result bool, error string) {
	item := params[0]
	candidates := params[1].([]interface{})
	for _, v := range candidates {
		if v == item {
			result = true
			break
		}
	}
	return
}
