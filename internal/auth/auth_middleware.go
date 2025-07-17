package auth

import (
	"log"
	"net/http"
	"time"

	"github.com/rebyul/chirpy/internal/responses"
)

type JwtAuthenticationMiddleware struct {
	Tokensecret string
}

func (j *JwtAuthenticationMiddleware) MiddlewareJwtAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("starting request %v", r.URL.Path)
		log.Printf("[Auth middleware start]: %v\n", r.Header.Get("Date"))
		headerToken, err := GetBearerToken(r.Header)

		if err != nil {
			responses.SendJsonErrorResponse(w, http.StatusUnauthorized, "invalid jwt token", err)
			log.Printf("[Auth middleware end]: %v\n", time.Now())
			return
		}

		log.Printf("token: %v", headerToken)
		_, err = ValidateJWT(headerToken, j.Tokensecret)

		if err != nil {
			responses.SendJsonErrorResponse(w, http.StatusUnauthorized, err.Error(), err)
			log.Printf("[Auth middleware end]: %v\n", time.Now())
			return
		}

		log.Printf("[Auth middleware end]: %v\n", time.Now())
		next.ServeHTTP(w, r)
	})
}
