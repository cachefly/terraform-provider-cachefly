package provider

import (
	"context"

	"github.com/avvvet/cachefly-sdk-go/pkg/cachefly"
	"github.com/avvvet/cachefly-sdk-go/pkg/cachefly/api"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &serviceResource{}
)

// serviceResource is the resource implementation.
type serviceResource struct {
	client *cachefly.Client
}

// serviceResourceModel maps the resource schema data.
type serviceResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	UniqueName        types.String `tfsdk:"unique_name"`
	Description       types.String `tfsdk:"description"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
	CreatedAt         types.String `tfsdk:"created_at"`
	AutoSSL           types.Bool   `tfsdk:"auto_ssl"`
	ConfigurationMode types.String `tfsdk:"configuration_mode"`
	Status            types.String `tfsdk:"status"`
}

// Metadata returns the resource type name.
func (r *serviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

// Schema defines the schema for the resource.
func (r *serviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CacheFly service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the service.",
			},
			"unique_name": schema.StringAttribute{
				Required:    true,
				Description: "The unique name of the service.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "The description of the service.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the service was last updated.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the service was created.",
			},
			"auto_ssl": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether auto SSL is enabled for the service.",
			},
			"configuration_mode": schema.StringAttribute{
				Computed:    true,
				Description: "The configuration mode of the service (e.g., 'MANUAL').",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the service (e.g., 'ACTIVE').",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *serviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare the create service request
	createRequest := api.CreateServiceRequest{
		Name:        plan.Name.ValueString(),
		UniqueName:  plan.UniqueName.ValueString(),
		Description: plan.Description.ValueString(),
	}

	// Call the CacheFly API to create the service
	created, err := r.client.Services.Create(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError("API Error", "Failed to create service: "+err.Error())
		return
	}

	// Map response to state
	plan.ID = types.StringValue(created.ID)
	plan.Name = types.StringValue(created.Name)
	plan.UniqueName = types.StringValue(created.UniqueName)
	plan.UpdatedAt = types.StringValue(created.UpdatedAt)
	plan.CreatedAt = types.StringValue(created.CreatedAt)
	plan.AutoSSL = types.BoolValue(created.AutoSSL)
	plan.ConfigurationMode = types.StringValue(created.ConfigurationMode)
	plan.Status = types.StringValue(created.Status)

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
// This is the Read method, which fetches the service using client.Services.Get.
func (r *serviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the service by ID using Get
	service, err := r.client.Services.Get(ctx, state.ID.ValueString(), "shallow", false)
	if err != nil {
		resp.Diagnostics.AddError("API Error", "Failed to get service: "+err.Error())
		return
	}

	// Check if service is nil (e.g., deleted externally)
	if service == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with latest data
	state.Name = types.StringValue(service.Name)
	state.UniqueName = types.StringValue(service.UniqueName)
	state.Description = types.StringNull() // Description not returned by Get
	state.UpdatedAt = types.StringValue(service.UpdatedAt)
	state.CreatedAt = types.StringValue(service.CreatedAt)
	state.AutoSSL = types.BoolValue(service.AutoSSL)
	state.ConfigurationMode = types.StringValue(service.ConfigurationMode)
	state.Status = types.StringValue(service.Status)

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state.
func (r *serviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"The CacheFly SDK does not support updating services. Consider deleting and recreating the service.",
	)
}

// Delete deletes the resource and removes the Terraform state.
func (r *serviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"Delete Not Supported",
		"The CacheFly SDK does not support deleting services.",
	)
}
