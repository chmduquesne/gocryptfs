package fusefrontend_reverse

import (
	"fmt"
	"syscall"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

type virtualFile struct {
	// Embed nodefs.defaultFile for a ENOSYS implementation of all methods
	nodefs.File
	// file content
	content []byte
	// absolute path to a parent file
	parentFile string
	// inode number
	ino uint64
}

func (rfs *reverseFS) NewVirtualFile(content []byte, parentFile string) (nodefs.File, fuse.Status) {
	return &virtualFile{
		File:       nodefs.NewDefaultFile(),
		content:    content,
		parentFile: parentFile,
		ino:        rfs.inoGen.next(),
	}, fuse.OK
}

// Read - FUSE call
func (f *virtualFile) Read(buf []byte, off int64) (resultData fuse.ReadResult, status fuse.Status) {
	if off >= int64(len(f.content)) {
		return nil, fuse.OK
	}
	end := int(off) + len(buf)
	if end > len(f.content) {
		end = len(f.content)
	}
	return fuse.ReadResultData(f.content[off:end]), fuse.OK
}

// GetAttr - FUSE call
func (f *virtualFile) GetAttr(a *fuse.Attr) fuse.Status {
	var st syscall.Stat_t
	err := syscall.Lstat(f.parentFile, &st)
	if err != nil {
		fmt.Printf("Lstat %q: %v\n", f.parentFile, err)
		return fuse.ToStatus(err)
	}
	st.Ino = f.ino
	st.Size = int64(len(f.content))
	st.Mode = syscall.S_IFREG | 0400
	st.Nlink = 1
	a.FromStat(&st)
	return fuse.OK
}
