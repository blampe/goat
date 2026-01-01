// Copyright 2025 Donald Mullis. All rights reserved.

package internal

import (
	"fmt"
	"runtime"
	"strings"
)

// Return last two elements of path name to source file, plus line number
// 'upDistance' of 1 would return info for immediate caller of Where(1)
// XX  Move to 'debug.go', or to sub-package?
func Where(upDistance int) string {
	_, file, line, ok := runtime.Caller(upDistance)
	if !ok {
		return "UNKNOWN file:line"
	}
	names := strings.Split(file, "/")
	last2 := names[len(names)-2:]
	return fmt.Sprintf("%s:%d", last2[0]+"/"+last2[1], line)
}

func Who(upDistance int) string {
	var pc []uintptr
	found := runtime.Callers(upDistance, pc)
	if found < 1 {
		return "UNKNOWN Caller"
	}
	frames := runtime.CallersFrames(pc)
	nextFrame, more := frames.Next()
	_ = more
	return fmt.Sprintf("%s", nextFrame.Function)
}
