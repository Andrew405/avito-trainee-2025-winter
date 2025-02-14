package coin

import (
	"encoding/json"
	"net/http"
)

type SendCoinRequest struct {
	ToUser string `json:"ToUser"`
	Amount int    `json:"amount"`
}

func MakeSendCoinHandler(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SendCoinRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value("userID").(int)
		err := s.SendCoin(r.Context(), userID, req.ToUser, req.Amount)
		if err != nil {
			switch err {
			case ErrInsufficientFunds:
				http.Error(w, "Not enough coins", http.StatusBadRequest)
			case ErrSameUser:
				http.Error(w, "Unable to send coins to yourself", http.StatusBadRequest)
			default:
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Coins sent successfully"))
	}
}
