package api

import (
	"context"
	"net/http"
	"time"
)

func (svc *Service) VersionHandler(w http.ResponseWriter, r *http.Request) {
	if err := svc.MarshalJSONResponse(w, http.StatusOK, map[string]string{
		"version":    svc.cfg.Version,
		"build_date": svc.cfg.BuildDate,
	}); err != nil {
		return
	}
}

func (svc *Service) HealthHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	logger, err := LoggerFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	start := time.Now().UTC()

	ctx, cancel := context.WithTimeout(ctx, svc.cfg.RequestTimeout)
	defer cancel()

	if err := svc.mongodb.Ping(ctx, nil); err != nil {
		logger.Errorf("failed to ping mongodb through api client: %s", err.Error())
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.MarshalJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":     "ok",
		"time_taken": time.Since(start).String(),
		"version":    svc.cfg.Version,
		"build_date": svc.cfg.BuildDate,
	}); err != nil {
		return
	}
}
