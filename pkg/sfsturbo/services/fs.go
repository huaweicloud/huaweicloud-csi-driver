package services

import (
	"encoding/json"
	"errors"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/sfs_turbo/v1/fs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
)

const (
	// dirQuotaNotCreatedCode is returned when the queried path has no quota.
	dirQuotaNotCreatedCode = "SFS.TURBO.0113"
)

// sfsTurboError mirrors the error body returned by the SFS Turbo API:
//
//	{"errCode":"SFS.TURBO.0113","errMsg":"dir not create quota"}
type sfsTurboError struct {
	ErrCode string `json:"errCode"`
	ErrMsg  string `json:"errMsg"`
}

// parseSFSTurboErrCode extracts the errCode field from the Body of a golangsdk
// HTTP error (e.g. ErrDefault400). Returns ("", false) when the error is not a
// 400/404/409 golangsdk HTTP error or the body cannot be parsed as an SFS Turbo
// error envelope.
//
// Note: we cannot use errors.As to unwrap ErrUnexpectedResponseCode from
// ErrDefault400, because the golangsdk error types embed it by value (not by
// pointer) and do not implement Unwrap. errors.As does not traverse
// value-embedded fields. Instead we type-switch on the concrete golangsdk
// error types, mirroring the approach used by the existing isHTTPNotFound /
// isHTTPConflict helpers.
func parseSFSTurboErrCode(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	var body []byte
	switch e := err.(type) {
	case golangsdk.ErrDefault400:
		body = e.Body
	case golangsdk.ErrDefault401:
		body = e.Body
	case golangsdk.ErrDefault403:
		body = e.Body
	case golangsdk.ErrDefault404:
		body = e.Body
	case golangsdk.ErrDefault409:
		body = e.Body
	case golangsdk.ErrUnexpectedResponseCode:
		body = e.Body
	default:
		return "", false
	}
	if len(body) == 0 {
		return "", false
	}
	var se sfsTurboError
	if err := json.Unmarshal(body, &se); err != nil {
		return "", false
	}
	if se.ErrCode == "" {
		return "", false
	}
	return se.ErrCode, true
}

// isDirQuotaNotCreated reports whether err is an SFS Turbo error indicating
// that the path has no quota set (HTTP 400 + errCode SFS.TURBO.0113).
func isDirQuotaNotCreated(err error) bool {
	code, ok := parseSFSTurboErrCode(err)
	if !ok {
		return false
	}
	return code == dirQuotaNotCreatedCode
}

// CreateDir creates a directory on the specified SFS Turbo share via the REST
// API. If the directory already exists (409), it is treated as success so the
// flow is idempotent.
func CreateDir(c *config.CloudCredentials, shareID, path string) error {
	client, err := getSFSTurboV1Client(c)
	if err != nil {
		return err
	}
	opts := fs.CreateDirOpts{
		Path: path,
	}
	log.V(4).Infof("[DEBUG] CreateDir shareID=%s path=%s", shareID, path)
	result := fs.CreateDir(client, shareID, opts)
	if result.Err != nil {
		// 409 means the directory already exists — treat as success.
		if isHTTPConflict(result.Err) {
			log.V(4).Infof("[DEBUG] directory %s already exists on share %s, using it", path, shareID)
			return nil
		}
		return status.Errorf(codes.Internal, "Failed to create directory %s on share %s: %v", path, shareID, result.Err)
	}
	return nil
}

// ShowDir checks whether a directory exists on the specified SFS Turbo share.
// Returns (info, true, nil) if the directory exists, (nil, false, nil) if it
// does not exist (404), or (nil, false, err) on other errors.
func ShowDir(c *config.CloudCredentials, shareID, path string) (*fs.DirInfo, bool, error) {
	client, err := getSFSTurboV1Client(c)
	if err != nil {
		return nil, false, err
	}
	result := fs.ShowDir(client, shareID, path)
	if result.Err != nil {
		if isHTTPNotFound(result.Err) {
			return nil, false, nil
		}
		return nil, false, status.Errorf(codes.Internal,
			"Failed to query directory %s on share %s: %v", path, shareID, result.Err)
	}
	info, err := result.Extract()
	if err != nil {
		return nil, false, status.Errorf(codes.Internal,
			"Failed to extract directory info: %v", err)
	}
	return info, true, nil
}

// DeleteDir deletes a directory on the specified SFS Turbo share.
func DeleteDir(c *config.CloudCredentials, shareID, path string) error {
	client, err := getSFSTurboV1Client(c)
	if err != nil {
		return err
	}
	opts := fs.DeleteDirOpts{
		Path: path,
	}
	log.V(4).Infof("[DEBUG] DeleteDir shareID=%s path=%s", shareID, path)
	result := fs.DeleteDir(client, shareID, opts)
	if result.Err != nil {
		if isHTTPNotFound(result.Err) {
			log.V(4).Infof("[DEBUG] directory %s not found on share %s, already deleted", path, shareID)
			return nil
		}
		return status.Errorf(codes.Internal, "Failed to delete directory %s on share %s: %v", path, shareID, result.Err)
	}
	return nil
}

