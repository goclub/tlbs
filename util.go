package tlbs

func safeSlice[T any](arr []T, cidx []int) []T {
	if len(cidx) != 2 {
		return []T{}
	}
	start := cidx[0]
	end := cidx[1]
	if len(arr) == 0 {
		return []T{}
	}
	// 确保 start 和 end 在数组/切片范围内
	if start < 0 {
		start = 0
	}
	if end > len(arr) {
		end = len(arr)
	}

	// 创建新的子切片并返回
	return arr[start:end]
}
