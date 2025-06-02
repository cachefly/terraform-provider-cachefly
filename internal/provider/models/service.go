package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type ServiceResourceModel struct {
	// Core fields - matches your SDK's Service struct
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	UniqueName  types.String `tfsdk:"unique_name"`
	Description types.String `tfsdk:"description"`

	AutoSSL           types.Bool   `tfsdk:"auto_ssl"`
	ConfigurationMode types.String `tfsdk:"configuration_mode"`
	TLSProfile        types.String `tfsdk:"tls_profile"`
	DeliveryRegion    types.String `tfsdk:"delivery_region"`

	//read-only fields
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

type ServiceDataSourceModel struct {
	// Lookup fields (one of these should be provided)
	ID         types.String `tfsdk:"id"`
	UniqueName types.String `tfsdk:"unique_name"`

	// Options for Get() method
	ResponseType    types.String `tfsdk:"response_type"`
	IncludeFeatures types.Bool   `tfsdk:"include_features"`

	Name              types.String `tfsdk:"name"`
	AutoSSL           types.Bool   `tfsdk:"auto_ssl"`
	ConfigurationMode types.String `tfsdk:"configuration_mode"`
	Status            types.String `tfsdk:"status"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

type ServicesDataSourceModel struct {
	Status          types.String `tfsdk:"status"`
	ResponseType    types.String `tfsdk:"response_type"`
	IncludeFeatures types.Bool   `tfsdk:"include_features"`
	Limit           types.Int64  `tfsdk:"limit"`
	Offset          types.Int64  `tfsdk:"offset"` // maps to ListOptions.Offset

	Services []ServiceListItem `tfsdk:"services"` // maps to ListServicesResponse.Services (data field)
	Meta     ServiceListMeta   `tfsdk:"meta"`
}

type ServiceListItem struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	UniqueName        types.String `tfsdk:"unique_name"`
	AutoSSL           types.Bool   `tfsdk:"auto_ssl"`
	ConfigurationMode types.String `tfsdk:"configuration_mode"`
	Status            types.String `tfsdk:"status"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

type ServiceListMeta struct {
	Limit  types.Int64 `tfsdk:"limit"`  // maps to MetaInfo.Limit
	Offset types.Int64 `tfsdk:"offset"` // maps to MetaInfo.Offset
	Count  types.Int64 `tfsdk:"count"`  // maps to MetaInfo.Count
}
