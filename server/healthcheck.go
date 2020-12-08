// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jcjones/ocsp-l2-cache/storage"
	blog "github.com/letsencrypt/boulder/log"
)

type HealthCheck struct {
	logger blog.Logger
	cache  storage.RemoteCache
}

func NewHealthCheck(logger blog.Logger, cache storage.RemoteCache) *HealthCheck {
	return &HealthCheck{logger, cache}
}

func (hc *HealthCheck) HandleQuery(response http.ResponseWriter, request *http.Request) {
	data, healtherr := hc.cache.Info(context.Background())

	if healtherr == nil {
		response.WriteHeader(200)
		_, err := response.Write([]byte(fmt.Sprintf("ok: cache is alive\ninfo:\n%v", data)))
		if err != nil {
			hc.logger.Warningf("Couldn't return ok health status: %+v", err)
		}
		return
	}

	response.WriteHeader(500)
	_, err := response.Write([]byte(fmt.Sprintf("failed: %v", healtherr)))
	if err != nil {
		hc.logger.Warningf("Couldn't return ok health status: %+v", err)
	}

}
