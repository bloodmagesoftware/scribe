package util

func TrimSliceEmptyString(in []string) []string {
	out := make([]string, 0, len(in))
	for _, x := range in {
		if len(x) != 0 {
			out = append(out, x)
		}
	}
	return out
}
