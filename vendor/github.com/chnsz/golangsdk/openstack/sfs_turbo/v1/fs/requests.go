package fs

import (
	"net/url"
	"strconv"

	"github.com/chnsz/golangsdk"
)

var requestOpts = golangsdk.RequestOpts{
	MoreHeaders: map[string]string{"Content-Type": "application/json", "X-Language": "en-us"},
}

// ---------- Directory ----------

// CreateDirOptsBuilder allows extensions to add additional parameters to the
// CreateDir request.
type CreateDirOptsBuilder interface {
	ToDirCreateMap() (map[string]interface{}, error)
}

// CreateDirOpts contains the options for creating a directory on an SFS Turbo
// file system. This object is passed to CreateDir().
type CreateDirOpts struct {
	// The full path of the directory to create. Must start with "/".
	Path string `json:"path" required:"true"`
	// Directory permission in octal, e.g. 0755. Defaults to 0755.
	Mode *int64 `json:"mode,omitempty"`
	// Owner user ID of the directory. Defaults to 0.
	UID *int64 `json:"uid,omitempty"`
	// Owner group ID of the directory. Defaults to 0.
	GID *int64 `json:"gid,omitempty"`
}

// ToDirCreateMap assembles a request body based on the contents of a
// CreateDirOpts.
func (opts CreateDirOpts) ToDirCreateMap() (map[string]interface{}, error) {
	return golangsdk.BuildRequestBody(opts, "")
}

// CreateDir creates a directory on the specified SFS Turbo file system.
// Returns 204 on success, 409 if the directory already exists.
func CreateDir(client *golangsdk.ServiceClient, shareID string, opts CreateDirOptsBuilder) (r CreateDirResult) {
	b, err := opts.ToDirCreateMap()
	if err != nil {
		r.Err = err
		return
	}
	_, r.Err = client.Post(dirURL(client, shareID), b, nil, &golangsdk.RequestOpts{
		OkCodes: []int{204},
	})
	return
}

// ShowDir queries whether a directory exists on the specified SFS Turbo file
// system. Returns the directory info on 200, or an error on 404 (not exist).
func ShowDir(client *golangsdk.ServiceClient, shareID string, path string) (r ShowDirResult) {
	q := url.Values{}
	q.Set("path", path)
	urlWithQuery := dirURL(client, shareID) + "?" + q.Encode()
	_, r.Err = client.Get(urlWithQuery, &r.Body, &golangsdk.RequestOpts{
		MoreHeaders: requestOpts.MoreHeaders,
	})
	return
}

// DeleteDirOptsBuilder allows extensions to add additional parameters to the
// DeleteDir request.
type DeleteDirOptsBuilder interface {
	ToDirDeleteMap() (map[string]interface{}, error)
}

// DeleteDirOpts contains the options for deleting a directory.
type DeleteDirOpts struct {
	// The full path of the directory to delete. Must start with "/".
	Path string `json:"path" required:"true"`
}

// ToDirDeleteMap assembles a request body based on the contents of a
// DeleteDirOpts.
func (opts DeleteDirOpts) ToDirDeleteMap() (map[string]interface{}, error) {
	return golangsdk.BuildRequestBody(opts, "")
}

// DeleteDir deletes a directory on the specified SFS Turbo file system.
// The SFS Turbo API accepts a request body (not a query parameter) for the
// path. Returns 202 on success.
func DeleteDir(client *golangsdk.ServiceClient, shareID string, opts DeleteDirOptsBuilder) (r DeleteDirResult) {
	b, err := opts.ToDirDeleteMap()
	if err != nil {
		r.Err = err
		return
	}
	_, r.Err = client.DeleteWithBody(dirURL(client, shareID), b, &golangsdk.RequestOpts{
		OkCodes: []int{202},
	})
	return
}

// ---------- Directory Quota ----------

// QuotaOptsBuilder allows extensions to add additional parameters to the
// quota create/update request.
type QuotaOptsBuilder interface {
	ToQuotaMap() (map[string]interface{}, error)
}

// QuotaOpts contains the options for creating or updating a directory quota.
// Capacity and inode are both optional but at least one must be set to a
// positive value when creating.
type QuotaOpts struct {
	// The full path of the directory. Must start with "/".
	Path string `json:"path" required:"true"`
	// Capacity quota in MB. 0 means unlimited (but both capacity and inode
	// cannot be 0 at the same time).
	Capacity *int `json:"capacity,omitempty"`
	// Inode count quota. 0 means unlimited.
	Inode *int `json:"inode,omitempty"`
}

// ToQuotaMap assembles a request body based on the contents of a QuotaOpts.
func (opts QuotaOpts) ToQuotaMap() (map[string]interface{}, error) {
	return golangsdk.BuildRequestBody(opts, "")
}

// CreateQuota creates a directory quota on the specified SFS Turbo file
// system. The directory must already exist. Returns the quota info on 200.
func CreateQuota(client *golangsdk.ServiceClient, shareID string, opts QuotaOptsBuilder) (r QuotaResult) {
	b, err := opts.ToQuotaMap()
	if err != nil {
		r.Err = err
		return
	}
	_, r.Err = client.Post(dirQuotaURL(client, shareID), b, &r.Body, &golangsdk.RequestOpts{
		OkCodes: []int{200},
	})
	return
}

// ShowQuota queries the directory quota. Returns the quota info on 200, or an
// error on 404 (quota not set).
func ShowQuota(client *golangsdk.ServiceClient, shareID string, path string) (r QuotaResult) {
	q := url.Values{}
	q.Set("path", path)
	urlWithQuery := dirQuotaURL(client, shareID) + "?" + q.Encode()
	_, r.Err = client.Get(urlWithQuery, &r.Body, &golangsdk.RequestOpts{
		MoreHeaders: requestOpts.MoreHeaders,
	})
	return
}

// UpdateQuota updates a directory quota on the specified SFS Turbo file
// system. Returns the quota info on 200.
func UpdateQuota(client *golangsdk.ServiceClient, shareID string, opts QuotaOptsBuilder) (r QuotaResult) {
	b, err := opts.ToQuotaMap()
	if err != nil {
		r.Err = err
		return
	}
	_, r.Err = client.Put(dirQuotaURL(client, shareID), b, &r.Body, &golangsdk.RequestOpts{
		OkCodes: []int{200},
	})
	return
}

// DeleteQuota deletes a directory quota. The SFS Turbo API accepts a request
// body (not a query parameter) for the path. Returns 204 on success.
func DeleteQuota(client *golangsdk.ServiceClient, shareID string, path string) (r DeleteQuotaResult) {
	b, err := DeleteDirOpts{Path: path}.ToDirDeleteMap()
	if err != nil {
		r.Err = err
		return
	}
	_, r.Err = client.DeleteWithBody(dirQuotaURL(client, shareID), b, &golangsdk.RequestOpts{
		OkCodes: []int{204},
	})
	return
}

// PathToCapacityMB converts a capacity in bytes to MB (the unit expected by
// the SFS Turbo quota API). The result is rounded up.
func PathToCapacityMB(capacityBytes int64) int {
	const mbByteSize = 1024 * 1024
	return int((capacityBytes + mbByteSize - 1) / mbByteSize)
}

// CapacityMBToString is a small helper for logging.
func CapacityMBToString(capacityMB int) string {
	return strconv.Itoa(capacityMB)
}
