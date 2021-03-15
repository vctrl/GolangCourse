package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteResponse(t *testing.T) {
	w := httptest.NewRecorder()
	WriteResponse(w, "test_message", http.StatusOK)
}
