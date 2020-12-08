package server

import (
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jcjones/ocsp-l2-cache/storage"
)

func TestHealthDown(t *testing.T) {
	t.Parallel()

	mock := storage.NewMockRemoteCache()
	mock.Alive = false
	hc := NewHealthCheck(mock)

	recorder := httptest.NewRecorder()
	hc.HandleQuery(recorder, httptest.NewRequest("GET", "/", nil))

	response := recorder.Result()

	if response.StatusCode != 500 {
		t.Errorf("Expected a 500 error, got %+v", response)
	}

	requestBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Error(err)
	}

	if string(requestBody) != "failed: Not alive" {
		t.Errorf("Unexpected body: %s", requestBody)
	}
}

func TestHealthUp(t *testing.T) {
	t.Parallel()

	mock := storage.NewMockRemoteCache()
	mock.Alive = true
	hc := NewHealthCheck(mock)

	recorder := httptest.NewRecorder()
	hc.HandleQuery(recorder, httptest.NewRequest("GET", "/", nil))

	response := recorder.Result()

	if response.StatusCode != 200 {
		t.Errorf("Expected a 200 okay, got %+v", response)
	}

	requestBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Error(err)
	}

	if !strings.HasPrefix(string(requestBody), "ok: cache is alive") {
		t.Errorf("Unexpected body: %s", requestBody)
	}
}
