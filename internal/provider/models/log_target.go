// internal/provider/models/log_target.go
package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// LogTargetResourceModel represents the Terraform resource model for log targets
type LogTargetResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`

	// Common log delivery options
	Format      types.String `tfsdk:"format"`
	Compression types.String `tfsdk:"compression"`
	Sampling    types.Int64  `tfsdk:"sampling"`

	// S3_BUCKET fields
	Endpoint         types.String `tfsdk:"endpoint"`
	Region           types.String `tfsdk:"region"`
	Bucket           types.String `tfsdk:"bucket"`
	AccessKey        types.String `tfsdk:"access_key"`
	SecretKey        types.String `tfsdk:"secret_key"`
	SignatureVersion types.String `tfsdk:"signature_version"`

	// GOOGLE_BUCKET fields
	JsonKey types.String `tfsdk:"json_key"`

	// AZURE_BLOB fields
	EndpointProtocol types.String `tfsdk:"endpoint_protocol"`
	EndpointSuffix   types.String `tfsdk:"endpoint_suffix"`
	AccountName      types.String `tfsdk:"account_name"`
	AccountKey       types.String `tfsdk:"account_key"`
	ContainerName    types.String `tfsdk:"container_name"`
	Prefix           types.String `tfsdk:"prefix"`

	// HTTP fields
	Uri      types.String `tfsdk:"uri"`
	Method   types.String `tfsdk:"method"`
	Auth     types.String `tfsdk:"auth"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Token    types.String `tfsdk:"token"`

	// Services logging
	AccessLogsServices types.Set `tfsdk:"access_logs_services"`
	OriginLogsServices types.Set `tfsdk:"origin_logs_services"`

	// Computed fields
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// LogTargetsDataSourceModel represents the data source for listing multiple log targets
type LogTargetsDataSourceModel struct {
	Type         types.String `tfsdk:"type"`
	Offset       types.Int64  `tfsdk:"offset"`
	Limit        types.Int64  `tfsdk:"limit"`
	ResponseType types.String `tfsdk:"response_type"`

	// Results
	LogTargets types.List `tfsdk:"log_targets"`
}
