package models

import (
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
