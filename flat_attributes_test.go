package sawmill

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestFlatAttributesBasicOperations(t *testing.T) {
	attrs := NewFlatAttributes()

	// Test Set and Get
	attrs.Set([]string{"key1"}, "value1")
	attrs.Set([]string{"nested", "key2"}, "value2")

	value1, ok1 := attrs.Get([]string{"key1"})
	if !ok1 || value1 != "value1" {
		t.Errorf("Expected 'value1', got %v, %v", value1, ok1)
	}

	value2, ok2 := attrs.Get([]string{"nested", "key2"})
	if !ok2 || value2 != "value2" {
		t.Errorf("Expected 'value2', got %v, %v", value2, ok2)
	}

	// Test non-existent key
	nonExistent, ok3 := attrs.Get([]string{"nonexistent"})
	if ok3 || nonExistent != nil {
		t.Errorf("Expected nil and false for non-existent key, got %v, %v", nonExistent, ok3)
	}
}

func TestFlatAttributesSetFast(t *testing.T) {
	attrs := NewFlatAttributes()

	attrs.SetFast("key1", "value1")
	attrs.SetFast("key2", 42)

	value1, ok1 := attrs.Get([]string{"key1"})
	if !ok1 || value1 != "value1" {
		t.Errorf("Expected 'value1', got %v, %v", value1, ok1)
	}

	value2, ok2 := attrs.Get([]string{"key2"})
	if !ok2 || value2 != 42 {
		t.Errorf("Expected 42, got %v, %v", value2, ok2)
	}
}

func TestFlatAttributesSetByDotNotation(t *testing.T) {
	attrs := NewFlatAttributes()

	attrs.SetByDotNotation("user.profile.name", "John Doe")
	attrs.SetByDotNotation("user.profile.age", 30)
	attrs.SetByDotNotation("system.version", "1.0")

	name, ok1 := attrs.Get([]string{"user", "profile", "name"})
	if !ok1 || name != "John Doe" {
		t.Errorf("Expected 'John Doe', got %v, %v", name, ok1)
	}

	age, ok2 := attrs.Get([]string{"user", "profile", "age"})
	if !ok2 || age != 30 {
		t.Errorf("Expected 30, got %v, %v", age, ok2)
	}

	version, ok3 := attrs.Get([]string{"system", "version"})
	if !ok3 || version != "1.0" {
		t.Errorf("Expected '1.0', got %v, %v", version, ok3)
	}
}

func TestFlatAttributesIsEmpty(t *testing.T) {
	attrs := NewFlatAttributes()

	if !attrs.IsEmpty() {
		t.Error("New FlatAttributes should be empty")
	}

	attrs.SetFast("key", "value")

	if attrs.IsEmpty() {
		t.Error("FlatAttributes should not be empty after adding value")
	}
}

func TestFlatAttributesClone(t *testing.T) {
	attrs := NewFlatAttributes()
	attrs.Set([]string{"key1"}, "value1")
	attrs.Set([]string{"nested", "key2"}, "value2")

	cloned := attrs.Clone()

	// Test that cloned has same values
	value1, ok1 := cloned.Get([]string{"key1"})
	if !ok1 || value1 != "value1" {
		t.Errorf("Cloned attrs should have same value1, got %v, %v", value1, ok1)
	}

	// Test that modifications to original don't affect clone
	attrs.Set([]string{"key1"}, "modified")
	clonedValue1, ok2 := cloned.Get([]string{"key1"})
	if !ok2 || clonedValue1 != "value1" {
		t.Errorf("Clone should be independent, got %v, %v", clonedValue1, ok2)
	}

	// Test that modifications to clone don't affect original
	cloned.Set([]string{"key3"}, "new")
	originalKey3, ok3 := attrs.Get([]string{"key3"})
	if ok3 || originalKey3 != nil {
		t.Errorf("Original should not have new key from clone, got %v, %v", originalKey3, ok3)
	}
}

func TestFlatAttributesMerge(t *testing.T) {
	attrs1 := NewFlatAttributes()
	attrs1.Set([]string{"key1"}, "value1")
	attrs1.Set([]string{"nested", "key2"}, "value2")

	attrs2 := NewFlatAttributes()
	attrs2.Set([]string{"key3"}, "value3")
	attrs2.Set([]string{"nested", "key4"}, "value4")

	attrs1.Merge(attrs2)

	// Check that attrs1 now has all values
	if val, ok := attrs1.Get([]string{"key1"}); !ok || val != "value1" {
		t.Error("Merge should preserve original values")
	}
	if val, ok := attrs1.Get([]string{"key3"}); !ok || val != "value3" {
		t.Error("Merge should add new values")
	}
	if val, ok := attrs1.Get([]string{"nested", "key2"}); !ok || val != "value2" {
		t.Error("Merge should preserve nested original values")
	}
	if val, ok := attrs1.Get([]string{"nested", "key4"}); !ok || val != "value4" {
		t.Error("Merge should add nested new values")
	}
}

