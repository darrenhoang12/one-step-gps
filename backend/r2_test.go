package main

import (
	"context"
	"testing"
)

func TestNewR2Storage(t *testing.T) {
	t.Setenv("R2_ENDPOINT", "https://account.r2.cloudflarestorage.com/icons")
	t.Setenv("R2_BUCKET_NAME", "icons")
	t.Setenv("R2_ACCESS_KEY_ID", "test-access-key")
	t.Setenv("R2_SECRET_ACCESS_KEY", "test-secret-key")
	t.Setenv("R2_PUBLIC_BASE_URL", "https://pub-example.r2.dev/")

	storage, err := newR2Storage(context.Background())
	if err != nil {
		t.Fatalf("newR2Storage() error = %v", err)
	}
	if got := storage.PublicURL("device-icons/icon.png"); got != "https://pub-example.r2.dev/device-icons/icon.png" {
		t.Fatalf("PublicURL() = %q", got)
	}
}

func TestNewR2StorageRejectsAPIEndpointAsPublicURL(t *testing.T) {
	t.Setenv("R2_ENDPOINT", "https://account.r2.cloudflarestorage.com")
	t.Setenv("R2_BUCKET_NAME", "icons")
	t.Setenv("R2_ACCESS_KEY_ID", "test-access-key")
	t.Setenv("R2_SECRET_ACCESS_KEY", "test-secret-key")
	t.Setenv("R2_PUBLIC_BASE_URL", "https://account.r2.cloudflarestorage.com/icons")

	if _, err := newR2Storage(context.Background()); err == nil {
		t.Fatal("newR2Storage() expected an invalid public URL error")
	}
}
