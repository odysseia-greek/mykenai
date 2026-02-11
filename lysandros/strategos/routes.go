package strategos

import (
	"io/fs"
	"net/http"
)

// InitRoutes to start up a mux router and return the routes
func InitRoutes(lysandrosHandler *LysandrosHandler) *http.ServeMux {
	mux := http.NewServeMux()
	staticSub := mustSub(staticFS, "static")
	mux.Handle("/lysandros/static/", http.StripPrefix("/lysandros/static/", http.FileServer(http.FS(staticSub))))
	mux.HandleFunc("/lysandros/", lysandrosHandler.handleRoot)
	mux.HandleFunc("/lysandros", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/lysandros/", http.StatusFound)
	})

	return mux
}

func mustSub(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
