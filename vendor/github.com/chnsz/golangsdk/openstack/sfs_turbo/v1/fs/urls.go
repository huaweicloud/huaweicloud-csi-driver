package fs

import "github.com/chnsz/golangsdk"

// dirURL used to assemble the URI for create/get/delete directory APIs.
// Base path: /v1/{project_id}/sfs-turbo/shares/{share_id}/fs/dir
func dirURL(c *golangsdk.ServiceClient, shareID string) string {
	return c.ServiceURL("sfs-turbo/shares", shareID, "fs/dir")
}

// dirQuotaURL used to assemble the URI for create/get/update/delete directory
// quota APIs.
// Base path: /v1/{project_id}/sfs-turbo/shares/{share_id}/fs/dir-quota
func dirQuotaURL(c *golangsdk.ServiceClient, shareID string) string {
	return c.ServiceURL("sfs-turbo/shares", shareID, "fs/dir-quota")
}
