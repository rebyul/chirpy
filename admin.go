package main

import (
	"net/http"

	"github.com/rebyul/chirpy/internal/responses"
)

type resetHandler struct {
	cfg *apiConfig
}

func (r resetHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.cfg.platform != "dev" {
		responses.SendJsonErrorResponse(w, http.StatusForbidden, "forbidden", nil)
		return
	}

	deletedIds, err := r.cfg.queries.DeleteUsers(req.Context())
	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "db failed to delete users", err)
		return
	}

	type postResetResponse struct {
		Ids []string `json:"ids"`
	}
	respIds := make([]string, 0, len(deletedIds))
	for _, id := range deletedIds {
		respIds = append(respIds, id.String())
	}
	res := postResetResponse{respIds}

	responses.SendJsonResponse(w, http.StatusOK, res)
	return
}
