package util

func StringOr(val, _default string) string {
	if val == "" {
		return _default
	}

	return val
}
