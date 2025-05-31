// internal/provider/models/service_domain.go
package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// ServiceDomainResourceModel represents the Terraform resource model for cachefly_service_domain
type ServiceDomainResourceModel struct {
	ID               types.String `tfsdk:"id"`
	ServiceID        types.String `tfsdk:"service_id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	ValidationMode   types.String `tfsdk:"validation_mode"`
	ValidationTarget types.String `tfsdk:"validation_target"`
	ValidationStatus types.String `tfsdk:"validation_status"`
	Certificates     types.List   `tfsdk:"certificates"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

// ServiceDomainDataSourceModel represents the Terraform data source model for cachefly_service_domain
type ServiceDomainDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	ServiceID        types.String `tfsdk:"service_id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	ValidationMode   types.String `tfsdk:"validation_mode"`
	ValidationTarget types.String `tfsdk:"validation_target"`
	ValidationStatus types.String `tfsdk:"validation_status"`
	Certificates     types.List   `tfsdk:"certificates"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`

	// Optional query parameters
	ResponseType types.String `tfsdk:"response_type"`
}

// ServiceDomainsDataSourceModel represents the data source for listing multiple domains
type ServiceDomainsDataSourceModel struct {
	ServiceID    types.String `tfsdk:"service_id"`
	Search       types.String `tfsdk:"search"`
	Offset       types.Int64  `tfsdk:"offset"`
	Limit        types.Int64  `tfsdk:"limit"`
	ResponseType types.String `tfsdk:"response_type"`

	// Results
	Domains types.List `tfsdk:"domains"`
}
