// internal/provider/models/origin.go
package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// model for cachefly_origin
type OriginResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Type                   types.String `tfsdk:"type"`
	Name                   types.String `tfsdk:"name"`
	Hostname               types.String `tfsdk:"hostname"`
	Scheme                 types.String `tfsdk:"scheme"`
	CacheByQueryParam      types.Bool   `tfsdk:"cache_by_query_param"`
	Gzip                   types.Bool   `tfsdk:"gzip"`
	TTL                    types.Int32  `tfsdk:"ttl"`
	MissedTTL              types.Int32  `tfsdk:"missed_ttl"`
	ConnectionTimeout      types.Int32  `tfsdk:"connection_timeout"`
	TimeToFirstByteTimeout types.Int32  `tfsdk:"time_to_first_byte_timeout"`

	// S3-specific fields
	AccessKey        types.String `tfsdk:"access_key"`
	SecretKey        types.String `tfsdk:"secret_key"`
	Region           types.String `tfsdk:"region"`
	SignatureVersion types.String `tfsdk:"signature_version"`

	// Computed fields
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// Terraform data source model for cachefly_origin
type OriginDataSourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Type                   types.String `tfsdk:"type"`
	Name                   types.String `tfsdk:"name"`
	Hostname               types.String `tfsdk:"hostname"`
	Scheme                 types.String `tfsdk:"scheme"`
	CacheByQueryParam      types.Bool   `tfsdk:"cache_by_query_param"`
	Gzip                   types.Bool   `tfsdk:"gzip"`
	TTL                    types.Int32  `tfsdk:"ttl"`
	MissedTTL              types.Int32  `tfsdk:"missed_ttl"`
	ConnectionTimeout      types.Int32  `tfsdk:"connection_timeout"`
	TimeToFirstByteTimeout types.Int32  `tfsdk:"time_to_first_byte_timeout"`

	// S3-specific fields
	AccessKey        types.String `tfsdk:"access_key"`
	SecretKey        types.String `tfsdk:"secret_key"`
	Region           types.String `tfsdk:"region"`
	SignatureVersion types.String `tfsdk:"signature_version"`

	// Computed fields
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`

	// Optional query parameters
	ResponseType types.String `tfsdk:"response_type"`
}

// represents the data source for listing multiple origins
type OriginsDataSourceModel struct {
	Type         types.String `tfsdk:"type"`
	Offset       types.Int64  `tfsdk:"offset"`
	Limit        types.Int64  `tfsdk:"limit"`
	ResponseType types.String `tfsdk:"response_type"`

	// Results
	Origins types.List `tfsdk:"origins"`
}