// CreateQuota creates a directory quota. capacityMB is the capacity in MB.
func CreateQuota(c *config.CloudCredentials, shareID, path string, capacityMB int) error {
	client, err := getSFSTurboV1Client(c)
	if err != nil {
		return err
	}
	opts := fs.QuotaOpts{
		Path:     path,
		Capacity: &capacityMB,
	}
	log.V(4).Infof("[DEBUG] CreateQuota shareID=%s path=%s capacityMB=%d", shareID, path, capacityMB)
	result := fs.CreateQuota(client, shareID, opts)
	if result.Err != nil {
		return status.Errorf(codes.Internal, "Failed to create quota for directory %s on share %s: %v", path, shareID, result.Err)
	}
	return nil
}

// ShowQuota queries the directory quota. Returns (quota, true, nil) if the
// quota exists, (nil, false, nil) if it does not exist. The "not exists" case
// covers both HTTP 404 and HTTP 400 with errCode SFS.TURBO.0113 ("dir not
// create quota"), which is what the SFS Turbo API returns when the path has
// no quota set.
func ShowQuota(c *config.CloudCredentials, shareID, path string) (*fs.Quota, bool, error) {
	client, err := getSFSTurboV1Client(c)
	if err != nil {
		return nil, false, err
	}
	result := fs.ShowQuota(client, shareID, path)
	if result.Err != nil {
		if isHTTPNotFound(result.Err) || isDirQuotaNotCreated(result.Err) {
			return nil, false, nil
		}
		return nil, false, status.Errorf(codes.Internal,
			"Failed to query quota for directory %s on share %s: %v", path, shareID, result.Err)
	}
	quota, err := result.Extract()
	if err != nil {
		return nil, false, status.Errorf(codes.Internal, "Failed to extract quota info: %v", err)
	}
	return quota, true, nil
}

// UpdateQuota updates an existing directory quota. capacityMB is the new
// capacity in MB.
func UpdateQuota(c *config.CloudCredentials, shareID, path string, capacityMB int) error {
	client, err := getSFSTurboV1Client(c)
	if err != nil {
		return err
	}
	opts := fs.QuotaOpts{
		Path:     path,
		Capacity: &capacityMB,
	}
	log.V(4).Infof("[DEBUG] UpdateQuota shareID=%s path=%s capacityMB=%d", shareID, path, capacityMB)
	result := fs.UpdateQuota(client, shareID, opts)
	if result.Err != nil {
		return status.Errorf(codes.Internal, "Failed to update quota for directory %s on share %s: %v", path, shareID, result.Err)
	}
	return nil
}

// DeleteQuota deletes a directory quota. If the quota does not exist (404 or
// 400 + SFS.TURBO.0113 "dir not create quota"), it is treated as success so
// the operation is idempotent.
func DeleteQuota(c *config.CloudCredentials, shareID, path string) error {
	client, err := getSFSTurboV1Client(c)
	if err != nil {
		return err
	}
	log.V(4).Infof("[DEBUG] DeleteQuota shareID=%s path=%s", shareID, path)
	result := fs.DeleteQuota(client, shareID, path)
	if result.Err != nil {
		if isHTTPNotFound(result.Err) || isDirQuotaNotCreated(result.Err) {
			log.V(4).Infof("[DEBUG] quota for directory %s not found on share %s, already deleted", path, shareID)
			return nil
		}
		return status.Errorf(codes.Internal, "Failed to delete quota for directory %s on share %s: %v", path, shareID, result.Err)
	}
	return nil
}

// EnsureQuota ensures the directory quota is set to the requested capacity.
// It first tries to show the existing quota: if not found, it creates one; if
// found, it only grows the quota — the existing quota is updated only when it
// is strictly less than the requested capacity. An existing quota that is
// greater than or equal to the request is left untouched so a sub-directory
// that was previously given a larger quota (or an unlimited one, capacity 0
// being treated as the literal value 0) is never shrunk on CreateVolume.
func EnsureQuota(c *config.CloudCredentials, shareID, path string, capacityMB int) error {
	quota, exists, err := ShowQuota(c, shareID, path)
	if err != nil {
		return err
	}
	if !exists {
		log.Infof("Quota does not exist for path %s on share %s, creating with %d MB", path, shareID, capacityMB)
		return CreateQuota(c, shareID, path, capacityMB)
	}
	// A capacity of 0 means "unlimited" in the SFS Turbo API. Treat it as
	// "no effective quota set" so that the requested capacity is applied,
	// rather than treating 0 as a literal value that is always < requested
	// (which would work) but also skip the update when comparing >= (which
	// it would not, since 0 < any positive). Log a warning so the operator
	// is aware that an unlimited quota is being replaced with a fixed one.
	if quota.Capacity == 0 {
		log.Warningf("Existing quota for path %s on share %s is unlimited (capacity=0), "+
			"treating as not-set and updating to %d MB", path, shareID, capacityMB)
		return UpdateQuota(c, shareID, path, capacityMB)
	}
	if quota.Capacity >= capacityMB {
		log.Infof("Quota for path %s on share %s already %d MB, no shrink needed (requested %d MB)",
			path, shareID, quota.Capacity, capacityMB)
		return nil
	}
	log.Infof("Updating quota for path %s on share %s from %d MB to %d MB", path, shareID, quota.Capacity, capacityMB)
	return UpdateQuota(c, shareID, path, capacityMB)
}

// --- HTTP error helpers ---

func isHTTPNotFound(err error) bool {
	if err == nil {
		return false
	}
	var target404 golangsdk.ErrDefault404
	return errors.As(err, &target404)
}

func isHTTPConflict(err error) bool {
	if err == nil {
		return false
	}
	var target409 golangsdk.ErrDefault409
	return errors.As(err, &target409)
}
