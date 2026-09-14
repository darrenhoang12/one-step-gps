package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func getDevices(w http.ResponseWriter, _ *http.Request) {
	deviceUrl := os.Getenv("ONE_STEP_GPS_DEVICE_URL")
	deviceKey := os.Getenv("ONE_STEP_GPS_DEVICE_KEY")
	requestUrl := fmt.Sprintf("%s?latest_point=true&api-key=%s", deviceUrl, deviceKey)

	resp, err := http.Get(requestUrl)
	if err != nil {
		http.Error(w, "could not fetch devices", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("statusCode: %v", resp.StatusCode)
		http.Error(w, "device service returned an error", http.StatusBadGateway)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("could not read device response: %v", err)
		http.Error(w, "could not read device response", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(body); err != nil {
		log.Printf("could not send device response: %v", err)
	}
}
