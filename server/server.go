package server

import (
	"context"
	"crypto"
	"encoding/hex"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/jcjones/ocsp-l2-cache/repo"
	"golang.org/x/crypto/ocsp"
)

type OcspFrontEnd struct {
	server *http.Server
	store  *repo.OcspStore
}

func NewOcspFrontEnd(listenAddr string, store *repo.OcspStore) (*OcspFrontEnd, error) {
	server := &http.Server{
		Addr: listenAddr,
	}
	return &OcspFrontEnd{server, store}, nil
}

func (ocs *OcspFrontEnd) HandleQuery(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Unable to read, %v", err)
		http.Error(w, "Unable to read", http.StatusBadRequest)
		return
	}

	req, err := ocsp.ParseRequest(body)
	if err != nil {
		log.Printf("Unable to parse: %v\n%s", err, hex.Dump(body))
		http.Error(w, "Unable to parse", http.StatusBadRequest)
		return
	}

	log.Printf("Request: %+v", req)

	if !ocs.isConfiguredIssuer(req.IssuerKeyHash, req.HashAlgorithm) {
		log.Printf("Unknown issuer: %s {%+v}", req.IssuerKeyHash, req)
		ocs.unknownIssuer(w)
		return
	}

	resp, err := ocs.store.Get(ctx, req, body)
	if err == repo.UpstreamError {
		log.Printf("Upstream error: %s {%+v}", err, req)
		ocs.upstreamError(w)
		return
	} else if err != nil {
		log.Printf("Unable to obtain response: %v", err)
		http.Error(w, "Failed", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/ocsp-response")
	w.Write(resp)
}

func (ocs *OcspFrontEnd) ListenAndServe() error {
	http.HandleFunc("/", ocs.HandleQuery)
	done := make(chan bool)
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		log.Printf("Signal caught, HTTP server shutting down.")

		// We received an interrupt signal, shut down.
		_ = ocs.server.Shutdown(context.Background())
		done <- true
	}()

	if err := ocs.server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	<-done
	log.Printf("HTTP server offline.")
	return nil
}

func (ocs *OcspFrontEnd) isConfiguredIssuer(issuerKeyHash []byte, hashAlgo crypto.Hash) bool {
	// TODO
	return true
}

func (ocs *OcspFrontEnd) unknownIssuer(w http.ResponseWriter) {
	// TODO
	http.Error(w, "TODO: return a real unknown issuer response", http.StatusNotFound)
}

func (ocs *OcspFrontEnd) upstreamError(w http.ResponseWriter) {
	// TODO
	http.Error(w, "TODO: return a real upstream error response", http.StatusServiceUnavailable)
}
