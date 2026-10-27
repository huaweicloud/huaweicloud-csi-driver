package fs

import "github.com/chnsz/golangsdk"

// DirInfo contains the information returned by the ShowDir API.
type DirInfo struct {
	// The full path of the directory.
	Path string `json:"path"`
	// Directory permission in octal (e.g. 40755). Only returned by some
	// SFS Turbo specs.
	Mode *int64 `json:"mode,omitempty"`
	// Owner user ID. Only returned by some SFS Turbo specs.
	UID *int64 `json:"uid,omitempty"`
	// Owner group ID. Only returned by some SFS Turbo specs.
	GID *int64 `json:"gid,omitempty"`
}

// Quota contains the directory quota information returned by the CreateQuota,
// ShowQuota, and UpdateQuota APIs.
type Quota struct {
	// The full path of the directory.
	Path string `json:"path"`
	// Capacity quota in MB.
	Capacity int `json:"capacity"`
	// Inode count quota.
	Inode int `json:"inode"`
	// Used capacity in MB. Only returned by some SFS Turbo specs.
	UsedCapacity int `json:"used_capacity,omitempty"`
	// Used inode count. Only returned by some SFS Turbo specs.
	UsedInode int `json:"used_inode,omitempty"`
}

// ErrorBody represents the error response returned by the SFS Turbo directory
// management APIs on failure.
type ErrorBody struct {
	ErrCode string `json:"errCode"`
	ErrMsg  string `json:"errMsg"`
}

// ---- Results ----

type commonResult struct {
	golangsdk.Result
}

// CreateDirResult contains the response from a CreateDir request. The API
// returns 204 with no body on success, or 409 with an error body if the
// directory already exists.
type CreateDirResult struct {
	golangsdk.ErrResult
}

// ShowDirResult contains the response body and error from a ShowDir request.
type ShowDirResult struct {
	commonResult
}

// DeleteDirResult contains the response from a DeleteDir request. The API
// returns 202 with no body on success.
type DeleteDirResult struct {
	golangsdk.ErrResult
}

// QuotaResult contains the response body and error from a quota request
// (Create/Show/Update).
type QuotaResult struct {
	commonResult
}

// DeleteQuotaResult contains the response from a DeleteQuota request. The API
// returns 204 with no body on success.
type DeleteQuotaResult struct {
	golangsdk.ErrResult
}

// Extract gets the DirInfo object from a ShowDirResult.
func (r ShowDirResult) Extract() (*DirInfo, error) {
	var s DirInfo
	err := r.ExtractInto(&s)
	return &s, err
}

// Extract gets the Quota object from a QuotaResult.
func (r QuotaResult) Extract() (*Quota, error) {
	var s Quota
	err := r.ExtractInto(&s)
	return &s, err
}
