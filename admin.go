package main

import "net/http"

type resetHandler struct {
	cfg *apiConfig
}

func (r resetHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.cfg.platform != "dev" {
		sendJsonErrorResponse(w, http.StatusForbidden, "forbidden", nil)
		return
	}

	deletedIds, err := r.cfg.queries.DeleteUsers(req.Context())
	if err != nil {
		sendJsonErrorResponse(w, http.StatusInternalServerError, "db failed to delete users", err)
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

	sendJsonResponse(w, http.StatusOK, res)
	return
}
