package handler_test

import (
	"encoding/json"
	"errors"
	"io"
	"testing"

	"course_server/internal/handler"

	"github.com/go-playground/validator/v10"
)

func TestJSONBindErrorMessage(t *testing.T) {
	t.Parallel()

	var requiredErr error
	validate := validator.New()
	if err := validate.Struct(struct {
		TargetType string `validate:"required"`
	}{}); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) && len(ve) > 0 {
			requiredErr = ve
		}
	}
	if requiredErr == nil {
		t.Fatal("expected validation error")
	}

	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "empty body", err: io.EOF, want: "request body is required"},
		{name: "unexpected eof", err: io.ErrUnexpectedEOF, want: "request body is required"},
		{name: "syntax error", err: &json.SyntaxError{}, want: "invalid JSON body"},
		{
			name: "type error",
			err:  &json.UnmarshalTypeError{Field: "favoriteRequest.TargetID"},
			want: "targetID has invalid type",
		},
		{name: "required field", err: requiredErr, want: "targetType is required"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := handler.JSONBindErrorMessage(tt.err); got != tt.want {
				t.Fatalf("JSONBindErrorMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}
