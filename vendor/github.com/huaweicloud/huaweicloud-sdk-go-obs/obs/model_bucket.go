// Copyright 2019 Huawei Technologies Co.,Ltd.
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use
// this file except in compliance with the License.  You may obtain a copy of the
// License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
// CONDITIONS OF ANY KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations under the License.
package obs

import (
	"encoding/xml"
)

// ListBucketsInput is the input parameter of ListBuckets function
type ListBucketsInput struct {
	QueryLocation bool
	BucketType    BucketType
}

// ListBucketsOutput is the result of ListBuckets function
type ListBucketsOutput struct {
	BaseModel
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	Owner   Owner    `xml:"Owner"`
	Buckets []Bucket `xml:"Buckets>Bucket"`
}

// CreateBucketInput is the input parameter of CreateBucket function
type CreateBucketInput struct {
	BucketLocation
	Bucket                      string           `xml:"-"`
	ACL                         AclType          `xml:"-"`
	StorageClass                StorageClassType `xml:"-"`
	GrantReadId                 string           `xml:"-"`
	GrantWriteId                string           `xml:"-"`
	GrantReadAcpId              string           `xml:"-"`
	GrantWriteAcpId             string           `xml:"-"`
	GrantFullControlId          string           `xml:"-"`
	GrantReadDeliveredId        string           `xml:"-"`
	GrantFullControlDeliveredId string           `xml:"-"`
	Epid                        string           `xml:"-"`
	AvailableZone               string           `xml:"-"`
	IsFSFileInterface           bool             `xml:"-"`
}

// SetBucketStoragePolicyInput is the input parameter of SetBucketStoragePolicy function
type SetBucketStoragePolicyInput struct {
	Bucket string `xml:"-"`
	BucketStoragePolicy
}

type getBucketStoragePolicyOutputS3 struct {
	BaseModel
	BucketStoragePolicy
}

// GetBucketStoragePolicyOutput is the result of GetBucketStoragePolicy function
type GetBucketStoragePolicyOutput struct {
	BaseModel
	StorageClass string
}

type getBucketStoragePolicyOutputObs struct {
	BaseModel
	bucketStoragePolicyObs
}

// SetBucketQuotaInput is the input parameter of SetBucketQuota function
type SetBucketQuotaInput struct {
	Bucket string `xml:"-"`
	BucketQuota
}

// GetBucketQuotaOutput is the result of GetBucketQuota function
type GetBucketQuotaOutput struct {
	BaseModel
	BucketQuota
}

// GetBucketStorageInfoOutput is the result of GetBucketStorageInfo function
type GetBucketStorageInfoOutput struct {
	BaseModel
	XMLName      xml.Name `xml:"GetBucketStorageInfoResult"`
	Size         int64    `xml:"Size"`
	ObjectNumber int      `xml:"ObjectNumber"`
}

type getBucketLocationOutputS3 struct {
	BaseModel
	BucketLocation
}
type getBucketLocationOutputObs struct {
	BaseModel
	bucketLocationObs
}

// GetBucketLocationOutput is the result of GetBucketLocation function
type GetBucketLocationOutput struct {
	BaseModel
	Location string `xml:"-"`
}

// GetBucketAclOutput is the result of GetBucketAcl function
type GetBucketAclOutput struct {
	BaseModel
	AccessControlPolicy
}

type getBucketACLOutputObs struct {
	BaseModel
	accessControlPolicyObs
}

// SetBucketAclInput is the input parameter of SetBucketAcl function
type SetBucketAclInput struct {
	Bucket string  `xml:"-"`
	ACL    AclType `xml:"-"`
	AccessControlPolicy
}

// SetBucketPolicyInput is the input parameter of SetBucketPolicy function
type SetBucketPolicyInput struct {
	Bucket string
	Policy string
}

// GetBucketPolicyOutput is the result of GetBucketPolicy function
type GetBucketPolicyOutput struct {
	BaseModel
	Policy string `json:"body"`
}

// SetBucketCorsInput is the input parameter of SetBucketCors function
type SetBucketCorsInput struct {
	Bucket string `xml:"-"`
	BucketCors
}

// GetBucketCorsOutput is the result of GetBucketCors function
type GetBucketCorsOutput struct {
	BaseModel
	BucketCors
}

// SetBucketVersioningInput is the input parameter of SetBucketVersioning function
type SetBucketVersioningInput struct {
	Bucket string `xml:"-"`
	BucketVersioningConfiguration
}

// GetBucketVersioningOutput is the result of GetBucketVersioning function
type GetBucketVersioningOutput struct {
	BaseModel
	BucketVersioningConfiguration
}

// SetBucketWebsiteConfigurationInput is the input parameter of SetBucketWebsiteConfiguration function
type SetBucketWebsiteConfigurationInput struct {
	Bucket string `xml:"-"`
	BucketWebsiteConfiguration
}

// GetBucketWebsiteConfigurationOutput is the result of GetBucketWebsiteConfiguration function
type GetBucketWebsiteConfigurationOutput struct {
	BaseModel
	BucketWebsiteConfiguration
}

// GetBucketMetadataInput is the input parameter of GetBucketMetadata function
type GetBucketMetadataInput struct {
	Bucket        string
	Origin        string
	RequestHeader string
}

// SetObjectMetadataInput is the input parameter of SetObjectMetadata function
type SetObjectMetadataInput struct {
	Bucket                  string
	Key                     string
	VersionId               string
	MetadataDirective       MetadataDirectiveType
	CacheControl            string
	ContentDisposition      string
	ContentEncoding         string
	ContentLanguage         string
	ContentType             string
	Expires                 string
	WebsiteRedirectLocation string
	StorageClass            StorageClassType
	Metadata                map[string]string
}

