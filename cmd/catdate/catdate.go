package main

import (
	"os"
	"time"

	"github.com/djherbis/times"
	"github.com/pluveto/noteman/internal/pkg"
)

func main() {
	pkg.Assert(len(os.Args) > 1, "missing argument")
	file := os.Args[1]
	t, err := times.Stat(file)
	pkg.Assert(err == nil, "failed to stat file: "+file)
	println(t.ModTime().Format(time.RFC3339Nano))

}
