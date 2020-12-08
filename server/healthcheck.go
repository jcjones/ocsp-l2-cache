// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package server

import (
	"fmt"
	"log"
	"net/http"
)

type HealthCheck struct {
}

func NewHealthCheck() *HealthCheck {
	return &HealthCheck{}
}

func (hc *HealthCheck) HandleQuery(response http.ResponseWriter, request *http.Request) {
	response.WriteHeader(200)
	_, err := response.Write([]byte(fmt.Sprintf("ok: something")))
	if err != nil {
		log.Printf("Couldn't return ok health status: %+v", err)
	}
}
