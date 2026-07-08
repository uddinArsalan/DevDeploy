package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/middlewares"
	"github.com/uddinArsalan/devdeploy/internals/services"
	"github.com/uddinArsalan/devdeploy/internals/utils"
)

type LogHandler struct {
	ls *services.LogService
}

func NewLogHandler(ls *services.LogService) *LogHandler {
	return &LogHandler{
		ls: ls,
	}
}

func (lh *LogHandler) StreamLogsHandler(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middlewares.UserFromContext(r.Context())
	if !ok {
		utils.FAIL(w, http.StatusUnauthorized, "authentication required")
		return
	}

	deployID := r.PathValue("deployID")
	if deployID == "" {
		utils.FAIL(w, http.StatusBadRequest, "missing deploy id")
		return
	}

	lastID := r.URL.Query().Get("last_id")
	if lastID == "" {
		lastID = "0"
	}

	deployIDInt, err := strconv.ParseInt(deployID, 10, 64)

	if err != nil {
		utils.FAIL(w, http.StatusBadRequest, "invalid deploy id")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		utils.FAIL(w, http.StatusInternalServerError, "Streaming not Supported")
		return
	}

	ch, err := lh.ls.StreamLogs(r.Context(), userClaim.UserID, lastID, deployIDInt)
	if err != nil {
		utils.FAIL(w, http.StatusInternalServerError, "invalid deploy id")
		return
	}
	ctx := r.Context()

	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return
			}
			if event.Type == domain.Status {
				fmt.Fprintf(w, "event: status\ndata: %s\n\n", event.Data)
			} else {
				fmt.Fprintf(w, "data: %s\n\n", event.Data)
			}

			fmt.Fprintf(w, "event: cursor\ndata: %s\n\n", event.ID)
			flusher.Flush()

			if event.Type == domain.Status && (event.Data == "running" || event.Data == "failed") {
				fmt.Fprintf(w, "event: done\ndata: %s\n\n", event.Data)
				flusher.Flush()
				return
			}

		case <-ctx.Done():
			return
		}
	}

}
