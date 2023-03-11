package avalanche

import (
	"io/fs"
	"net"
	"time"
)

func Ping(url string) bool {
	tcp, e := net.DialTimeout("tcp", url, time.Second)
	if e == nil {
		_ = tcp.Close()
	}
	return e == nil
}

func Traverse(efs fs.FS) (files []string, err error) {
	if err := fs.WalkDir(efs, "bundle", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		return nil, err
	}
	return files, nil
}