//SetObjectMetadataOutput is the result of SetObjectMetadata function
type SetObjectMetadataOutput struct {
	BaseModel
	MetadataDirective       MetadataDirectiveType
	CacheControl            string
	ContentDisposition      string
	ContentEncoding         string
	ContentLanguage         string
	ContentType             string
	Expires                 string
	WebsiteRedirectLocation string
	StorageClass            StorageClassType
	Metadata                map[string]string
}

// GetBucketMetadataOutput is the result of GetBucketMetadata function
type GetBucketMetadataOutput struct {
	BaseModel
	StorageClass  StorageClassType
	Location      string
	Version       string
	AllowOrigin   string
	AllowMethod   string
	AllowHeader   string
	MaxAgeSeconds int
	ExposeHeader  string
	Epid          string
	AZRedundancy  string
	FSStatus      FSStatusType
}

// SetBucketLoggingConfigurationInput is the input parameter of SetBucketLoggingConfiguration function
type SetBucketLoggingConfigurationInput struct {
	Bucket string `xml:"-"`
	BucketLoggingStatus
}

// GetBucketLoggingConfigurationOutput is the result of GetBucketLoggingConfiguration function
type GetBucketLoggingConfigurationOutput struct {
	BaseModel
	BucketLoggingStatus
}

// BucketLifecyleConfiguration defines the bucket lifecycle configuration
type BucketLifecyleConfiguration struct {
	XMLName        xml.Name        `xml:"LifecycleConfiguration"`
	LifecycleRules []LifecycleRule `xml:"Rule"`
}

// SetBucketLifecycleConfigurationInput is the input parameter of SetBucketLifecycleConfiguration function
type SetBucketLifecycleConfigurationInput struct {
	Bucket string `xml:"-"`
	BucketLifecyleConfiguration
}

// GetBucketLifecycleConfigurationOutput is the result of GetBucketLifecycleConfiguration function
type GetBucketLifecycleConfigurationOutput struct {
	BaseModel
	BucketLifecyleConfiguration
}

// SetBucketEncryptionInput is the input parameter of SetBucketEncryption function
type SetBucketEncryptionInput struct {
	Bucket string `xml:"-"`
	BucketEncryptionConfiguration
}

// GetBucketEncryptionOutput is the result of GetBucketEncryption function
type GetBucketEncryptionOutput struct {
	BaseModel
	BucketEncryptionConfiguration
}

// SetBucketTaggingInput is the input parameter of SetBucketTagging function
type SetBucketTaggingInput struct {
	Bucket string `xml:"-"`
	BucketTagging
}

// GetBucketTaggingOutput is the result of GetBucketTagging function
type GetBucketTaggingOutput struct {
	BaseModel
	BucketTagging
}

// SetBucketNotificationInput is the input parameter of SetBucketNotification function
type SetBucketNotificationInput struct {
	Bucket string `xml:"-"`
	BucketNotification
}

type getBucketNotificationOutputS3 struct {
	BaseModel
	bucketNotificationS3
}

// GetBucketNotificationOutput is the result of GetBucketNotification function
type GetBucketNotificationOutput struct {
	BaseModel
	BucketNotification
}

// SetBucketFetchPolicyInput is the input parameter of SetBucketFetchPolicy function
type SetBucketFetchPolicyInput struct {
	Bucket string
	Status FetchPolicyStatusType `json:"status"`
	Agency string                `json:"agency"`
}

// GetBucketFetchPolicyInput is the input parameter of GetBucketFetchPolicy function
type GetBucketFetchPolicyInput struct {
	Bucket string
}

// GetBucketFetchPolicyOutput is the result of GetBucketFetchPolicy function
type GetBucketFetchPolicyOutput struct {
	BaseModel
	FetchResponse `json:"fetch"`
}

// DeleteBucketFetchPolicyInput is the input parameter of DeleteBucketFetchPolicy function
type DeleteBucketFetchPolicyInput struct {
	Bucket string
}

// SetBucketFetchJobInput is the input parameter of SetBucketFetchJob function
type SetBucketFetchJobInput struct {
	Bucket           string            `json:"bucket"`
	URL              string            `json:"url"`
	Host             string            `json:"host,omitempty"`
	Key              string            `json:"key,omitempty"`
	Md5              string            `json:"md5,omitempty"`
	CallBackURL      string            `json:"callbackurl,omitempty"`
	CallBackBody     string            `json:"callbackbody,omitempty"`
	CallBackBodyType string            `json:"callbackbodytype,omitempty"`
	CallBackHost     string            `json:"callbackhost,omitempty"`
	FileType         string            `json:"file_type,omitempty"`
	IgnoreSameKey    bool              `json:"ignore_same_key,omitempty"`
	ObjectHeaders    map[string]string `json:"objectheaders,omitempty"`
	Etag             string            `json:"etag,omitempty"`
	TrustName        string            `json:"trustname,omitempty"`
}

// SetBucketFetchJobOutput is the result of SetBucketFetchJob function
type SetBucketFetchJobOutput struct {
	BaseModel
	SetBucketFetchJobResponse
}

// GetBucketFetchJobInput is the input parameter of GetBucketFetchJob function
type GetBucketFetchJobInput struct {
	Bucket string
	JobID  string
}

// GetBucketFetchJobOutput is the result of GetBucketFetchJob function
type GetBucketFetchJobOutput struct {
	BaseModel
	GetBucketFetchJobResponse
}
