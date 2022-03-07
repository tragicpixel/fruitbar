package utils

// GetMapKeys returns a slice containing all the keys that are set in the supplied map.
func GetMapKeys(m map[string]interface{}) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}
