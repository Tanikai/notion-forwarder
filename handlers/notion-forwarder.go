package handlers

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"notion-forwarder/dependencies"
)

type NotionForwardHandler struct {
	notionClient *dependencies.NotionForwardingClient
}

func (h NotionForwardHandler) ForwardItem(w http.ResponseWriter, r *http.Request) {
	databaseId := chi.URLParam(r, "databaseId")
	itemId := chi.URLParam(r, "itemId")
	slog.Debug("ForwardItem", "databaseId", databaseId, "itemId", itemId)

	forwardings, err := h.notionClient.GetForwarding(databaseId, itemId)
	if err != nil {
		if errors.Is(err, dependencies.ErrDatabaseNotFound) {
			http.Error(w, "Database name not found in configuration", http.StatusNotFound)
		} else if errors.Is(err, dependencies.ErrItemNotFound) {
			http.Error(w, "Item ID not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if len(forwardings) == 0 { // method shouldnt return a length of 0 but just in case
		http.Error(w, "No forwarding found", http.StatusNotFound)
		return
	}
	if len(forwardings) == 1 {
		http.Redirect(w, r, forwardings[0], http.StatusFound)
		return
	}
	if len(forwardings) > 1 {
		jsonData, err := json.Marshal(forwardings)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMultipleChoices)
		_, err = w.Write(jsonData)
		return
	}
}

func NotionForwarderRoutes(notionForwarder *dependencies.NotionForwardingClient) chi.Router {
	r := chi.NewRouter()
	n := NotionForwardHandler{notionClient: notionForwarder}
	r.Get("/{databaseId}/{itemId}", n.ForwardItem)
	return r
}
