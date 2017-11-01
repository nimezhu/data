package data

import (
	"testing"
)

func TestGetRoot(t *testing.T) {
	roots := []string{"/home/zhuxp/1", "/home/zhuxp/2", "/home/zhuxp/image/3"}
	r := getRootDir(roots)
	t.Log(r)
}
