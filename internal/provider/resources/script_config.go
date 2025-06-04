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

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &ScriptConfigResource{}
	_ resource.ResourceWithImportState = &ScriptConfigResource{}
)

// NewScriptConfigResource is a helper function to simplify the provider implementation
func NewScriptConfigResource() resource.Resource {
	return &ScriptConfigResource{}
}

// ScriptConfigResource defines the resource implementation
type ScriptConfigResource struct {
	client *cachefly.Client
}

// Metadata returns the resource type name
func (r *ScriptConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_script_config"
}

// Schema defines the schema for the resource
func (r *ScriptConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CacheFly Script Config resource. Manages script configurations based on global script config definitions.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the script config.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the script config.",
				Required:    true,
			},
			"services": schema.SetAttribute{
				Description: "Set of service IDs this script config applies to.",
				ElementType: types.StringType,
				Required:    true,
			},
			"script_config_definition": schema.StringAttribute{
				Description: "ID of the global script config definition to use.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mime_type": schema.StringAttribute{
				Description: "MIME type of the script config value.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("application/json"),
			},
			"value": schema.StringAttribute{
				Description: "Configuration value as JSON string. Leave empty for initial creation, then configure via updates.",
				Optional:    true,
			},
			"activated": schema.BoolAttribute{
				Description: "Whether the script config should be activated. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			// Computed attributes
			"purpose": schema.StringAttribute{
				Description: "Purpose of the script config definition.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the script config was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the script config was last updated.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *ScriptConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ScriptConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.ScriptConfigModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to SDK create request
	createReq := data.ToSDKCreateRequest(ctx)

	tflog.Debug(ctx, "Creating script config", map[string]interface{}{
		"name":                     createReq.Name,
		"script_config_definition": createReq.ScriptConfigDefinition,
		"mime_type":                createReq.MimeType,
		"services_count":           len(createReq.Services),
	})

	// Create script config via API
	config, err := r.client.ScriptConfigs.Create(ctx, *createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating CacheFly Script Config",
			"Could not create script config, unexpected error: "+err.Error(),
		)
		return
	}

	// Activate the script config if requested (default is true)
	shouldActivate := !data.Activated.IsNull() && !data.Activated.IsUnknown() && data.Activated.ValueBool()
	if shouldActivate {
		tflog.Debug(ctx, "Activating script config", map[string]interface{}{
			"config_id": config.ID,
		})

		activatedConfig, err := r.client.ScriptConfigs.ActivateByID(ctx, config.ID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Activating CacheFly Script Config",
				"Script config was created but could not be activated: "+err.Error(),
			)
			return
		}
		config = activatedConfig
	}

	// Map response to state
	r.mapScriptConfigToState(config, &data)

	// Set activation status in state
	data.Activated = types.BoolValue(shouldActivate)

	tflog.Debug(ctx, "Script config created successfully", map[string]interface{}{
		"config_id": config.ID,
		"name":      config.Name,
		"purpose":   config.Purpose,
		"activated": shouldActivate,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create creates the resource and sets the initial Terraform state
func (r *ScriptConfigResource) CreateOld(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.ScriptConfigModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to SDK create request
	createReq := data.ToSDKCreateRequest(ctx)

	tflog.Debug(ctx, "Creating script config", map[string]interface{}{
		"name":                     createReq.Name,
		"script_config_definition": createReq.ScriptConfigDefinition,
		"mime_type":                createReq.MimeType,
		"services_count":           len(createReq.Services),
	})

	// Create script config via API
	config, err := r.client.ScriptConfigs.Create(ctx, *createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating CacheFly Script Config",
			"Could not create script config, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to state
	r.mapScriptConfigToState(config, &data)

	tflog.Debug(ctx, "Script config created successfully", map[string]interface{}{
		"config_id": config.ID,
		"name":      config.Name,
		"purpose":   config.Purpose,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data
func (r *ScriptConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.ScriptConfigModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configID := data.ID.ValueString()

	tflog.Debug(ctx, "Reading script config", map[string]interface{}{
		"config_id": configID,
	})

	config, err := r.client.ScriptConfigs.GetByID(ctx, configID, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading CacheFly Script Config",
			"Could not read script config with ID "+configID+": "+err.Error(),
		)
		return
	}

	// Map response to state
	r.mapScriptConfigToState(config, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScriptConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.ScriptConfigModel
	var state models.ScriptConfigModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configID := data.ID.ValueString()

	// Convert to SDK update request
	updateReq := data.ToSDKUpdateRequest(ctx)

	tflog.Debug(ctx, "Updating script config", map[string]interface{}{
		"config_id": configID,
		"name":      updateReq.Name,
	})

	// Update script config via API
	config, err := r.client.ScriptConfigs.UpdateByID(ctx, configID, *updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating CacheFly Script Config",
			"Could not update script config with ID "+configID+": "+err.Error(),
		)
		return
	}

	// Handle activation status changes
	newActivated := data.Activated.ValueBool()
	oldActivated := state.Activated.ValueBool()

	if newActivated != oldActivated {
		if newActivated {
			tflog.Debug(ctx, "Activating script config", map[string]interface{}{
				"config_id": configID,
			})
			config, err = r.client.ScriptConfigs.ActivateByID(ctx, configID)
		} else {
			tflog.Debug(ctx, "Deactivating script config", map[string]interface{}{
				"config_id": configID,
			})
			config, err = r.client.ScriptConfigs.DeactivateByID(ctx, configID)
		}

		if err != nil {
			action := "activate"
			if !newActivated {
				action = "deactivate"
			}
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error %sing CacheFly Script Config", action),
				fmt.Sprintf("Could not %s script config with ID %s: %s", action, configID, err.Error()),
			)
			return
		}
	}

	// Map response to state
	r.mapScriptConfigToState(config, &data)
	data.Activated = types.BoolValue(newActivated)

	tflog.Debug(ctx, "Script config updated successfully", map[string]interface{}{
		"config_id": config.ID,
		"name":      config.Name,
		"activated": newActivated,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success
func (r *ScriptConfigResource) UpdateOld(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.ScriptConfigModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configID := data.ID.ValueString()

	// Convert to SDK update request
	updateReq := data.ToSDKUpdateRequest(ctx)

	tflog.Debug(ctx, "Updating script config", map[string]interface{}{
		"config_id": configID,
		"name":      updateReq.Name,
	})

	// Update script config via API
	config, err := r.client.ScriptConfigs.UpdateByID(ctx, configID, *updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating CacheFly Script Config",
			"Could not update script config with ID "+configID+": "+err.Error(),
		)
		return
	}

	// Map response to state
	r.mapScriptConfigToState(config, &data)

	tflog.Debug(ctx, "Script config updated successfully", map[string]interface{}{
		"config_id": config.ID,
		"name":      config.Name,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource by deactivating it
func (r *ScriptConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.ScriptConfigModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configID := data.ID.ValueString()

	tflog.Debug(ctx, "Deactivating script config", map[string]interface{}{
		"config_id": configID,
	})

	// Deactivate script config (this is the "delete" operation)
	_, err := r.client.ScriptConfigs.DeactivateByID(ctx, configID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deactivating CacheFly Script Config",
			"Could not deactivate script config with ID "+configID+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Script config deactivated successfully", map[string]interface{}{
		"config_id": configID,
	})
}

// ImportState imports an existing resource into Terraform state
func (r *ScriptConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to map SDK ScriptConfig to Terraform state
func (r *ScriptConfigResource) mapScriptConfigToState(config *api.ScriptConfig, data *models.ScriptConfigModel) {
	data.FromSDKScriptConfig(context.Background(), config)
}
