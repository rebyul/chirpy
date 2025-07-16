package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rebyul/chirpy/internal/auth"
)

func TestValidateJWT(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		userID         uuid.UUID
		expiresIn      time.Duration
		tokenSecret    string
		validateSecret string
		wantCreateErr  bool
		wantErr        bool
	}{
		{
			name:           "valid token",
			userID:         uuid.New(),
			expiresIn:      time.Duration(100 * time.Second),
			tokenSecret:    "supersecretstringhorse",
			validateSecret: "supersecretstringhorse",
			wantCreateErr:  false,
			wantErr:        false,
		}, {
			name:           "invalid secret",
			userID:         uuid.New(),
			expiresIn:      time.Duration(100 * time.Second),
			tokenSecret:    "supersecretstringhorse",
			validateSecret: "incorrectsecret",
			wantCreateErr:  false,
			wantErr:        true,
		}, {
			name:           "expired token",
			userID:         uuid.New(),
			expiresIn:      time.Duration(-100 * time.Second),
			tokenSecret:    "supersecretstringhorse",
			validateSecret: "supersecretstringhorse",
			wantCreateErr:  false,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := auth.MakeJWT(tt.userID, tt.tokenSecret, tt.expiresIn)

			if (err != nil) != tt.wantCreateErr {
				t.Errorf("Unexpected MakeJWT() failed got: %v, want: %v error: %v", err != nil, tt.wantCreateErr, err)
			}

			got, gotErr := auth.ValidateJWT(token, tt.validateSecret)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ValidateJWT() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ValidateJWT() succeeded unexpectedly")
			}

			if tt.userID != got {
				t.Errorf("ValidateJWT() = %v, want %v", got, tt.userID)
			}
		})
	}
}
