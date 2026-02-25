package handler

import (
	"encoding/json"
	"net/http"
)

func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{
		"message": "Hello World",
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonResp)
}
