// internal/provider/resources/service_domain.go
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
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/cachefly/cachefly-go-sdk/pkg/cachefly"
	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// satisfy framework interfaces.
var (
	_ resource.Resource                = &ServiceDomainResource{}
	_ resource.ResourceWithImportState = &ServiceDomainResource{}
)

func NewServiceDomainResource() resource.Resource {
	return &ServiceDomainResource{}
}

// ServiceDomainResource defines the resource implementation.
type ServiceDomainResource struct {
	client *cachefly.Client
}

func (r *ServiceDomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_domain"
}

func (r *ServiceDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CacheFly Service Domain resource. Manages custom domains attached to CacheFly services.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the service domain.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_id": schema.StringAttribute{
				Description: "The ID of the service this domain belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The domain name (e.g., 'example.com', 'cdn.example.com').",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Optional description for the domain.",
				Optional:    true,
				Computed:    true,
			},
			"validation_mode": schema.StringAttribute{
				Description: "Domain validation mode. Common values include 'DNS', 'HTTP', etc.",
				Optional:    true,
				Computed:    true,
			},
			"validation_target": schema.StringAttribute{
				Description: "The validation target (set by CacheFly during domain validation).",
				Computed:    true,
			},
			"validation_status": schema.StringAttribute{
				Description: "The current validation status of the domain.",
				Computed:    true,
			},
			"certificates": schema.ListAttribute{
				Description: "List of certificate IDs associated with this domain.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the domain was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the domain was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *ServiceDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServiceDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.ServiceDomainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the domain
	createReq := api.CreateServiceDomainRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
	}

	// we set validation_mode if provided
	if !data.ValidationMode.IsNull() && !data.ValidationMode.IsUnknown() {
		createReq.ValidationMode = data.ValidationMode.ValueString()
	}

	tflog.Debug(ctx, "Creating service domain", map[string]interface{}{
		"service_id": data.ServiceID.ValueString(),
		"name":       data.Name.ValueString(),
	})

	domain, err := r.client.ServiceDomains.Create(ctx, data.ServiceID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating CacheFly Service Domain",
			"Could not create service domain, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to state
	r.mapDomainToState(domain, &data)

	tflog.Debug(ctx, "Service domain created successfully", map[string]interface{}{
		"domain_id": domain.ID,
		"name":      domain.Name,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.ServiceDomainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading service domain", map[string]interface{}{
		"service_id": data.ServiceID.ValueString(),
		"domain_id":  data.ID.ValueString(),
	})

	domain, err := r.client.ServiceDomains.GetByID(ctx, data.ServiceID.ValueString(), data.ID.ValueString(), "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading CacheFly Service Domain",
			"Could not read service domain ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to state
	r.mapDomainToState(domain, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.ServiceDomainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the domain
	updateReq := api.UpdateServiceDomainRequest{}

	// Only include fields that have values
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		updateReq.Name = data.Name.ValueString()
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		updateReq.Description = data.Description.ValueString()
	}
	if !data.ValidationMode.IsNull() && !data.ValidationMode.IsUnknown() {
		updateReq.ValidationMode = data.ValidationMode.ValueString()
	}

	tflog.Debug(ctx, "Updating service domain", map[string]interface{}{
		"service_id": data.ServiceID.ValueString(),
		"domain_id":  data.ID.ValueString(),
	})

	domain, err := r.client.ServiceDomains.UpdateByID(ctx, data.ServiceID.ValueString(), data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating CacheFly Service Domain",
			"Could not update service domain, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to state
	r.mapDomainToState(domain, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.ServiceDomainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting service domain", map[string]interface{}{
		"service_id": data.ServiceID.ValueString(),
		"domain_id":  data.ID.ValueString(),
	})

	err := r.client.ServiceDomains.DeleteByID(ctx, data.ServiceID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting CacheFly Service Domain",
			"Could not delete service domain, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Service domain deleted successfully")
}

func (r *ServiceDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "service_id:domain_id"
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in format 'service_id:domain_id'",
		)
		return
	}

	serviceID := parts[0]
	domainID := parts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_id"), serviceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), domainID)...)
}

func (r *ServiceDomainResource) mapDomainToState(domain *api.ServiceDomain, data *models.ServiceDomainResourceModel) {
	data.ID = types.StringValue(domain.ID)
	data.ServiceID = types.StringValue(domain.Service)
	data.Name = types.StringValue(domain.Name)
	data.Description = types.StringValue(domain.Description)
	data.ValidationMode = types.StringValue(domain.ValidationMode)
	data.ValidationTarget = types.StringValue(domain.ValidationTarget)
	data.ValidationStatus = types.StringValue(domain.ValidationStatus)
	data.CreatedAt = types.StringValue(domain.CreatedAt)
	data.UpdatedAt = types.StringValue(domain.UpdatedAt)

	// Convert certificates slice to Terraform list
	if len(domain.Certificates) > 0 {
		certElements := make([]attr.Value, len(domain.Certificates))
		for i, cert := range domain.Certificates {
			certElements[i] = types.StringValue(cert)
		}
		data.Certificates, _ = types.ListValue(types.StringType, certElements)
	} else {
		data.Certificates = types.ListNull(types.StringType)
	}
}
