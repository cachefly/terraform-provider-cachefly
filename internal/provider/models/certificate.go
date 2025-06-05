package models

import (
	"context"

	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CertificateModel represents the Terraform model for a CacheFly certificate
type CertificateModel struct {
	ID                types.String `tfsdk:"id"`
	Certificate       types.String `tfsdk:"certificate"`     //  certificate (required for create)
	CertificateKey    types.String `tfsdk:"certificate_key"` //  private key (required for create)
	Password          types.String `tfsdk:"password"`        // Optional password for key
	SubjectCommonName types.String `tfsdk:"subject_common_name"`
	SubjectNames      types.Set    `tfsdk:"subject_names"`
	Expired           types.Bool   `tfsdk:"expired"`
	Expiring          types.Bool   `tfsdk:"expiring"`
	InUse             types.Bool   `tfsdk:"in_use"`
	Managed           types.Bool   `tfsdk:"managed"`
	Services          types.Set    `tfsdk:"services"`
	Domains           types.Set    `tfsdk:"domains"`
	NotBefore         types.String `tfsdk:"not_before"`
	NotAfter          types.String `tfsdk:"not_after"`
	CreatedAt         types.String `tfsdk:"created_at"`
}

// ToSDKCreateRequest converts the Terraform model to SDK CreateCertificateRequest
func (m *CertificateModel) ToSDKCreateRequest(ctx context.Context) *api.CreateCertificateRequest {
	req := &api.CreateCertificateRequest{
		Certificate:    m.Certificate.ValueString(),
		CertificateKey: m.CertificateKey.ValueString(),
	}

	// Add password if provided
	if !m.Password.IsNull() && !m.Password.IsUnknown() && m.Password.ValueString() != "" {
		req.Password = m.Password.ValueString()
	}

	return req
}

// FromSDKCertificate converts an SDK Certificate to the Terraform model
func (m *CertificateModel) FromSDKCertificate(ctx context.Context, cert *api.Certificate) {
	m.ID = types.StringValue(cert.ID)
	m.SubjectCommonName = types.StringValue(cert.SubjectCommonName)
	m.Expired = types.BoolValue(cert.Expired)
	m.Expiring = types.BoolValue(cert.Expiring)
	m.InUse = types.BoolValue(cert.InUse)
	m.Managed = types.BoolValue(cert.Managed)
	m.NotBefore = types.StringValue(cert.NotBefore)
	m.NotAfter = types.StringValue(cert.NotAfter)
	m.CreatedAt = types.StringValue(cert.CreatedAt)

	// Convert SubjectNames slice to set
	if len(cert.SubjectNames) > 0 {
		subjectNameValues := make([]attr.Value, len(cert.SubjectNames))
		for i, name := range cert.SubjectNames {
			subjectNameValues[i] = types.StringValue(name)
		}
		m.SubjectNames = types.SetValueMust(types.StringType, subjectNameValues)
	} else {
		m.SubjectNames = types.SetValueMust(types.StringType, []attr.Value{})
	}

	// Convert Services slice to set
	if len(cert.Services) > 0 {
		serviceValues := make([]attr.Value, len(cert.Services))
		for i, service := range cert.Services {
			serviceValues[i] = types.StringValue(service)
		}
		m.Services = types.SetValueMust(types.StringType, serviceValues)
	} else {
		m.Services = types.SetValueMust(types.StringType, []attr.Value{})
	}

	// Convert Domains slice to set
	if len(cert.Domains) > 0 {
		domainValues := make([]attr.Value, len(cert.Domains))
		for i, domain := range cert.Domains {
			domainValues[i] = types.StringValue(domain)
		}
		m.Domains = types.SetValueMust(types.StringType, domainValues)
	} else {
		m.Domains = types.SetValueMust(types.StringType, []attr.Value{})
	}

	// Note: We don't populate Certificate and CertificateKey from the API response
	// as these contain sensitive data and are typically not returned by the API
	// They should remain as configured in the Terraform state
}
