package handlers

import (
	"fmt"
	"net/url"
	"regexp"
	"unicode/utf8"
)

type Validator struct {
	location string
	field    string
	value    *string
}

func (rv *Validator) Required() *CustomError {
	if rv.value == nil {
		// todo no need value here
		return &CustomError{Location: rv.location, Param: rv.field, Msg: "is required"}
	}

	return nil
}

func (rv *Validator) Empty() *CustomError {
	if utf8.RuneCountInString(*rv.value) == 0 {
		return &CustomError{Location: rv.location, Param: rv.field, Value: *rv.value,
			Msg: "cannot be blank"}
	}

	return nil
}

func (rv *Validator) MinLength(min int) *CustomError {
	lenStr := utf8.RuneCountInString(*rv.value)
	if lenStr < min {
		return &CustomError{Location: rv.location, Param: rv.field, Value: *rv.value,
			Msg: fmt.Sprintf("must be at least %d characters long", min)}
	}

	return nil
}

func (rv *Validator) MaxLength(max int) *CustomError {
	lenStr := utf8.RuneCountInString(*rv.value)
	if lenStr > max {
		return &CustomError{Location: rv.location, Param: rv.field, Value: *rv.value,
			Msg: fmt.Sprintf("must be at most %d characters long", max)}
	}

	return nil
}

func (rv *Validator) Custom(validate func(string) bool, msg string) *CustomError {
	if !validate(*rv.value) {
		return &CustomError{Location: rv.location, Param: rv.field, Value: *rv.value, Msg: msg}
	}

	return nil
}

func (rv *Validator) Matches(regexpStr string) *CustomError {
	// todo cache for compiled regexps
	r, _ := regexp.Compile(regexpStr)
	if !r.MatchString(*rv.value) {
		return &CustomError{Location: rv.location, Param: rv.field, Value: *rv.value,
			Msg: "contains invalid characters"}
	}

	return nil
}

func (rv *Validator) URL() *CustomError {
	_, err := url.ParseRequestURI(*rv.value)
	if err != nil {
		return &CustomError{Location: rv.location, Param: rv.field, Value: *rv.value,
			Msg: "is invalid"}
	}

	return nil
}

func mergeErrors(validations ...*CustomError) []*CustomError {
	result := make([]*CustomError, 0, 2)

	for _, err := range validations {
		if err == nil {
			continue
		}

		result = append(result, err)
	}

	return result
}
