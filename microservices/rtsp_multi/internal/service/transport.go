package service

import (
	"fmt"
	"net/http"

	"github.com/Impisigmatus/service_core/log"
	"github.com/Impisigmatus/service_core/utils"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

type Transport struct {
	apps map[string]*Application
}

func NewTransport(apps map[string]*Application) *Transport {
	return &Transport{
		apps: apps,
	}
}

func (h *Transport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log, ok := r.Context().Value(log.CtxKey).(zerolog.Logger)
	if !ok {
		utils.WriteString(zerolog.Logger{}, w, http.StatusInternalServerError, fmt.Errorf("Invalid logger"), "Невалидный логгер")
		return
	}

	cameraName := chi.URLParam(r, "camera_name")
	log.Info().Msgf("Stream request for camera: %s", cameraName)
	if app, ok := h.apps[cameraName]; ok {
		app.StreamHandler(log, w, r)
	} else {
		utils.WriteString(log, w, http.StatusNotFound, nil, "Camera not found")
	}
}
