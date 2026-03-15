// Package args provides shared helper functions for extracting typed values
// from MCP tool argument maps (map[string]interface{}).
// This eliminates repeated boilerplate across all agk MCP tool servers.
package args

// GetString safely extracts a string from an argument map.
// Returns the empty string if the key is absent or the value is not a string.
func GetString(args map[string]interface{}, key string) string {
	v, _ := args[key].(string)
	return v
}

// GetStringOrDefault returns the string value for key, or defaultVal if absent/wrong type.
func GetStringOrDefault(args map[string]interface{}, key, defaultVal string) string {
	if v, ok := args[key].(string); ok && v != "" {
		return v
	}
	return defaultVal
}

// GetStringSlice safely extracts a []string from an argument map.
// The value must be []interface{} where each element is a string.
// Returns nil if the key is absent or the value has the wrong type.
func GetStringSlice(args map[string]interface{}, key string) []string {
	raw, ok := args[key].([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// GetInt safely extracts an int from an argument map.
// JSON numbers are float64 in Go's interface{} unmarshalling.
// Returns 0 if the key is absent or the value is not numeric.
func GetInt(args map[string]interface{}, key string) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	return 0
}

// GetIntOrDefault returns the int value for key, or defaultVal if absent/zero.
func GetIntOrDefault(args map[string]interface{}, key string, defaultVal int) int {
	if v, ok := args[key].(float64); ok && v > 0 {
		return int(v)
	}
	return defaultVal
}

// GetBool safely extracts a bool from an argument map.
// Returns false if the key is absent or the value is not a bool.
func GetBool(args map[string]interface{}, key string) bool {
	v, _ := args[key].(bool)
	return v
}
