// This source code file is AUTO-GENERATED by github.com/taskcluster/jsonschema2go

package tcobject

import (
	"encoding/json"
	"errors"

	tcclient "github.com/taskcluster/taskcluster/v39/clients/client-go"
)

type (
	Details struct {
		URL string `json:"url"`
	}

	// See [Download Methods](https://docs.taskcluster.net/docs/docs/reference/platform/object/download-methods) for details.
	DownloadObjectRequest struct {

		// Download methods that the caller can suport, together with parameters for each method.
		// The server will choose one method and make the corresponding response.
		AcceptDownloadMethods SupportedDownloadMethods `json:"acceptDownloadMethods"`
	}

	// See [Download Methods](https://docs.taskcluster.net/docs/docs/reference/platform/object/download-methods) for details.
	//
	// One of:
	//   * HTTPGETDownloadResponse
	//   * SimpleDownloadResponse
	DownloadObjectResponse json.RawMessage

	HTTPGETDownloadResponse struct {
		Details Details `json:"details"`

		// Constant value: "HTTP:GET"
		Method string `json:"method"`
	}

	// A simple download returns a URL to which the caller should make a GET request.
	// See [Simple Downloads](https://docs.taskcluster.net/docs/docs/reference/platform/object/simple-downloads) for details.
	SimpleDownloadResponse struct {

		// Constant value: "simple"
		Method string `json:"method"`

		URL string `json:"url"`
	}

	// Download methods that the caller can suport, together with parameters for each method.
	// The server will choose one method and make the corresponding response.
	SupportedDownloadMethods struct {

		// Constant value: %!q(bool=true)
		HTTPGET bool `json:"HTTP:GET,omitempty"`

		// Constant value: %!q(bool=true)
		Simple bool `json:"simple,omitempty"`
	}

	// Representation of the object entry to insert.  This is a temporary API.
	UploadObjectRequest struct {

		// The data to upload, base64-encoded
		Data string `json:"data"`

		// Date at which this entry expires from the object table.
		Expires tcclient.Time `json:"expires"`

		// Project identifier.
		ProjectID string `json:"projectId"`
	}
)

// MarshalJSON calls json.RawMessage method of the same name. Required since
// DownloadObjectResponse is of type json.RawMessage...
func (this *DownloadObjectResponse) MarshalJSON() ([]byte, error) {
	x := json.RawMessage(*this)
	return (&x).MarshalJSON()
}

// UnmarshalJSON is a copy of the json.RawMessage implementation.
func (this *DownloadObjectResponse) UnmarshalJSON(data []byte) error {
	if this == nil {
		return errors.New("DownloadObjectResponse: UnmarshalJSON on nil pointer")
	}
	*this = append((*this)[0:0], data...)
	return nil
}
