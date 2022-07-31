package pkg

import "fmt"

func Assert(cond bool, msg string, args ...interface{}) {
	if !cond {
		panic(fmt.Sprintf(msg, args...))
	}
}
