package getsize

const (
	_  = iota
	KB = 1 << (iota * 10)
	MB
	GB
	TB
)

func GetSize(size int64) (float64, string) {
	switch {
	case size < KB:
		return float64(size), "B"
	case size < MB:
		return float64(size) / KB, "KB"
	case size < GB:
		return float64(size) / MB, "MB"
	case size < TB:
		return float64(size) / GB, "GB"
	default:
		return float64(size) / TB, "TB"
	}
}
