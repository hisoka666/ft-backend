package backend

import (
	"encoding/json"
	"net/http"
)

func ConvertJSON(n interface{}) []byte {
	m, _ := json.Marshal(n)
	return m
}

func ServerResponseTemplate(w http.ResponseWriter, n interface{}) {
	m := ConvertJSON(n)
	json.NewEncoder(w).Encode(m)
}
