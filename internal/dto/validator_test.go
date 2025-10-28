package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorAdd(t *testing.T) {
	v := NewValidator()

	v.Add("field1", "error message 1")
	v.Add("field2", "error message 2")

	assert.False(t, v.Valid())
	assert.Len(t, v.Errors(), 2)
	assert.Equal(t, "field1", v.Errors()[0].Field)
	assert.Equal(t, "error message 1", v.Errors()[0].Message)
	assert.Equal(t, "field2", v.Errors()[1].Field)
	assert.Equal(t, "error message 2", v.Errors()[1].Message)
}

func TestValidatorValid_NoErrors(t *testing.T) {
	v := NewValidator()

	assert.True(t, v.Valid())
	assert.Empty(t, v.Errors())
}

func TestValidatorValid_WithErrors(t *testing.T) {
	v := NewValidator()
	v.Add("field", "error")

	assert.False(t, v.Valid())
	assert.Len(t, v.Errors(), 1)
}

func TestValidatorErrors(t *testing.T) {
	v := NewValidator()
	v.Add("field1", "error1")
	v.Add("field2", "error2")
	v.Add("field3", "error3")

	errors := v.Errors()

	assert.Len(t, errors, 3)
	assert.Equal(t, "field1", errors[0].Field)
	assert.Equal(t, "field2", errors[1].Field)
	assert.Equal(t, "field3", errors[2].Field)
}

func TestNewValidator(t *testing.T) {
	v := NewValidator()

	assert.NotNil(t, v)
	assert.True(t, v.Valid())
	assert.Empty(t, v.Errors())
}

func TestValidatorMultipleValidations(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*Validator)
		wantValid  bool
		errorCount int
	}{
		{
			name: "no errors",
			setup: func(v *Validator) {
			},
			wantValid:  true,
			errorCount: 0,
		},
		{
			name: "one error",
			setup: func(v *Validator) {
				v.Add("field1", "error1")
			},
			wantValid:  false,
			errorCount: 1,
		},
		{
			name: "multiple errors",
			setup: func(v *Validator) {
				v.Add("field1", "error1")
				v.Add("field2", "error2")
				v.Add("field3", "error3")
			},
			wantValid:  false,
			errorCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			tt.setup(v)

			assert.Equal(t, tt.wantValid, v.Valid())
			assert.Len(t, v.Errors(), tt.errorCount)
		})
	}
}
