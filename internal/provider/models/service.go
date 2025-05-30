package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// ServiceResourceModel describes the resource data model for managing CacheFly services.
// This maps to the Terraform resource schema and state.
type ServiceResourceModel struct {
	// Core fields - matches your SDK's Service struct
	ID         types.String `tfsdk:"id"`          // maps to Service._id
	Name       types.String `tfsdk:"name"`        // maps to Service.name
	UniqueName types.String `tfsdk:"unique_name"` // maps to Service.uniqueName

	// Configuration fields - from CreateServiceRequest and UpdateServiceRequest
	Description types.String `tfsdk:"description"` // from CreateServiceRequest.description

	// Update-only fields - from UpdateServiceRequest
	AutoSSL           types.Bool   `tfsdk:"auto_ssl"`           // maps to Service.autoSsl and UpdateServiceRequest.autoSsl
	ConfigurationMode types.String `tfsdk:"configuration_mode"` // maps to Service.configurationMode and UpdateServiceRequest.configurationMode
	TLSProfile        types.String `tfsdk:"tls_profile"`        // from UpdateServiceRequest.tlsProfile
	DeliveryRegion    types.String `tfsdk:"delivery_region"`    // from UpdateServiceRequest.deliveryRegion

	// Computed/read-only fields - from Service struct
	Status    types.String `tfsdk:"status"`     // maps to Service.status
	CreatedAt types.String `tfsdk:"created_at"` // maps to Service.createdAt
	UpdatedAt types.String `tfsdk:"updated_at"` // maps to Service.updateAt (note: your SDK uses "updateAt", not "updatedAt")
}

// ServiceDataSourceModel describes the data source data model for reading CacheFly services.
// This is used when looking up existing services using Get() or GetByID() methods.
type ServiceDataSourceModel struct {
	// Lookup fields (one of these should be provided)
	ID         types.String `tfsdk:"id"`          // for GetByID()
	UniqueName types.String `tfsdk:"unique_name"` // for finding by unique name (via List then filter)

	// Options for Get() method
	ResponseType    types.String `tfsdk:"response_type"`    // maps to Get() responseType parameter
	IncludeFeatures types.Bool   `tfsdk:"include_features"` // maps to Get() includeFeatures parameter

	// Returned information - matches Service struct
	Name              types.String `tfsdk:"name"`
	AutoSSL           types.Bool   `tfsdk:"auto_ssl"`
	ConfigurationMode types.String `tfsdk:"configuration_mode"`
	Status            types.String `tfsdk:"status"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

// ServicesDataSourceModel describes the data source for listing multiple services.
// Maps to your SDK's List() method and ListServicesResponse.
type ServicesDataSourceModel struct {
	// Filter options - matches your SDK's ListOptions
	Status          types.String `tfsdk:"status"`           // maps to ListOptions.Status
	ResponseType    types.String `tfsdk:"response_type"`    // maps to ListOptions.ResponseType
	IncludeFeatures types.Bool   `tfsdk:"include_features"` // maps to ListOptions.IncludeFeatures

	// Pagination - matches your SDK's ListOptions
	Limit  types.Int64 `tfsdk:"limit"`  // maps to ListOptions.Limit
	Offset types.Int64 `tfsdk:"offset"` // maps to ListOptions.Offset

	// Results - matches your SDK's ListServicesResponse structure
	Services []ServiceListItem `tfsdk:"services"` // maps to ListServicesResponse.Services (data field)

	// Metadata - matches your SDK's MetaInfo
	Meta ServiceListMeta `tfsdk:"meta"` // maps to ListServicesResponse.Meta
}

// ServiceListItem represents a single service in a list of services.
// Matches your SDK's Service struct exactly.
type ServiceListItem struct {
	ID                types.String `tfsdk:"id"`                 // maps to Service._id
	Name              types.String `tfsdk:"name"`               // maps to Service.name
	UniqueName        types.String `tfsdk:"unique_name"`        // maps to Service.uniqueName
	AutoSSL           types.Bool   `tfsdk:"auto_ssl"`           // maps to Service.autoSsl
	ConfigurationMode types.String `tfsdk:"configuration_mode"` // maps to Service.configurationMode
	Status            types.String `tfsdk:"status"`             // maps to Service.status
	CreatedAt         types.String `tfsdk:"created_at"`         // maps to Service.createdAt
	UpdatedAt         types.String `tfsdk:"updated_at"`         // maps to Service.updateAt
}

// ServiceListMeta represents the metadata returned with service lists.
// Matches your SDK's MetaInfo struct exactly.
type ServiceListMeta struct {
	Limit  types.Int64 `tfsdk:"limit"`  // maps to MetaInfo.Limit
	Offset types.Int64 `tfsdk:"offset"` // maps to MetaInfo.Offset
	Count  types.Int64 `tfsdk:"count"`  // maps to MetaInfo.Count
}
