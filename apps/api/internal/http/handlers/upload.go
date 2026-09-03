package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"dpmptsp/api/internal/http/middleware"
	"dpmptsp/api/internal/http/render"
	"dpmptsp/api/internal/infrastructure/storage"
)

type Uploads struct {
	Files *storage.Manager
	Log   *slog.Logger
}

type uploadDTO struct {
	Key         string `json:"key"`
	URL         string `json:"url,omitempty"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

// maxMultipartMemory caps what is buffered in RAM; larger parts spill to a
// temporary file rather than being held whole.
const maxMultipartMemory = 8 << 20

// Create stores an uploaded file and returns the key to reference it by.
//
// Callers send multipart/form-data with a `file` part and an optional `prefix`.
// The stored key is generated; the client's filename is used only to pick an
// extension when the sniffed type is unrecognised.
func (h *Uploads) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxMultipartMemory); err != nil {
		render.BadRequest(w, "expected multipart/form-data with a file part")
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("file")
	if err != nil {
		render.BadRequest(w, "missing file part")
		return
	}
	defer file.Close()

	visibility := storage.Private
	if r.FormValue("visibility") == "public" {
		visibility = storage.Public
	}

	// A public upload defaults to the public disk. The default disk is `local`,
	// which has no base URL, so without this a caller asking for a public file
	// got back a key that nothing can serve.
	name := r.FormValue("disk")
	if name == "" && visibility == storage.Public {
		name = "public"
	}
	disk, err := h.Files.Disk(name)
	if err != nil {
		render.BadRequest(w, err.Error())
		return
	}

	obj, err := storage.Upload(r.Context(), disk, file, header.Filename, storage.UploadRules{
		Prefix:     r.FormValue("prefix"),
		Visibility: visibility,
	})
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrTypeRejected):
			render.JSON(w, http.StatusUnsupportedMediaType, map[string]string{"error": err.Error()})
		case errors.Is(err, storage.ErrTooLarge):
			render.JSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": err.Error()})
		default:
			render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		}
		return
	}

	out := uploadDTO{Key: obj.Key, Size: obj.Size, ContentType: obj.ContentType}
	// A private disk has no URL; the key is enough to fetch it through a
	// handler later.
	if url, err := disk.URL(obj.Key); err == nil {
		out.URL = url
	}
	render.Created(w, out)
}
