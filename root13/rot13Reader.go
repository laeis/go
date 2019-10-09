package main

import (
	"io"
	"os"
	"strings"
)

type rot13Reader struct {
	r io.Reader
}

func change(m byte,start byte) byte {
	return start+(m-start+13)%26
}
func root13(i byte) byte {
	switch {
		case i >= 'A' && i <= 'Z' :
			i = change(i, 'A')
		case i >= 'a' && i <= 'z' :
			i = change(i, 'a')	
	}	
	return i
	
}

func (r13 rot13Reader) Read(p []byte) (int, error) {
	var n int
	var err error
	n, err = r13.r.Read(p)
	for i, _ := range p[:n] {
		p[i] = root13(p[i])
	}
	return n, err
}
func main() {
	s := strings.NewReader("Lbh penpxrq gur pbqr!")
	
	r := rot13Reader{s}
	io.Copy(os.Stdout, &r)
}
