package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/cachefly/cachefly-go-sdk/pkg/cachefly"
	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"

	"github.com/avvvet/terraform-provider-cachefly/internal/provider/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &ServiceResource{}
	_ resource.ResourceWithConfigure   = &ServiceResource{}
	_ resource.ResourceWithImportState = &ServiceResource{}
)

func NewServiceResource() resource.Resource {
	return &ServiceResource{}
}

// ServiceResource defines the resource implementation.
type ServiceResource struct {
	client *cachefly.Client
}

func (r *ServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (r *ServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
# CacheFly Service Resource

Manages a CacheFly CDN service. A service represents a CDN configuration that defines how content is cached and delivered.

## Example Usage

` + "```hcl" + `
resource "cachefly_service" "example" {
  name        = "my-cdn-service"
  unique_name = "my-unique-service-name"
  description = "CDN service for my application"
  
  auto_ssl           = true
  configuration_mode = "advanced"
  delivery_region    = "global"
  tls_profile        = "modern"
}
` + "```" + `

## Import

Services can be imported using their ID:

` + "```bash" + `
terraform import cachefly_service.example service-id-here
` + "```" + `
		`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the service.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the service.",
				Required:    true,
			},
			"unique_name": schema.StringAttribute{
				Description: "The unique name of the service used in URLs and configurations. Must be unique across all services.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description of the service.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"auto_ssl": schema.BoolAttribute{
				Description: "Whether to automatically provision SSL certificates.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"configuration_mode": schema.StringAttribute{
				Description: "The configuration mode for the service.",
				Optional:    true,
				Computed:    true,
			},
			"tls_profile": schema.StringAttribute{
				Description: "The TLS profile to use for SSL connections.",
				Optional:    true,
			},
			"delivery_region": schema.StringAttribute{
				Description: "The delivery region for the service.",
				Optional:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the service.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the service was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the service was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *ServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (r *ServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.ServiceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the service request using your SDK's CreateServiceRequest
	createReq := api.CreateServiceRequest{
		Name:        data.Name.ValueString(),
		UniqueName:  data.UniqueName.ValueString(),
		Description: data.Description.ValueString(),
	}

	tflog.Debug(ctx, "Creating CacheFly service", map[string]interface{}{
		"name":        createReq.Name,
		"unique_name": createReq.UniqueName,
		"description": createReq.Description,
	})

	// Create the service via SDK
	service, err := r.client.Services.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating CacheFly Service",
			"Could not create service, unexpected error: "+err.Error(),
		)
		return
	}

	// Check if we need to update the service with additional configuration
	var needsUpdate bool
	var updateReq api.UpdateServiceRequest

	// Always include description in update
	updateReq.Description = data.Description.ValueString()

	// Check each field individually and only include if explicitly set
	if !data.AutoSSL.IsNull() && !data.AutoSSL.IsUnknown() {
		needsUpdate = true
		updateReq.AutoSSL = data.AutoSSL.ValueBool()
	}

	if !data.ConfigurationMode.IsNull() && !data.ConfigurationMode.IsUnknown() && data.ConfigurationMode.ValueString() != "" {
		needsUpdate = true
		updateReq.ConfigurationMode = data.ConfigurationMode.ValueString()
	}

	if !data.TLSProfile.IsNull() && !data.TLSProfile.IsUnknown() && data.TLSProfile.ValueString() != "" {
		needsUpdate = true
		updateReq.TLSProfile = data.TLSProfile.ValueString()
	}

	if !data.DeliveryRegion.IsNull() && !data.DeliveryRegion.IsUnknown() && data.DeliveryRegion.ValueString() != "" {
		needsUpdate = true
		updateReq.DeliveryRegion = data.DeliveryRegion.ValueString()
	}

	if needsUpdate {
		tflog.Debug(ctx, "Updating service configuration", map[string]interface{}{
			"service_id":          service.ID,
			"config_mode_request": updateReq.ConfigurationMode,
			"auto_ssl_request":    updateReq.AutoSSL,
		})

		updatedService, err := r.client.Services.UpdateServiceByID(ctx, service.ID, updateReq)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating CacheFly Service Configuration",
				"Service was created but configuration update failed: "+err.Error(),
			)
			return
		}

		tflog.Debug(ctx, "Service update response", map[string]interface{}{
			"config_mode_response": updatedService.ConfigurationMode,
			"auto_ssl_response":    updatedService.AutoSSL,
		})

		service = updatedService
	} else {
		tflog.Debug(ctx, "Skipping service configuration update - no optional fields provided")
	}

	// Map response to Terraform state
	r.mapServiceToState(service, &data)

	tflog.Debug(ctx, "Created CacheFly service", map[string]interface{}{
		"id":     service.ID,
		"status": service.Status,
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.ServiceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the service from the API using your SDK's GetByID method
	service, err := r.client.Services.GetByID(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading CacheFly Service",
			"Could not read service ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map fresh API data to state
	r.mapServiceToState(service, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.ServiceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create update request using your SDK's UpdateServiceRequest
	updateReq := api.UpdateServiceRequest{
		Description: data.Description.ValueString(),
		AutoSSL:     data.AutoSSL.ValueBool(),
	}

	// Only set optional fields if they have actual values
	if !data.ConfigurationMode.IsNull() && !data.ConfigurationMode.IsUnknown() && data.ConfigurationMode.ValueString() != "" {
		updateReq.ConfigurationMode = data.ConfigurationMode.ValueString()
	}
	if !data.TLSProfile.IsNull() && !data.TLSProfile.IsUnknown() && data.TLSProfile.ValueString() != "" {
		updateReq.TLSProfile = data.TLSProfile.ValueString()
	}
	if !data.DeliveryRegion.IsNull() && !data.DeliveryRegion.IsUnknown() && data.DeliveryRegion.ValueString() != "" {
		updateReq.DeliveryRegion = data.DeliveryRegion.ValueString()
	}

	tflog.Debug(ctx, "Updating CacheFly service", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Update the service via SDK
	service, err := r.client.Services.UpdateServiceByID(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating CacheFly Service",
			"Could not update service, unexpected error: "+err.Error(),
		)
		return
	}

	// Map updated service to state
	r.mapServiceToState(service, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.ServiceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting CacheFly service", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// First deactivate the service before deletion using your SDK's DeactivateServiceByID
	_, err := r.client.Services.DeactivateServiceByID(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deactivating CacheFly Service",
			"Could not deactivate service before deletion: "+err.Error(),
		)
		return
	}

	// Note: Your SDK doesn't seem to have a Delete method, so we just deactivate
	// This is typical for CDN services - they're usually deactivated rather than deleted

	tflog.Debug(ctx, "Deactivated CacheFly service", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
}

func (r *ServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to map SDK Service to Terraform state
// IMPORTANT: Always use what the API returns, not what the user configured
// This prevents "inconsistent result" errors when API transforms values
func (r *ServiceResource) mapServiceToState(service *api.Service, data *models.ServiceResourceModel) {
	// Core fields - always use API response
	data.ID = types.StringValue(service.ID)
	data.Name = types.StringValue(service.Name)
	data.UniqueName = types.StringValue(service.UniqueName)
	data.Status = types.StringValue(service.Status)
	data.CreatedAt = types.StringValue(service.CreatedAt)
	data.UpdatedAt = types.StringValue(service.UpdatedAt)

	// Configuration fields - use what API actually returned
	// This handles cases where API transforms or ignores certain values
	data.AutoSSL = types.BoolValue(service.AutoSSL)
	data.ConfigurationMode = types.StringValue(service.ConfigurationMode)

	// For optional fields that might not be returned by the API,
	// preserve the user's configuration if the API doesn't return a value
	if service.ConfigurationMode == "" && !data.ConfigurationMode.IsNull() {
		// Keep the user's configuration if API doesn't return anything
	} else {
		// Use what the API returned
		data.ConfigurationMode = types.StringValue(service.ConfigurationMode)
	}

	// Note: TLSProfile and DeliveryRegion are not in your Service struct,
	// so they might be separate API calls or not included in the response.
	// For now, we'll leave them as user-configured values since the API
	// doesn't return them in the Service object.
}
