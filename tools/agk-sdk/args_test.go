package args

import "testing"

func TestGetString_Present(t *testing.T) {
	m := map[string]interface{}{"key": "hello"}
	if got := GetString(m, "key"); got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
}

func TestGetString_Absent(t *testing.T) {
	m := map[string]interface{}{}
	if got := GetString(m, "key"); got != "" {
		t.Errorf("expected '', got %q", got)
	}
}

func TestGetString_WrongType(t *testing.T) {
	m := map[string]interface{}{"key": 42}
	if got := GetString(m, "key"); got != "" {
		t.Errorf("expected '', got %q", got)
	}
}

func TestGetStringOrDefault(t *testing.T) {
	m := map[string]interface{}{}
	if got := GetStringOrDefault(m, "key", "default"); got != "default" {
		t.Errorf("expected 'default', got %q", got)
	}
	m["key"] = "value"
	if got := GetStringOrDefault(m, "key", "default"); got != "value" {
		t.Errorf("expected 'value', got %q", got)
	}
}

func TestGetStringSlice_Normal(t *testing.T) {
	m := map[string]interface{}{
		"tags": []interface{}{"go", "rust", "python"},
	}
	got := GetStringSlice(m, "tags")
	if len(got) != 3 {
		t.Errorf("expected 3 elements, got %d", len(got))
	}
	if got[0] != "go" {
		t.Errorf("expected 'go', got %q", got[0])
	}
}

func TestGetStringSlice_Absent(t *testing.T) {
	m := map[string]interface{}{}
	if got := GetStringSlice(m, "tags"); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestGetStringSlice_MixedTypes(t *testing.T) {
	m := map[string]interface{}{
		"items": []interface{}{"a", 42, "b"},
	}
	got := GetStringSlice(m, "items")
	// Non-string elements are skipped
	if len(got) != 2 {
		t.Errorf("expected 2 string elements, got %d: %v", len(got), got)
	}
}

func TestGetInt_Present(t *testing.T) {
	m := map[string]interface{}{"count": float64(5)}
	if got := GetInt(m, "count"); got != 5 {
		t.Errorf("expected 5, got %d", got)
	}
}

func TestGetInt_Absent(t *testing.T) {
	m := map[string]interface{}{}
	if got := GetInt(m, "count"); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestGetIntOrDefault(t *testing.T) {
	m := map[string]interface{}{}
	if got := GetIntOrDefault(m, "count", 10); got != 10 {
		t.Errorf("expected 10, got %d", got)
	}
	m["count"] = float64(7)
	if got := GetIntOrDefault(m, "count", 10); got != 7 {
		t.Errorf("expected 7, got %d", got)
	}
}

func TestGetBool(t *testing.T) {
	m := map[string]interface{}{"flag": true}
	if !GetBool(m, "flag") {
		t.Error("expected true")
	}
	if GetBool(m, "absent") {
		t.Error("expected false for absent key")
	}
}
