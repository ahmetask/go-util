package stream

import "runtime"

func getThreadCount(desiredThreadCount ...int) int {
	if len(desiredThreadCount) == 0 {
		return 1
	}

	maxCores := runtime.NumCPU()

	if desiredThreadCount[0] > maxCores {
		return maxCores
	} else if desiredThreadCount[0] > 0 {
		return desiredThreadCount[0]
	}

	return 1
}
