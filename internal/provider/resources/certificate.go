package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly"
	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &CertificateResource{}
	_ resource.ResourceWithImportState = &CertificateResource{}
)

// NewCertificateResource is a helper function to simplify the provider implementation
func NewCertificateResource() resource.Resource {
	return &CertificateResource{}
}

// CertificateResource defines the resource implementation
type CertificateResource struct {
	client *cachefly.Client
}

// Metadata returns the resource type name
func (r *CertificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

// Schema defines the schema for the resource
func (r *CertificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CacheFly Certificate resource. Manages TLS/SSL certificates for CacheFly services.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the certificate.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"certificate": schema.StringAttribute{
				Description: "PEM-encoded certificate content. Required for certificate creation.",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"certificate_key": schema.StringAttribute{
				Description: "PEM-encoded private key for the certificate. Required for certificate creation.",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Description: "Optional password for the private key if it's encrypted.",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// Computed attributes from the API
			"subject_common_name": schema.StringAttribute{
				Description: "The common name (CN) from the certificate's subject.",
				Computed:    true,
			},
			"subject_names": schema.SetAttribute{
				Description: "All subject names from the certificate (including CN and SAN).",
				ElementType: types.StringType,
				Computed:    true,
			},
			"expired": schema.BoolAttribute{
				Description: "Whether the certificate has expired.",
				Computed:    true,
			},
			"expiring": schema.BoolAttribute{
				Description: "Whether the certificate is expiring soon.",
				Computed:    true,
			},
			"in_use": schema.BoolAttribute{
				Description: "Whether the certificate is currently in use by services.",
				Computed:    true,
			},
			"managed": schema.BoolAttribute{
				Description: "Whether this is a CacheFly-managed certificate.",
				Computed:    true,
			},
			"services": schema.SetAttribute{
				Description: "List of service IDs using this certificate.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"domains": schema.SetAttribute{
				Description: "List of domains covered by this certificate.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"not_before": schema.StringAttribute{
				Description: "Certificate validity start date (ISO 8601 format).",
				Computed:    true,
			},
			"not_after": schema.StringAttribute{
				Description: "Certificate validity end date (ISO 8601 format).",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the certificate was uploaded to CacheFly.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *CertificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*cachefly.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *cachefly.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state
func (r *CertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.CertificateModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := api.CreateCertificateRequest{
		Certificate:    data.Certificate.ValueString(),
		CertificateKey: data.CertificateKey.ValueString(),
	}

	if !data.Password.IsNull() && !data.Password.IsUnknown() && data.Password.ValueString() != "" {
		createReq.Password = data.Password.ValueString()
	}

	cert, err := r.client.Certificates.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating CacheFly Certificate",
			"Could not create certificate, unexpected error: "+err.Error(),
		)
		return
	}

	r.mapCertificateToState(cert, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data
func (r *CertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.CertificateModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	certID := data.ID.ValueString()

	cert, err := r.client.Certificates.GetByID(ctx, certID, "")
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Error Reading CacheFly Certificate",
				"Could not read certificate with ID "+certID+": "+err.Error(),
			)
		}
		return
	}

	// Preserve the sensitive input data from state
	existingCertificate := data.Certificate
	existingCertificateKey := data.CertificateKey
	existingPassword := data.Password

	// Map response to state
	r.mapCertificateToState(cert, &data)

	// Restore sensitive input data
	data.Certificate = existingCertificate
	data.CertificateKey = existingCertificateKey
	data.Password = existingPassword

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource - certificates are immutable, so this mainly handles drift
func (r *CertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.CertificateModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Certificates are immutable in CacheFly - any change requires replacement
	// This should not be called due to RequiresReplace plan modifiers
	resp.Diagnostics.AddError(
		"Certificate Update Not Supported",
		"Certificates cannot be updated. Any changes to certificate content require replacement.",
	)
}

// Delete deletes the resource
func (r *CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.CertificateModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	certID := data.ID.ValueString()

	err := r.client.Certificates.Delete(ctx, certID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting CacheFly Certificate",
			"Could not delete certificate with ID "+certID+": "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform state
func (r *CertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// When importing, we cannot recover the original certificate and key content
	// as these are not returned by the API for security reasons
	resp.Diagnostics.AddWarning(
		"Certificate Content Not Available",
		"When importing a certificate, the original certificate and private key content cannot be retrieved from the API. "+
			"You will need to manually set these values in your Terraform configuration to match the imported certificate.",
	)
}

// Helper function to map SDK Certificate to Terraform state
func (r *CertificateResource) mapCertificateToState(cert *api.Certificate, data *models.CertificateModel) {
	data.ID = types.StringValue(cert.ID)
	data.SubjectCommonName = types.StringValue(cert.SubjectCommonName)
	data.Expired = types.BoolValue(cert.Expired)
	data.Expiring = types.BoolValue(cert.Expiring)
	data.InUse = types.BoolValue(cert.InUse)
	data.Managed = types.BoolValue(cert.Managed)
	data.NotBefore = types.StringValue(cert.NotBefore)
	data.NotAfter = types.StringValue(cert.NotAfter)
	data.CreatedAt = types.StringValue(cert.CreatedAt)

	// Convert SubjectNames slice to set
	if len(cert.SubjectNames) > 0 {
		subjectNameValues := make([]attr.Value, len(cert.SubjectNames))
		for i, name := range cert.SubjectNames {
			subjectNameValues[i] = types.StringValue(name)
		}
		data.SubjectNames = types.SetValueMust(types.StringType, subjectNameValues)
	} else {
		data.SubjectNames = types.SetValueMust(types.StringType, []attr.Value{})
	}

	// Convert Services slice to set
	if len(cert.Services) > 0 {
		serviceValues := make([]attr.Value, len(cert.Services))
		for i, service := range cert.Services {
			serviceValues[i] = types.StringValue(service)
		}
		data.Services = types.SetValueMust(types.StringType, serviceValues)
	} else {
		data.Services = types.SetValueMust(types.StringType, []attr.Value{})
	}

	// Convert Domains slice to set
	if len(cert.Domains) > 0 {
		domainValues := make([]attr.Value, len(cert.Domains))
		for i, domain := range cert.Domains {
			domainValues[i] = types.StringValue(domain)
		}
		data.Domains = types.SetValueMust(types.StringType, domainValues)
	} else {
		data.Domains = types.SetValueMust(types.StringType, []attr.Value{})
	}

}
