package handlers

import "testing"

var str = "test_value"

func TestValidator(t *testing.T) {
	validator := &Validator{
		location: "test_location",
		field:    "test_field",
		value:    &str,
	}

	validator.Required()
	validator.Empty()
	validator.Matches("someregexp")
	validator.MaxLength(10)
	validator.MinLength(20)
	validator.URL()
	validator.Custom(func(string) bool { return true }, "test")
}
