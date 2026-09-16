package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestCompleteDeviceOrderAppendsSavedDevicesNotInRequest(t *testing.T) {
	requested := []string{"c", "a", "b"}
	existing := []string{"a", "hidden", "b", "old", "c"}
	want := []string{"c", "a", "b", "hidden", "old"}
	if got := completeDeviceOrder(requested, existing); !reflect.DeepEqual(got, want) {
		t.Fatalf("complete order = %v, want %v", got, want)
	}
}

func TestPreferencesRejectClientSortOrder(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/preferences", strings.NewReader(`{"device_id":"abc","sort_order":1}`))
	response := httptest.NewRecorder()
	preferencesHandler(nil).ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("client sort order status = %d, want 400", response.Code)
	}
}
