package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "golang.org/x/image/webp"
)

const (
	maxIconBytes      = 5 << 20
	maxIconDimension  = 1024
	multipartOverhead = 1 << 20
)

type devicePreferences struct {
	DeviceID          string  `json:"device_id"`
	SortOrder         *int    `json:"sort_order"`
	Archived          bool    `json:"archived"`
	Hidden            bool    `json:"hidden"`
	CustomDisplayName *string `json:"custom_display_name"`
	IconStoragePath   *string `json:"icon_storage_path"`
}

type deviceOrderRequest struct {
	DeviceIDs []string `json:"device_ids"`
}

func preferencesHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.Header().Set("Allow", http.MethodPut)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var prefs devicePreferences
		if !decodeJSON(w, r, &prefs) {
			return
		}
		prefs.DeviceID = strings.TrimSpace(prefs.DeviceID)
		if prefs.DeviceID == "" || (prefs.SortOrder != nil && *prefs.SortOrder < 0) {
			http.Error(w, "device_id is required and sort_order must be nonnegative", http.StatusBadRequest)
			return
		}
		prefs.CustomDisplayName = cleanOptionalString(prefs.CustomDisplayName)

		_, err := db.Exec(r.Context(), `
			INSERT INTO device_preferences
				(device_id, sort_order, archived, hidden, custom_display_name, icon_storage_path)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (device_id) DO UPDATE SET
				sort_order = EXCLUDED.sort_order,
				archived = EXCLUDED.archived,
				hidden = EXCLUDED.hidden,
				custom_display_name = EXCLUDED.custom_display_name,
				icon_storage_path = COALESCE(EXCLUDED.icon_storage_path, device_preferences.icon_storage_path),
				updated_at = NOW()
		`, prefs.DeviceID, prefs.SortOrder, prefs.Archived, prefs.Hidden,
			prefs.CustomDisplayName, prefs.IconStoragePath)
		if err != nil {
			log.Printf("saving device preferences: %v", err)
			http.Error(w, "could not save preferences", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, prefs)
	}
}

func deviceOrderHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.Header().Set("Allow", http.MethodPut)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var request deviceOrderRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		seen := make(map[string]struct{}, len(request.DeviceIDs))
		for index, id := range request.DeviceIDs {
			id = strings.TrimSpace(id)
			if id == "" {
				http.Error(w, "device_ids cannot contain empty values", http.StatusBadRequest)
				return
			}
			if _, duplicate := seen[id]; duplicate {
				http.Error(w, "device_ids cannot contain duplicates", http.StatusBadRequest)
				return
			}
			seen[id] = struct{}{}
			request.DeviceIDs[index] = id
		}

		tx, err := db.Begin(r.Context())
		if err != nil {
			http.Error(w, "could not update device order", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback(r.Context())
		for index, id := range request.DeviceIDs {
			_, err = tx.Exec(r.Context(), `
				INSERT INTO device_preferences (device_id, sort_order)
				VALUES ($1, $2)
				ON CONFLICT (device_id) DO UPDATE SET sort_order = EXCLUDED.sort_order, updated_at = NOW()
			`, id, index)
			if err != nil {
				http.Error(w, "could not update device order", http.StatusInternalServerError)
				return
			}
		}
		if err = tx.Commit(r.Context()); err != nil {
			http.Error(w, "could not update device order", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func deviceIconHandler(db *pgxpool.Pool, storage iconStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			uploadDeviceIcon(w, r, db, storage)
		case http.MethodDelete:
			removeDeviceIcon(w, r, db, storage)
		default:
			w.Header().Set("Allow", http.MethodPost+", "+http.MethodDelete)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func uploadDeviceIcon(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool, storage iconStorage) {
	r.Body = http.MaxBytesReader(w, r.Body, maxIconBytes+multipartOverhead)
	if err := r.ParseMultipartForm(maxIconBytes); err != nil {
		http.Error(w, "icon must be smaller than 5 MB", http.StatusBadRequest)
		return
	}
	deviceID := strings.TrimSpace(r.FormValue("device_id"))
	if deviceID == "" {
		http.Error(w, "device_id is required", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("icon")
	if err != nil {
		http.Error(w, "icon file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()
	if header.Size > maxIconBytes {
		http.Error(w, "icon must be 5 MB or smaller", http.StatusBadRequest)
		return
	}
	extension, contentType, err := imageMetadata(file, header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	filename, err := randomFilename(extension)
	if err != nil {
		http.Error(w, "could not create icon name", http.StatusInternalServerError)
		return
	}
	storagePath := "device-icons/" + filename
	var oldStoragePath *string
	err = db.QueryRow(r.Context(), `
			SELECT icon_storage_path FROM device_preferences WHERE device_id = $1
		`, deviceID).Scan(&oldStoragePath)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "could not read existing icon", http.StatusInternalServerError)
		return
	}
	if err := storage.Put(r.Context(), storagePath, file, contentType); err != nil {
		log.Printf("uploading device icon: %v", err)
		http.Error(w, "could not upload icon", http.StatusBadGateway)
		return
	}
	_, err = db.Exec(r.Context(), `
			INSERT INTO device_preferences (device_id, icon_storage_path)
			VALUES ($1, $2)
			ON CONFLICT (device_id) DO UPDATE SET icon_storage_path = EXCLUDED.icon_storage_path, updated_at = NOW()
		`, deviceID, storagePath)
	if err != nil {
		_ = storage.Delete(r.Context(), storagePath)
		log.Printf("saving device icon: %v", err)
		http.Error(w, "could not save icon", http.StatusInternalServerError)
		return
	}
	if oldStoragePath != nil && *oldStoragePath != storagePath {
		if err := storage.Delete(r.Context(), *oldStoragePath); err != nil {
			log.Printf("deleting previous device icon: %v", err)
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"icon_storage_path": storagePath,
		"icon_url":          storage.PublicURL(storagePath),
	})
}

func removeDeviceIcon(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool, storage iconStorage) {
	var request struct {
		DeviceID string `json:"device_id"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	request.DeviceID = strings.TrimSpace(request.DeviceID)
	if request.DeviceID == "" {
		http.Error(w, "device_id is required", http.StatusBadRequest)
		return
	}

	var storagePath *string
	err := db.QueryRow(r.Context(), `
		SELECT icon_storage_path
		FROM device_preferences
		WHERE device_id = $1
	`, request.DeviceID).Scan(&storagePath)
	if errors.Is(err, pgx.ErrNoRows) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		log.Printf("clearing device icon: %v", err)
		http.Error(w, "could not remove icon", http.StatusInternalServerError)
		return
	}
	_, err = db.Exec(r.Context(), `
		UPDATE device_preferences
		SET icon_storage_path = NULL, updated_at = NOW()
		WHERE device_id = $1
	`, request.DeviceID)
	if err != nil {
		log.Printf("clearing device icon: %v", err)
		http.Error(w, "could not remove icon", http.StatusInternalServerError)
		return
	}
	if storagePath != nil && strings.HasPrefix(*storagePath, "device-icons/") {
		if err := storage.Delete(r.Context(), *storagePath); err != nil {
			if _, restoreErr := db.Exec(r.Context(), `
				UPDATE device_preferences
				SET icon_storage_path = $2, updated_at = NOW()
				WHERE device_id = $1
			`, request.DeviceID, *storagePath); restoreErr != nil {
				log.Printf("restoring device icon after failed deletion: %v", restoreErr)
			}
			log.Printf("deleting device icon: %v", err)
			http.Error(w, "could not remove icon", http.StatusBadGateway)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		http.Error(w, "expected one JSON object", http.StatusBadRequest)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("writing JSON response: %v", err)
	}
}

func cleanOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	cleaned := strings.TrimSpace(*value)
	if cleaned == "" {
		return nil
	}
	return &cleaned
}

func imageMetadata(file multipart.File, header *multipart.FileHeader) (string, string, error) {
	buffer := make([]byte, 512)
	count, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", "", errors.New("could not read icon")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", "", errors.New("could not read icon")
	}
	mimeType := http.DetectContentType(buffer[:count])
	extensions := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"image/gif":  ".gif",
	}
	extension, ok := extensions[mimeType]
	if !ok {
		return "", "", fmt.Errorf("%s is not a supported image type", header.Filename)
	}
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return "", "", fmt.Errorf("%s is not a valid image", header.Filename)
	}
	if config.Width > maxIconDimension || config.Height > maxIconDimension {
		return "", "", fmt.Errorf("icon dimensions must be %dx%d pixels or smaller", maxIconDimension, maxIconDimension)
	}
	if config.Width != config.Height {
		return "", "", errors.New("icon must be square")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", "", errors.New("could not read icon")
	}
	return extension, mimeType, nil
}

func randomFilename(extension string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes) + extension, nil
}
