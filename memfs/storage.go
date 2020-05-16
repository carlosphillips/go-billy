package memfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type storage struct {
	files    map[string]*fileNode
	children map[string]map[string]*fileNode
}

func newStorage() *storage {
	return &storage{
		files:    make(map[string]*fileNode, 0),
		children: make(map[string]map[string]*fileNode, 0),
	}
}

func (s *storage) Has(path string) bool {
	path = clean(path)

	_, ok := s.files[path]
	return ok
}

func (s *storage) New(path string, mode os.FileMode, flag int) (*file, error) {
	path = clean(path)
	if s.Has(path) {
		if !s.MustGet(path).mode.IsDir() {
			return nil, fmt.Errorf("file already exists %q", path)
		}

		return nil, nil
	}

	name := filepath.Base(path)

	n := &fileNode{
		content: &content{},
		mode:    mode,
	}
	f := &file{
		name:     name,
		flag:     flag,
		fileNode: n,
	}

	s.files[path] = n
	s.createParent(path, mode, n)
	return f, nil
}

func (s *storage) createParent(path string, mode os.FileMode, n *fileNode) error {
	base, name := filepath.Split(path)
	base = clean(base)
	if name == "" {
		return nil
	}

	if _, err := s.New(base, mode.Perm()|os.ModeDir, 0); err != nil {
		return err
	}

	if _, ok := s.children[base]; !ok {
		s.children[base] = make(map[string]*fileNode, 0)
	}

	s.children[base][name] = n
	return nil
}

func (s *storage) Children(path string) []*file {
	path = clean(path)

	l := make([]*file, 0)
	for p, n := range s.children[path] {
		f := prepareFile(p, n)
		l = append(l, f)
	}

	return l
}

func (s *storage) MustGet(path string) *file {
	f, ok := s.Get(path)
	if !ok {
		panic(fmt.Errorf("couldn't find %q", path))
	}

	return f
}

func (s *storage) Get(path string) (*file, bool) {
	path = clean(path)
	if !s.Has(path) {
		return nil, false
	}

	n, ok := s.files[path]
	if !ok {
		return nil, ok
	}
	file := prepareFile(path, n)
	return file, ok
}

func prepareFile(path string, n *fileNode) *file {
	path = clean(path)
	base := filepath.Base(path)
	return &file{
		fileNode: n,
		name:     base,
	}
}

func (s *storage) Link(target, link string) error {
	target = clean(target)
	link = clean(link)
	linkBase := filepath.Dir(link)
	linkBase = clean(linkBase)

	f, ok := s.Get(target)
	if !ok {
		return &os.LinkError{"link", target, link, os.ErrNotExist}
	}

	if s.Has(link) {
		return &os.LinkError{"link", target, link, os.ErrExist}
	}

	d, ok := s.Get(linkBase)
	if !ok || !d.mode.IsDir() {
		return &os.LinkError{"link", target, link, os.ErrNotExist}
	}

	s.files[link] = f.fileNode
	s.createParent(link, 0666, f.fileNode)
	return nil
}

func (s *storage) Rename(from, to string) error {
	from = clean(from)
	to = clean(to)

	if !s.Has(from) {
		return os.ErrNotExist
	}

	move := [][2]string{{from, to}}

	for pathFrom := range s.files {
		if pathFrom == from || !filepath.HasPrefix(pathFrom, from) {
			continue
		}

		rel, _ := filepath.Rel(from, pathFrom)
		pathTo := filepath.Join(to, rel)

		move = append(move, [2]string{pathFrom, pathTo})
	}

	for _, ops := range move {
		from := ops[0]
		to := ops[1]

		if err := s.move(from, to); err != nil {
			return err
		}
	}

	return nil
}

func (s *storage) move(from, to string) error {
	s.files[to] = s.files[from]
	s.children[to] = s.children[from]

	defer func() {
		delete(s.children, from)
		delete(s.files, from)
		delete(s.children[filepath.Dir(from)], filepath.Base(from))
	}()

	return s.createParent(to, 0644, s.files[to])
}

func (s *storage) Remove(path string) error {
	path = clean(path)

	n, has := s.Get(path)
	if !has {
		return os.ErrNotExist
	}

	if n.mode.IsDir() && len(s.children[path]) != 0 {
		return fmt.Errorf("dir: %s contains files", path)
	}

	base, name := filepath.Split(path)
	base = filepath.Clean(base)

	delete(s.children[base], name)
	delete(s.files, path)
	return nil
}

func clean(path string) string {
	return filepath.Clean(filepath.FromSlash(path))
}

type content struct {
	bytes []byte
}

func (c *content) WriteAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, errors.New("negative offset")
	}

	prev := len(c.bytes)

	diff := int(off) - prev
	if diff > 0 {
		c.bytes = append(c.bytes, make([]byte, diff)...)
	}

	c.bytes = append(c.bytes[:off], p...)
	if len(c.bytes) < prev {
		c.bytes = c.bytes[:prev]
	}

	return len(p), nil
}

func (c *content) ReadAt(b []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, errors.New("negative offset")
	}

	size := int64(len(c.bytes))
	if off >= size {
		return 0, io.EOF
	}

	l := int64(len(b))
	if off+l > size {
		l = size - off
	}

	btr := c.bytes[off : off+l]
	if len(btr) < len(b) {
		err = io.EOF
	}
	n = copy(b, btr)

	return
}
