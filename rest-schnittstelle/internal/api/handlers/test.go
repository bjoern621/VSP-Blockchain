package handlers

import "net/http"

func TestHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Test successful"))
	if err != nil {
		return
	}
}
