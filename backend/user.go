package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type devicePreferences struct {
	DeviceID          string  `json:"device_id"`
	SortOrder         *int    `json:"sort_order"`
	Archived          bool    `json:"archived"`
	Hidden            bool    `json:"hidden"`
	CustomDisplayName *string `json:"custom_display_name"`
	IconStoragePath   *string `json:"icon_storage_path"`
}

func setUserPreferences(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.Header().Set("Allow", http.MethodPut)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var prefs devicePreferences
		if err := decoder.Decode(&prefs); err != nil {
			http.Error(w, "invalid preferences JSON", http.StatusBadRequest)
			return
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			http.Error(w, "expected one JSON object", http.StatusBadRequest)
			return
		}

		prefs.DeviceID = strings.TrimSpace(prefs.DeviceID)
		if prefs.DeviceID == "" || (prefs.SortOrder != nil && *prefs.SortOrder < 0) {
			http.Error(w, "device_id is required and sort_order must be nonnegative", http.StatusBadRequest)
			return
		}

		_, err := db.Exec(r.Context(), `
			INSERT INTO device_preferences
				(device_id, sort_order, archived, hidden, custom_display_name, icon_storage_path)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (device_id) DO UPDATE SET
				sort_order = EXCLUDED.sort_order,
				archived = EXCLUDED.archived,
				hidden = EXCLUDED.hidden,
				custom_display_name = EXCLUDED.custom_display_name,
				icon_storage_path = EXCLUDED.icon_storage_path,
				updated_at = NOW()
		`, prefs.DeviceID, prefs.SortOrder, prefs.Archived, prefs.Hidden,
			prefs.CustomDisplayName, prefs.IconStoragePath)
		if err != nil {
			log.Printf("saving device preferences: %v", err)
			http.Error(w, "could not save preferences", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
