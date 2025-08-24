package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly"
	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

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
				Optional:    true,
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
			"purpose": schema.StringAttribute{
				Description: "Purpose of the script config definition.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the script config.",
				Computed:    true,
			},
			"use_schema": schema.BoolAttribute{
				Description: "Whether to use the schema for the script config value.",
				Computed:    true,
			},
			"data_model": schema.StringAttribute{
				Description: "Data model of the script config.",
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

	createReq := data.ToSDKCreateRequest(ctx)

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

	config, err := r.client.ScriptConfigs.GetByID(ctx, configID, "")
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Error Reading CacheFly Script Config",
				"Could not read script config with ID "+configID+": "+err.Error(),
			)
		}
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

	updateReq := &api.UpdateScriptConfigRequest{}

	if !data.Name.Equal(state.Name) {
		updateReq.Name = data.Name.ValueString()
	}
	if !data.MimeType.Equal(state.MimeType) {
		updateReq.MimeType = data.MimeType.ValueString()
	}
	if !data.ScriptConfigDefinition.Equal(state.ScriptConfigDefinition) {
		updateReq.ScriptConfigDefinition = data.ScriptConfigDefinition.ValueString()
	}

	if !data.Services.Equal(state.Services) {
		var services []string
		serviceElements := make([]types.String, 0, len(data.Services.Elements()))
		data.Services.ElementsAs(ctx, &serviceElements, false)
		for _, elem := range serviceElements {
			services = append(services, elem.ValueString())
		}
		updateReq.Services = services
	}

	if !data.Value.Equal(state.Value) {
		valueStr := data.Value.ValueString()
		if valueStr != "" {
			var value interface{}
			if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
				// If JSON parsing fails, use as string
				value = valueStr
			}
			updateReq.Value = value
		}
	}

	config, err := r.client.ScriptConfigs.UpdateByID(ctx, configID, *updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating CacheFly Script Config",
			"Could not update script config with ID "+configID+": "+err.Error(),
		)
		return
	}

	if !data.Activated.Equal(state.Activated) {
		if data.Activated.ValueBool() {
			config, err = r.client.ScriptConfigs.ActivateByID(ctx, configID)
		} else {
			config, err = r.client.ScriptConfigs.DeactivateByID(ctx, configID)
		}

		if err != nil {
			action := "activat"
			if !data.Activated.ValueBool() {
				action = "deactivat"
			}
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error %sing CacheFly Script Config", action),
				fmt.Sprintf("Could not %s script config with ID %s: %s", action, configID, err.Error()),
			)
			return
		}
	}

	r.mapScriptConfigToState(config, &data)

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

	// Deactivate script config (there is no "delete" operation)
	_, err := r.client.ScriptConfigs.DeactivateByID(ctx, configID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deactivating CacheFly Script Config",
			"Could not deactivate script config with ID "+configID+": "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform state
func (r *ScriptConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to map SDK ScriptConfig to Terraform state
func (r *ScriptConfigResource) mapScriptConfigToState(config *api.ScriptConfig, data *models.ScriptConfigModel) {
	data.ID = types.StringValue(config.ID)
	data.Name = types.StringValue(config.Name)
	data.ScriptConfigDefinition = types.StringValue(config.ScriptConfigDefinition)
	data.MimeType = types.StringValue(config.MimeType)
	data.Purpose = types.StringValue(config.Purpose)
	data.CreatedAt = types.StringValue(config.CreatedAt)
	data.UpdatedAt = types.StringValue(config.UpdatedAt)
	data.Status = types.StringValue(config.Status)
	data.UseSchema = types.BoolValue(config.UseSchema)
	data.DataModel = types.StringValue(config.DataModel)

	// Convert Services slice to set
	if len(config.Services) > 0 {
		serviceValues := make([]attr.Value, len(config.Services))
		for i, service := range config.Services {
			serviceValues[i] = types.StringValue(service)
		}
		data.Services = types.SetValueMust(types.StringType, serviceValues)
	} else {
		data.Services = types.SetValueMust(types.StringType, []attr.Value{})
	}

	// Convert Value interface{} to JSON string
	if config.Value != nil {
		valueBytes, err := json.Marshal(config.Value)
		if err != nil {
			// If marshaling fails, try to convert to string
			if str, ok := config.Value.(string); ok {
				data.Value = types.StringValue(str)
			} else {
				data.Value = types.StringNull()
			}
		} else {
			data.Value = types.StringValue(string(valueBytes))
		}
	} else {
		data.Value = types.StringNull()
	}
}