func TestFlatAttributesWalk(t *testing.T) {
	attrs := NewFlatAttributes()
	attrs.Set([]string{"key1"}, "value1")
	attrs.Set([]string{"nested", "key2"}, "value2")
	attrs.Set([]string{"deep", "nested", "key3"}, "value3")

	visited := make(map[string]interface{})
	attrs.Walk(func(path []string, value interface{}) {
		key := strings.Join(path, ".")
		visited[key] = value
	})

	expected := map[string]interface{}{
		"key1":             "value1",
		"nested.key2":      "value2",
		"deep.nested.key3": "value3",
	}

	if !reflect.DeepEqual(visited, expected) {
		t.Errorf("Walk visited %v, expected %v", visited, expected)
	}
}

func TestFlatAttributesToNestedMap(t *testing.T) {
	attrs := NewFlatAttributes()
	attrs.Set([]string{"key1"}, "value1")
	attrs.Set([]string{"user", "name"}, "John")
	attrs.Set([]string{"user", "age"}, 30)
	attrs.Set([]string{"user", "profile", "email"}, "john@example.com")

	nested := attrs.ToNestedMap()

	// Check top-level key
	if nested["key1"] != "value1" {
		t.Errorf("Expected key1=value1, got %v", nested["key1"])
	}

	// Check nested user object
	user, ok := nested["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected user to be a map, got %T", nested["user"])
	}

	if user["name"] != "John" {
		t.Errorf("Expected user.name=John, got %v", user["name"])
	}
	if user["age"] != 30 {
		t.Errorf("Expected user.age=30, got %v", user["age"])
	}

	// Check deeply nested profile
	profile, ok := user["profile"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected user.profile to be a map, got %T", user["profile"])
	}

	if profile["email"] != "john@example.com" {
		t.Errorf("Expected user.profile.email=john@example.com, got %v", profile["email"])
	}
}

func TestFlatAttributesExpandStruct(t *testing.T) {
	attrs := NewFlatAttributes()

	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
		Zip    int    `json:"zip"`
	}

	type User struct {
		ID      int     `json:"id"`
		Name    string  `json:"name"`
		Address Address `json:"address"`
		Active  bool    `json:"active"`
	}

	user := User{
		ID:   123,
		Name: "John Doe",
		Address: Address{
			Street: "123 Main St",
			City:   "Anytown",
			Zip:    12345,
		},
		Active: true,
	}

	attrs.ExpandStruct("user", user)

	// Check expanded fields
	tests := []struct {
		path     []string
		expected interface{}
	}{
		{[]string{"user", "id"}, 123},
		{[]string{"user", "name"}, "John Doe"},
		{[]string{"user", "address", "street"}, "123 Main St"},
		{[]string{"user", "address", "city"}, "Anytown"},
		{[]string{"user", "address", "zip"}, 12345},
		{[]string{"user", "active"}, true},
	}

	for _, test := range tests {
		value, ok := attrs.Get(test.path)
		if !ok || value != test.expected {
			t.Errorf("Expected %v at path %v, got %v, %v", test.expected, test.path, value, ok)
		}
	}
}

func TestFlatAttributesExpandStructWithPointer(t *testing.T) {
	attrs := NewFlatAttributes()

	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	user := &User{ID: 456, Name: "Jane Doe"}

	attrs.ExpandStruct("user", user)

	// Check expanded fields from pointer
	if val, ok := attrs.Get([]string{"user", "id"}); !ok || val != 456 {
		t.Errorf("Expected user.id=456, got %v, %v", val, ok)
	}
	if val, ok := attrs.Get([]string{"user", "name"}); !ok || val != "Jane Doe" {
		t.Errorf("Expected user.name=Jane Doe, got %v, %v", val, ok)
	}
}

