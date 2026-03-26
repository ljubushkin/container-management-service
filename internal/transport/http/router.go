package httptransport

import "net/http"

func NewRouter(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/containers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.CreateContainer(w, r)
		case http.MethodGet:
			h.GetByID(w, r)
		}
	})

	mux.HandleFunc("/containers/assign", h.AssignWarehouse)

	return mux
}
