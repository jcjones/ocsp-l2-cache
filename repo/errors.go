// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package repo

const UpstreamError = OcspStoreError("upstream")

const UnknownIssuerError = OcspStoreError("unknown issuer")

type OcspStoreError string

func (e OcspStoreError) Error() string { return string(e) }

func (OcspStoreError) OcspStoreError() {}
