package server_test

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"
)

func assertJSONResponse[T any](t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, want T) {
	t.Helper()

	if rec.Code != wantStatus {
		t.Fatalf("got status %d, want %d", rec.Code, wantStatus)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("got Content-Type %q, want %q", ct, "application/json")
	}

	var got T
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func assertErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantError string) {
	t.Helper()

	if rec.Code != wantStatus {
		t.Fatalf("got status %d, want %d", rec.Code, wantStatus)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("got Content-Type %q, want %q", ct, "application/json")
	}

	var got struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if got.Error != wantError {
		t.Errorf("got error %q, want %q", got.Error, wantError)
	}
}
