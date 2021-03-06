package proxy

import (
	"encoding/json"
	"io"
	"net/http"
)

// MaximalBodySize specifies maximal supported request body size.
const MaximalBodySize = 1000000 // 1 MB

func readJSON(v interface{}, r io.ReadCloser) error {
	if r == nil {
		return io.EOF
	}
	d := json.NewDecoder(io.LimitReader(r, MaximalBodySize))
	return d.Decode(v)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
