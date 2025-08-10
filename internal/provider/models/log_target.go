// internal/provider/models/log_target.go
package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// LogTargetResourceModel represents the Terraform resource model for log targets
type LogTargetResourceModel struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Type                       types.String `tfsdk:"type"`
	Endpoint                   types.String `tfsdk:"endpoint"`
	Region                     types.String `tfsdk:"region"`
	Bucket                     types.String `tfsdk:"bucket"`
	AccessKey                  types.String `tfsdk:"access_key"`
	SecretKey                  types.String `tfsdk:"secret_key"`
	SignatureVersion           types.String `tfsdk:"signature_version"`
	JsonKey                    types.String `tfsdk:"json_key"`
	Hosts                      types.Set    `tfsdk:"hosts"`
	SSL                        types.Bool   `tfsdk:"ssl"`
	SSLCertificateVerification types.Bool   `tfsdk:"ssl_certificate_verification"`
	Index                      types.String `tfsdk:"index"`
	User                       types.String `tfsdk:"user"`
	Password                   types.String `tfsdk:"password"`
	ApiKey                     types.String `tfsdk:"api_key"`
	AccessLogsServices         types.Set    `tfsdk:"access_logs_services"`
	OriginLogsServices         types.Set    `tfsdk:"origin_logs_services"`

	// Computed fields
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// LogTargetDataSourceModel represents the data source model for log targets
type LogTargetDataSourceModel struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Type                       types.String `tfsdk:"type"`
	Endpoint                   types.String `tfsdk:"endpoint"`
	Region                     types.String `tfsdk:"region"`
	Bucket                     types.String `tfsdk:"bucket"`
	AccessKey                  types.String `tfsdk:"access_key"`
	SecretKey                  types.String `tfsdk:"secret_key"`
	SignatureVersion           types.String `tfsdk:"signature_version"`
	JsonKey                    types.String `tfsdk:"json_key"`
	Hosts                      types.Set    `tfsdk:"hosts"`
	SSL                        types.Bool   `tfsdk:"ssl"`
	SSLCertificateVerification types.Bool   `tfsdk:"ssl_certificate_verification"`
	Index                      types.String `tfsdk:"index"`
	User                       types.String `tfsdk:"user"`
	Password                   types.String `tfsdk:"password"`
	ApiKey                     types.String `tfsdk:"api_key"`
	AccessLogsServices         types.Set    `tfsdk:"access_logs_services"`
	OriginLogsServices         types.Set    `tfsdk:"origin_logs_services"`

	// Computed fields
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`

	// Optional query parameters
	ResponseType types.String `tfsdk:"response_type"`
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
