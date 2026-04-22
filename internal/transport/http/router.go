package httptransport

import "net/http"

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	// health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.Health(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

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

	// get by id
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

	// list wharehouse
	mux.HandleFunc("/warehouses", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.ListWarehouses(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// list container types
	mux.HandleFunc("/types", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.ListContainerTypes(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	return LoggingMiddlware(mux)
}
