package api

import (
	"net/http"
)

func writeErrCode(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, envelope{OK: false, Error: code})
}
