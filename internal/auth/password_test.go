package auth_test

import (
	"testing"

	"github.com/rebyul/chirpy/internal/auth"
)

func TestHashPassword(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := auth.HashPassword(password1)
	hash2, _ := auth.HashPassword(password2)
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password1,
			hash:     hash1,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "wrongPassword",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Password doesn't match different hash",
			password: password1,
			hash:     hash2,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password1,
			hash:     "invalidhash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			failedPwCheck := auth.CheckPasswordHash(tt.password, tt.hash)
			if (failedPwCheck != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() err = %v, wantErr %v", failedPwCheck, tt.wantErr)
			}
		})
	}
}
