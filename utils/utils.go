package utils

func Contains(elm string, list []string) bool {
	for _, v := range list {
		if v == elm {
			return true
		}
	}
	return false
}

func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func MergeMaps(maps ...map[string]string) map[string]string {
	out := make(map[string]string, 0)
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}
