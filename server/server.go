package server

import (
	"context"
	"crypto"
	"encoding/base64"
	"encoding/hex"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/jcjones/ocsp-l2-cache/common"
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

func (ocs *OcspFrontEnd) HandleQuery(response http.ResponseWriter, request *http.Request) {
	ctx := context.Background()

	// By default we set a 'max-age=0, no-cache' Cache-Control header, this
	// is only returned to the client if a valid authorized OCSP response
	// is not found or an error is returned. If a response if found the header
	// will be altered to contain the proper max-age and modifiers.
	response.Header().Set("Cache-Control", "max-age=0, no-cache")

	// Read response from request
	var requestBody []byte
	var err error

	switch request.Method {
	case "GET":
		base64Request, err := url.QueryUnescape(request.URL.Path)
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		// url.QueryUnescape not only unescapes %2B escaping, but it additionally
		// turns the resulting '+' into a space, which makes base64 decoding fail.
		// So we go back afterwards and turn ' ' back into '+'. This means we
		// accept some malformed input that includes ' ' or %20, but that's fine.
		base64RequestBytes := []byte(base64Request)
		for i := range base64RequestBytes {
			if base64RequestBytes[i] == ' ' {
				base64RequestBytes[i] = '+'
			}
		}
		// In certain situations a UA may construct a request that has a double
		// slash between the host name and the base64 request body due to naively
		// constructing the request URL. In that case strip the leading slash
		// so that we can still decode the request.
		if len(base64RequestBytes) > 0 && base64RequestBytes[0] == '/' {
			base64RequestBytes = base64RequestBytes[1:]
		}
		requestBody, err = base64.StdEncoding.DecodeString(string(base64RequestBytes))
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			return
		}
	case "POST":
		requestBody, err = ioutil.ReadAll(http.MaxBytesReader(nil, request.Body, 10000))
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		response.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	req, err := ocsp.ParseRequest(requestBody)
	if err != nil {
		log.Printf("Unable to parse: %v\n%s", err, hex.Dump(requestBody))
		http.Error(response, "Unable to parse", http.StatusBadRequest)
		return
	}

	log.Printf("Request: %+v", req)

	if !ocs.isConfiguredIssuer(req.IssuerKeyHash, req.HashAlgorithm) {
		log.Printf("Unknown issuer: %s {%+v}", req.IssuerKeyHash, req)
		ocs.unknownIssuer(response)
		return
	}

	responseBody, headers, err := ocs.store.Get(ctx, req, requestBody)
	if err == repo.UpstreamError {
		log.Printf("Upstream error: %s {%+v}", err, req)
		ocs.upstreamError(response)
		return
	} else if err != nil {
		log.Printf("Unable to obtain response: %v", err)
		http.Error(response, "Failed", http.StatusInternalServerError)
		return
	}

	for k, v := range headers {
		response.Header().Set(k, v)
	}
	response.Header().Set(common.HeaderContentType, common.MimeOcspResponse)

	response.Write(responseBody)
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
