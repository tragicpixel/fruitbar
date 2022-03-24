package utils

// IsStringInSlice returns true if the supplied string matches a string in the supplied slice of strings.
func IsStringInSlice(s string, slice []string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
