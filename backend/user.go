package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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

func deviceIconHandler(db *pgxpool.Pool, uploadDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
		if err := r.ParseMultipartForm(5 << 20); err != nil {
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
		extension, err := imageExtension(file, header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := os.MkdirAll(uploadDir, 0o755); err != nil {
			http.Error(w, "could not prepare icon storage", http.StatusInternalServerError)
			return
		}
		filename, err := randomFilename(extension)
		if err != nil {
			http.Error(w, "could not create icon name", http.StatusInternalServerError)
			return
		}
		destinationPath := filepath.Join(uploadDir, filename)
		destination, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			http.Error(w, "could not store icon", http.StatusInternalServerError)
			return
		}
		_, copyErr := io.Copy(destination, file)
		closeErr := destination.Close()
		if copyErr != nil || closeErr != nil {
			_ = os.Remove(destinationPath)
			http.Error(w, "could not store icon", http.StatusInternalServerError)
			return
		}

		storagePath := "/uploads/" + filename
		_, err = db.Exec(r.Context(), `
			INSERT INTO device_preferences (device_id, icon_storage_path)
			VALUES ($1, $2)
			ON CONFLICT (device_id) DO UPDATE SET icon_storage_path = EXCLUDED.icon_storage_path, updated_at = NOW()
		`, deviceID, storagePath)
		if err != nil {
			_ = os.Remove(destinationPath)
			log.Printf("saving device icon: %v", err)
			http.Error(w, "could not save icon", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"icon_storage_path": storagePath})
	}
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

func imageExtension(file multipart.File, header *multipart.FileHeader) (string, error) {
	buffer := make([]byte, 512)
	count, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", errors.New("could not read icon")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", errors.New("could not read icon")
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
		return "", fmt.Errorf("%s is not a supported image type", header.Filename)
	}
	return extension, nil
}

func randomFilename(extension string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes) + extension, nil
}
