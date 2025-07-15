package auth_test

import (
	"testing"

	"github.com/rebyul/chirpy/internal/auth"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		password string
		wantErr  bool
	}{
		{
			name:     "Hash and check password hash",
			password: "asdfasdf",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := auth.HashPassword(tt.password)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("HashPassword() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("HashPassword() succeeded unexpectedly")
			}
			failedPwCheck := auth.CheckPasswordHash(tt.password, got)
			if failedPwCheck != nil {
				t.Errorf("HashPassword() = %v, failed CheckPasswordHash(): %v", got, failedPwCheck)
			}
		})
	}
}
