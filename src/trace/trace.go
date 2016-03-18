package trace
import (
       "runtime"
       "fmt"
)

func T(dbg string) string {
	pc := make([]uintptr, 10)  // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	s:= fmt.Sprintf("%s:%d %s\n%s\n", file, line, f.Name(), dbg)
	fmt.Printf("%s\n", s)
	return s
}
