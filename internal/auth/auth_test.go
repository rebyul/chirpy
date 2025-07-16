package auth_test

import (
	"net/http"
	"testing"

	"github.com/rebyul/chirpy/internal/auth"
)

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		headers http.Header
		want    string
		wantErr bool
	}{
		{
			name:    "valid header with bearer token",
			headers: http.Header{"Authorization": {"Bearer asdfasdf"}},
			want:    "asdfasdf",
			wantErr: false,
		}, {
			name:    "no authorization header",
			headers: http.Header{},
			want:    "asdfasdf",
			wantErr: true,
		}, {
			name:    "lowercase bearer text",
			headers: http.Header{"Authorization": {"bearer asdfasdf"}},
			want:    "asdfasdf",
			wantErr: false,
		}, {
			name:    "no bearer text",
			headers: http.Header{"Authorization": {"asdfasdf"}},
			want:    "",
			wantErr: true,
		}, {
			name:    "partial bearer text",
			headers: http.Header{"Authorization": {"bear asdfasdf"}},
			want:    "",
			wantErr: true,
		}, {
			name:    "bearer in token",
			headers: http.Header{"Authorization": {"abearer sdfasdf"}},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := auth.GetBearerToken(tt.headers)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetBearerToken() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetBearerToken() succeeded unexpectedly")
			}

			if got != tt.want {
				t.Errorf("GetBearerToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
