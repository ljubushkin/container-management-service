package httptransport

import "net/http"

func NewRouter(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()

	// /containers → list + create
	mux.HandleFunc("/containers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListContainers(w, r)
		case http.MethodPost:
			h.CreateContainer(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// batch create
	mux.HandleFunc("/containers/batch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CreateContainerBatch(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// get by id (оставляем через query, пока без chi/gorilla)
	mux.HandleFunc("/containers/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetByID(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// assign warehouse
	mux.HandleFunc("/containers/assign", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.AssignWarehouse(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	return mux
}
