package ft

import (
	"encoding/json"
)

func ConvertJSON(n interface{}) []byte {
   m, _ := json.Marshal(n)
   return m
}