func TestFlatAttributesMarshalJSON(t *testing.T) {
	attrs := NewFlatAttributes()
	attrs.Set([]string{"key1"}, "value1")
	attrs.Set([]string{"user", "name"}, "John")
	attrs.Set([]string{"user", "age"}, 30)

	data, err := attrs.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	jsonStr := string(data)

	// Check that all values are present in JSON
	if !strings.Contains(jsonStr, `"key1":"value1"`) {
		t.Errorf("Expected key1 in JSON: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"user.name":"John"`) {
		t.Errorf("Expected user.name in JSON: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"user.age":30`) {
		t.Errorf("Expected user.age in JSON: %s", jsonStr)
	}
}

func TestFlatAttributesMaskingFunctionality(t *testing.T) {
	attrs := NewFlatAttributes()

	// Test maskValue function with different mask tags
	tests := []struct {
		input    interface{}
		maskTag  string
		expected string
	}{
		{"password123", "mask", "***********"},
		{"email@test.com", "mask[5]", "email*********"},
		{"short", "mask[10]", "short"}, // unmask count >= length
		{12345, "mask[2]", "12***"},
		{"", "mask", ""},
		{"test", "mask[0]", "****"},
		{"invalid", "invalid_tag", "invalid"},
		{"test", "", "test"},
	}

	for i, test := range tests {
		result := attrs.maskValue(test.input, test.maskTag)
		if result != test.expected {
			t.Errorf("Test %d: maskValue(%v, %s) = %v, expected %v",
				i, test.input, test.maskTag, result, test.expected)
		}
	}
}

func TestFlatAttributesStructWithMaskTags(t *testing.T) {
	attrs := NewFlatAttributes()

	type UserCredentials struct {
		Username string `json:"username"`
		Password string `json:"password" sawmill:"mask"`
		Email    string `json:"email" sawmill:"mask[4]"`
		APIKey   string `json:"api_key" sawmill:"mask[8]"`
	}

	creds := UserCredentials{
		Username: "johndoe",
		Password: "secret123",
		Email:    "john@example.com",
		APIKey:   "abc123def456ghi789",
	}

	attrs.ExpandStruct("creds", creds)

	// Check that password is masked
	password, ok1 := attrs.Get([]string{"creds", "password"})
	if !ok1 || password != "*********" {
		t.Errorf("Expected masked password, got %v, %v", password, ok1)
	}

	// Check that email shows first 4 characters
	email, ok2 := attrs.Get([]string{"creds", "email"})
	if !ok2 || email != "john************" {
		t.Errorf("Expected email with first 4 chars, got %v, %v", email, ok2)
	}

	// Check that API key shows first 8 characters
	apiKey, ok3 := attrs.Get([]string{"creds", "apikey"})
	if !ok3 || apiKey != "abc123de**********" {
		t.Errorf("Expected API key with first 8 chars, got %v, %v", apiKey, ok3)
	}

	// Check that username is not masked
	username, ok4 := attrs.Get([]string{"creds", "username"})
	if !ok4 || username != "johndoe" {
		t.Errorf("Expected unmasked username, got %v, %v", username, ok4)
	}
}

func TestFlatAttributesEdgeCases(t *testing.T) {
	attrs := NewFlatAttributes()

	// Test with nil value
	attrs.Set([]string{"nil_key"}, nil)
	nilValue, ok1 := attrs.Get([]string{"nil_key"})
	if !ok1 || nilValue != nil {
		t.Errorf("Expected nil value, got %v, %v", nilValue, ok1)
	}

	// Test with empty path - should be ignored by Set method
	attrs.Set([]string{}, "empty_path")
	emptyPathValue, ok2 := attrs.Get([]string{})
	if ok2 || emptyPathValue != nil {
		t.Errorf("Expected empty path to be ignored, got %v, %v", emptyPathValue, ok2)
	}

	// Test with path containing empty strings
	attrs.Set([]string{"", "empty", ""}, "value")
	emptyStringValue, ok3 := attrs.Get([]string{"", "empty", ""})
	if !ok3 || emptyStringValue != "value" {
		t.Errorf("Expected value with empty string path, got %v, %v", emptyStringValue, ok3)
	}
}

func TestFlatAttributesConcurrentAccess(t *testing.T) {
	attrs := NewFlatAttributes()

	// Test concurrent writes and reads
	done := make(chan bool, 2)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			attrs.SetFast("key", i)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			attrs.Get([]string{"key"})
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Should not panic or race
}

func TestFlatAttributesLargeDataset(t *testing.T) {
	attrs := NewFlatAttributes()

	// Add many attributes
	for i := 0; i < 1000; i++ {
		key := []string{"batch", "item", fmt.Sprintf("item_%d", i)}
		attrs.Set(key, i)
	}

	// Verify some random values
	if val, ok := attrs.Get([]string{"batch", "item", "item_0"}); !ok || val != 0 {
		t.Error("Expected first item to be 0")
	}

	// Test walk performance with large dataset
	count := 0
	attrs.Walk(func(path []string, value interface{}) {
		count++
	})

	if count != 1000 {
		t.Errorf("Expected to walk 1000 items, got %d", count)
	}
}

func TestFlatAttributesSpecialCharacters(t *testing.T) {
	attrs := NewFlatAttributes()

	// Test with special characters in keys and values
	attrs.Set([]string{"key with spaces"}, "value with spaces")
	attrs.Set([]string{"key.with.dots"}, "value.with.dots")
	attrs.Set([]string{"key/with/slashes"}, "value/with/slashes")
	attrs.Set([]string{"key\"with\"quotes"}, "value\"with\"quotes")

	// Verify all values are preserved
	if val, ok := attrs.Get([]string{"key with spaces"}); !ok || val != "value with spaces" {
		t.Error("Failed to handle spaces in key/value")
	}
	if val, ok := attrs.Get([]string{"key.with.dots"}); !ok || val != "value.with.dots" {
		t.Error("Failed to handle dots in key/value")
	}
	if val, ok := attrs.Get([]string{"key/with/slashes"}); !ok || val != "value/with/slashes" {
		t.Error("Failed to handle slashes in key/value")
	}
	if val, ok := attrs.Get([]string{"key\"with\"quotes"}); !ok || val != "value\"with\"quotes" {
		t.Error("Failed to handle quotes in key/value")
	}
}