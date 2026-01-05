package request

import "net/http"

func Ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Server is running\n"))
	w.WriteHeader(http.StatusOK)
}
