package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSubject(t *testing.T) {
	tests := []struct {
		name            string
		subject         string
		allowedSubjects []string
		wantErr         bool
	}{
		{
			name:    "matching subject",
			subject: "system:serviceaccount:test:collector",
			allowedSubjects: []string{
				"system:serviceaccount:test:collector",
			},
			wantErr: false,
		},
		{
			name:    "matching subject in list",
			subject: "system:serviceaccount:test:collector",
			allowedSubjects: []string{
				"system:serviceaccount:test:other",
				"system:serviceaccount:test:collector",
			},
			wantErr: false,
		},
		{
			name:    "non-matching subject",
			subject: "system:serviceaccount:test:unauthorized",
			allowedSubjects: []string{
				"system:serviceaccount:test:collector",
			},
			wantErr: true,
		},
		{
			name:            "empty subject",
			subject:         "",
			allowedSubjects: []string{"system:serviceaccount:test:collector"},
			wantErr:         true,
		},
		{
			name:            "empty allowed list matches nothing",
			subject:         "system:serviceaccount:test:collector",
			allowedSubjects: []string{},
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSubject(tt.subject, tt.allowedSubjects)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
