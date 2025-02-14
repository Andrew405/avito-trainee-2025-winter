package info

import (
	"encoding/json"
	"net/http"
)

func MakeInfoHandler(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)
		info, err := s.GetInfo(r.Context(), userID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	}
}
