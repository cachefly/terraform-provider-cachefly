package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/cachefly/cachefly-go-sdk/pkg/cachefly"
	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// satisfy framework interfaces.
var (
	_ resource.Resource              = &ServiceOptionsResource{}
	_ resource.ResourceWithConfigure = &ServiceOptionsResource{}
)

func NewServiceOptionsResource() resource.Resource {
	return &ServiceOptionsResource{}
}

// ServiceOptionsResource defines the resource implementation.
type ServiceOptionsResource struct {
	client *cachefly.Client
}

func (r *ServiceOptionsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_options"
}

func (r *ServiceOptionsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages service options configuration for a CacheFly service using the new options API.",
		Attributes: map[string]schema.Attribute{
			"service_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the service to configure options for.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"options": schema.DynamicAttribute{
				MarkdownDescription: "Service options configuration as key-value pairs. Each option follows the enabled/value structure for feature options.",
				Optional:            true,
				Computed:            true,
			},
			"last_updated": schema.StringAttribute{
				MarkdownDescription: "Timestamp of the last update.",
				Computed:            true,
			},
		},
	}
}

func (r *ServiceOptionsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServiceOptionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.ServiceOptionsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()

	tflog.Debug(ctx, "Creating service options", map[string]interface{}{
		"service_id": serviceID,
	})

	// Convert Terraform model to API ServiceOptions
	apiOptions, err := data.ToAPIServiceOptions()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Converting Service Options",
			"Could not convert service options to API format: "+err.Error(),
		)
		return
	}

	// Update service options via new API
	_, err = r.client.ServiceOptions.UpdateOptions(ctx, serviceID, apiOptions)
	if err != nil {
		// Check if it's a validation error
		if validationErr, ok := err.(api.ServiceOptionsValidationError); ok {
			resp.Diagnostics.AddError(
				"Service Options Validation Failed",
				fmt.Sprintf("Validation failed: %s. Details: %v", validationErr.Message, validationErr.Errors),
			)
			return
		}

		resp.Diagnostics.AddError(
			"Error Creating CacheFly Service Options",
			"Could not create service options, unexpected error: "+err.Error(),
		)
		return
	}

	// Instead of reading all options from API, preserve what we sent
	// This ensures consistency between plan and state
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	tflog.Debug(ctx, "Service options created successfully", map[string]interface{}{
		"service_id": serviceID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceOptionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.ServiceOptionsModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()
	tflog.Debug(ctx, "Reading service options", map[string]interface{}{
		"service_id": serviceID,
	})

	// Get all service options from new API
	allOpts, err := r.client.ServiceOptions.GetOptions(ctx, serviceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading CacheFly Service Options",
			"Could not read service options for service ID "+serviceID+": "+err.Error(),
		)
		return
	}

	// Only keep the options that we're managing in Terraform
	managedOpts, err := r.extractManagedOptions(data, allOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Extracting Managed Options",
			"Could not extract managed options: "+err.Error(),
		)
		return
	}

	// Convert filtered API response back to Terraform model
	if err := data.FromAPIServiceOptions(serviceID, managedOpts); err != nil {
		resp.Diagnostics.AddError(
			"Error Converting API Response",
			"Could not convert API response to Terraform model: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper method to extract only the options we're managing
func (r *ServiceOptionsResource) extractManagedOptions(data models.ServiceOptionsModel, allOpts api.ServiceOptions) (api.ServiceOptions, error) {
	if data.Options.IsNull() || data.Options.IsUnknown() {
		return api.ServiceOptions{}, nil
	}

	managedOpts := make(api.ServiceOptions)

	// Get the options we're managing from the current state
	underlyingValue := data.Options.UnderlyingValue()
	if objValue, ok := underlyingValue.(basetypes.ObjectValue); ok {
		attributes := objValue.Attributes()

		// Only include options that we're managing
		for key := range attributes {
			if value, exists := allOpts[key]; exists {
				managedOpts[key] = value
			}
		}
	}

	return managedOpts, nil
}

func (r *ServiceOptionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.ServiceOptionsModel
	var state models.ServiceOptionsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()

	tflog.Debug(ctx, "Updating service options", map[string]interface{}{
		"service_id": serviceID,
	})

	// Convert Terraform model to API ServiceOptions
	apiOptions, err := data.ToAPIServiceOptions()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Converting Service Options",
			"Could not convert service options to API format: "+err.Error(),
		)
		return
	}

	// Update service options via new API with validation
	_, err = r.client.ServiceOptions.UpdateOptions(ctx, serviceID, apiOptions)
	if err != nil {
		// Check if it's a validation error
		if validationErr, ok := err.(api.ServiceOptionsValidationError); ok {
			resp.Diagnostics.AddError(
				"Service Options Validation Failed",
				fmt.Sprintf("Validation failed: %s. Details: %v", validationErr.Message, validationErr.Errors),
			)
			return
		}

		resp.Diagnostics.AddError(
			"Error Updating CacheFly Service Options",
			"Could not update service options, unexpected error: "+err.Error(),
		)
		return
	}

	// Use the planned data directly to ensure consistency
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	tflog.Debug(ctx, "Service options updated successfully", map[string]interface{}{
		"service_id": serviceID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceOptionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.ServiceOptionsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()
	tflog.Debug(ctx, "Deleting service options (resetting to defaults)", map[string]interface{}{
		"service_id": serviceID,
	})

	// Get the current options from the API to see what's actually set
	currentOpts, err := r.client.ServiceOptions.GetOptions(ctx, serviceID)
	if err != nil {
		tflog.Warn(ctx, "Could not read current options before delete", map[string]interface{}{
			"service_id": serviceID,
			"error":      err.Error(),
		})
		return
	}

	tflog.Debug(ctx, "Current options to potentially reset", map[string]interface{}{
		"current_options": currentOpts,
	})

	// Reset options in groups to avoid validation conflicts
	resetGroups := []map[string]interface{}{
		// Group 1: Security keys (reset these first)
		{
			"protectServeKeyEnabled": false,
			"apiKeyEnabled":          false,
		},

		// Group 2: Simple boolean options
		{},

		// Group 3: Complex objects with enabled/value structure
		{},

		// Group 4: Arrays
		{
			"expiryHeaders": []interface{}{},
		},
	}

	// Build Group 2: Simple boolean options (only if they exist in current options)
	booleanOptions := []string{
		"nocache", "allowretry", "servestale", "normalizequerystring",
		"forceorigqstring", "cors", "autoRedirect", "livestreaming",
		"linkpreheat", "purgenoquery", "cachebygeocountry", "cachebyreferer",
		"cachebyregion", "send-xff", "brotli_support", "use_slicer",
	}

	for _, option := range booleanOptions {
		if _, exists := currentOpts[option]; exists {
			resetGroups[1][option] = false
		}
	}

	// Build Group 3: Complex objects (only if they exist in current options)
	complexOptions := []string{
		"reverseProxy", "error_ttl", "ttfb_timeout", "contimeout", "maxcons",
		"bwthrottle", "sharedshield", "originhostheader", "purgemode",
		"dirpurgeskip", "httpmethods", "skip_pserve_ext", "skip_encoding_ext",
		"redirect",
	}

	for _, option := range complexOptions {
		if _, exists := currentOpts[option]; exists {
			resetGroups[2][option] = map[string]interface{}{
				"enabled": false,
			}
		}
	}

	// Apply resets in groups
	for i, resetGroup := range resetGroups {
		if len(resetGroup) == 0 {
			continue
		}

		tflog.Debug(ctx, "Applying reset group", map[string]interface{}{
			"group":   i + 1,
			"options": resetGroup,
		})

		_, err := r.client.ServiceOptions.UpdateOptions(ctx, serviceID, resetGroup)
		if err != nil {
			tflog.Warn(ctx, "Reset group failed, trying individual options", map[string]interface{}{
				"group": i + 1,
				"error": err.Error(),
			})

			// Try individual options in this group
			for optName, optValue := range resetGroup {
				individualReset := api.ServiceOptions{optName: optValue}
				_, err := r.client.ServiceOptions.UpdateOptions(ctx, serviceID, individualReset)
				if err != nil {
					tflog.Warn(ctx, "Failed to reset individual option", map[string]interface{}{
						"option": optName,
						"error":  err.Error(),
					})
				}
			}
		}
	}

	tflog.Debug(ctx, "Service options reset completed")
}

func (r *ServiceOptionsResource) DeleteOptional(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.ServiceOptionsModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()

	tflog.Debug(ctx, "Deleting service options (resetting to defaults)", map[string]interface{}{
		"service_id": serviceID,
	})

	// First, get the current options to see what needs to be disabled
	currentOpts, err := r.client.ServiceOptions.GetOptions(ctx, serviceID)
	if err != nil {
		tflog.Warn(ctx, "Could not read current options before delete", map[string]interface{}{
			"service_id": serviceID,
			"error":      err.Error(),
		})
		// Continue with reset anyway
	}

	// Handle special cases like ProtectServe key deletion
	if currentOpts != nil {
		// Check if reverseProxy is enabled and disable it
		if reverseProxyVal, exists := currentOpts["reverseProxy"]; exists {
			if reverseProxyMap, ok := reverseProxyVal.(map[string]interface{}); ok {
				if enabled, ok := reverseProxyMap["enabled"].(bool); ok && enabled {
					tflog.Debug(ctx, "Disabling reverse proxy before reset")
					// Disable reverse proxy
					disabledReverseProxy := map[string]interface{}{
						"enabled": false,
					}
					resetOpts := api.ServiceOptions{
						"reverseProxy": disabledReverseProxy,
					}
					_, err := r.client.ServiceOptions.UpdateOptions(ctx, serviceID, resetOpts)
					if err != nil {
						tflog.Warn(ctx, "Failed to disable reverse proxy", map[string]interface{}{
							"error": err.Error(),
						})
					}
				}
			}
		}

		// Check other complex options and reset them (todo: (awet) need to be dynamic, use metadata or allowed options)
		featureOptions := []string{
			"error_ttl", "ttfb_timeout", "contimeout", "maxcons",
			"bwthrottle", "sharedshield", "originhostheader", "purgemode",
			"dirpurgeskip", "httpmethods", "skip_pserve_ext", "skip_encoding_ext",
			"redirect",
		}

		resetOpts := make(api.ServiceOptions)
		for _, optionName := range featureOptions {
			if _, exists := currentOpts[optionName]; exists {
				resetOpts[optionName] = map[string]interface{}{
					"enabled": false,
				}
			}
		}

		if len(resetOpts) > 0 {
			tflog.Debug(ctx, "Resetting feature options to disabled state", map[string]interface{}{
				"options_count": len(resetOpts),
			})

			_, err := r.client.ServiceOptions.UpdateOptions(ctx, serviceID, resetOpts)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Resetting Service Options",
					"Could not reset service options to defaults, unexpected error: "+err.Error(),
				)
				return
			}
		}
	}

	tflog.Debug(ctx, "Service options reset to defaults successfully")

	// Note: In Terraform, the Delete operation should remove the resource from state
	// The service itself and its default options remain, but Terraform no longer manages them
}

// Helper function to get available option names for a service
func (r *ServiceOptionsResource) getAvailableOptions(ctx context.Context, serviceID string) ([]string, error) {
	return r.client.ServiceOptions.GetAvailableOptionNames(ctx, serviceID)
}

// Helper function to check if an option is available
func (r *ServiceOptionsResource) isOptionAvailable(ctx context.Context, serviceID, optionName string) (bool, error) {
	available, _, err := r.client.ServiceOptions.IsOptionAvailable(ctx, serviceID, optionName)
	return available, err
}
