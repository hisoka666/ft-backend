package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func init() {
	http.HandleFunc("/json", jsonFunc)
}

type Member struct {
	Name, Address string
}

func jsonFunc(w http.ResponseWriter, r *http.Request) {
	members := []Member{}
	members = append(members, Member{"Andy", "Batubulan"})
	members = append(members, Member{"Betty", "Ubud"})
	members = append(members, Member{"Rian", "Celuk"})

	memJ, err := json.Marshal(members)
	if err != nil {
		fmt.Fprintf(w, "Error parsing json: %v", err)
	}

	fmt.Fprintln(w, string(memJ))
}
