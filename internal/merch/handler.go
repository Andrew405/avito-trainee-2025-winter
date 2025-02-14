package merch

import (
	"errors"
	"net/http"
)

func MakeBuyHandler(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := r.PathValue("item")
		if item == "" {
			http.Error(w, "Item not specified", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value("userID").(int)
		err := s.BuyItem(r.Context(), userID, item)
		if err != nil {
			switch {
			case errors.Is(err, ErrInsufficientCoins):
				http.Error(w, "Not enough coins", http.StatusBadRequest)
			case errors.Is(err, ErrItemNotFound):
				http.Error(w, "Item not found", http.StatusBadRequest)
			default:
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Item purchased successfully"))
	}
}
