package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

type upstreamDevicesResponse struct {
	ResultList []map[string]any `json:"result_list"`
}

func getDevices(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		request, err := http.NewRequestWithContext(r.Context(), http.MethodGet,
			fmt.Sprintf("%s?latest_point=true&api-key=%s", os.Getenv("ONE_STEP_GPS_DEVICE_URL"), os.Getenv("ONE_STEP_GPS_DEVICE_KEY")), nil)
		if err != nil {
			http.Error(w, "could not create device request", http.StatusInternalServerError)
			return
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			http.Error(w, "could not fetch devices", http.StatusBadGateway)
			return
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			log.Printf("device service status: %v", response.StatusCode)
			http.Error(w, "device service returned an error", http.StatusBadGateway)
			return
		}
		var payload upstreamDevicesResponse
		if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
			log.Printf("decoding device response: %v", err)
			http.Error(w, "could not read device response", http.StatusBadGateway)
			return
		}
		preferences, err := loadPreferences(r, db)
		if err != nil {
			log.Printf("loading device preferences: %v", err)
			http.Error(w, "could not load device preferences", http.StatusInternalServerError)
			return
		}
		for _, device := range payload.ResultList {
			id, _ := device["device_id"].(string)
			prefs, ok := preferences[id]
			if !ok {
				continue
			}
			device["sort_order"] = prefs.SortOrder
			device["archived"] = prefs.Archived
			device["hidden"] = prefs.Hidden
			device["custom_display_name"] = prefs.CustomDisplayName
			device["icon_storage_path"] = prefs.IconStoragePath
		}
		sort.SliceStable(payload.ResultList, func(i, j int) bool {
			leftID, _ := payload.ResultList[i]["device_id"].(string)
			rightID, _ := payload.ResultList[j]["device_id"].(string)
			left, leftOK := preferences[leftID]
			right, rightOK := preferences[rightID]
			if leftOK && left.SortOrder != nil && rightOK && right.SortOrder != nil {
				return *left.SortOrder < *right.SortOrder
			}
			return leftOK && left.SortOrder != nil
		})
		writeJSON(w, http.StatusOK, payload)
	}
}

func loadPreferences(r *http.Request, db *pgxpool.Pool) (map[string]devicePreferences, error) {
	rows, err := db.Query(r.Context(), `
		SELECT device_id, sort_order, archived, hidden, custom_display_name, icon_storage_path
		FROM device_preferences
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]devicePreferences)
	for rows.Next() {
		var prefs devicePreferences
		if err := rows.Scan(&prefs.DeviceID, &prefs.SortOrder, &prefs.Archived, &prefs.Hidden,
			&prefs.CustomDisplayName, &prefs.IconStoragePath); err != nil {
			return nil, err
		}
		result[prefs.DeviceID] = prefs
	}
	return result, rows.Err()
}
