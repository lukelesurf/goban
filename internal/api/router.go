package api

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(boards *BoardHandler, columns *ColumnHandler, tasks *TaskHandler, attachments *AttachmentHandler, ui embed.FS) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/boards", boards.List)
		r.Post("/boards", boards.Create)
		r.Get("/boards/{id}", boards.Get)
		r.Put("/boards/{id}", boards.Update)
		r.Delete("/boards/{id}", boards.Delete)
		r.Put("/boards/{id}/default", boards.SetDefault)
		r.Delete("/boards/{id}/default", boards.ClearDefault)

		r.Get("/boards/{id}/columns", columns.List)
		r.Post("/boards/{id}/columns", columns.Create)
		r.Put("/columns/{id}", columns.Update)
		r.Delete("/columns/{id}", columns.Delete)

		r.Get("/boards/{id}/tasks", tasks.List)
		r.Post("/boards/{id}/tasks", tasks.Create)
		r.Get("/tasks/{id}", tasks.Get)
		r.Put("/tasks/{id}", tasks.Update)
		r.Delete("/tasks/{id}", tasks.Delete)
		r.Post("/tasks/{id}/move", tasks.Move)
		r.Get("/tasks/{id}/comments", tasks.ListComments)
		r.Post("/tasks/{id}/comments", tasks.AddComment)
		r.Get("/tasks/{id}/attachments", attachments.List)
		r.Post("/tasks/{id}/attachments", attachments.Upload)
		r.Get("/attachments/{id}/download", attachments.Download)
		r.Delete("/attachments/{id}", attachments.Delete)
	})

	// Serve embedded UI at root
	uiFS, _ := fs.Sub(ui, "dist")
	r.Handle("/*", http.FileServer(http.FS(uiFS)))

	return r
}
