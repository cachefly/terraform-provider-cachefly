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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly"
	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// satisfy framework interfaces.
var (
	_ resource.Resource                = &ServiceResource{}
	_ resource.ResourceWithConfigure   = &ServiceResource{}
	_ resource.ResourceWithImportState = &ServiceResource{}
	_ resource.ResourceWithModifyPlan  = &ServiceResource{}
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
		MarkdownDescription: "A service represents a CDN configuration that defines how content is cached and delivered.",
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			},
			"configuration_mode": schema.StringAttribute{
				Description: "The configuration mode for the service.",
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
			"options": schema.DynamicAttribute{
				MarkdownDescription: `Service options as a map.`,
				Description:         "Service options configuration as key-value pairs. Each option follows the enabled/value structure for feature options.",
				Optional:            true,
				// Computed:    true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the service. Set this to 'ACTIVE' to activate the service or 'DEACTIVATED' to deactivate it.",
				Description:         "The current status of the service.",
				Computed:            true,
				Optional:            true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the service was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := api.CreateServiceRequest{
		Name:        data.Name.ValueString(),
		UniqueName:  data.UniqueName.ValueString(),
		Description: data.Description.ValueString(),
	}

	service, err := r.client.Services.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating CacheFly Service",
			"Could not create service, unexpected error: "+err.Error(),
		)
		return
	}

	var needsUpdate bool
	var updateReq api.UpdateServiceRequest

	if !data.AutoSSL.IsNull() && !data.AutoSSL.IsUnknown() {
		needsUpdate = true
		updateReq.AutoSSL = data.AutoSSL.ValueBool()
	}

	if !data.TLSProfile.IsNull() {
		needsUpdate = true
		updateReq.TLSProfile = data.TLSProfile.ValueString()
	}

	if !data.DeliveryRegion.IsNull() {
		needsUpdate = true
		updateReq.DeliveryRegion = data.DeliveryRegion.ValueString()
	}

	if needsUpdate {
		updatedService, err := r.client.Services.UpdateServiceByID(ctx, service.ID, updateReq)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating CacheFly Service Configuration",
				"Service was created but configuration update failed: "+err.Error(),
			)
			return
		}

		service = updatedService
	}

	if data.Status.ValueString() == "DEACTIVATED" {
		service, err = r.client.Services.DeactivateServiceByID(ctx, service.ID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Deactivating CacheFly Service",
				"Could not deactivate service: "+err.Error(),
			)
			return
		}
	}

	if !data.Options.IsNull() && !data.Options.IsUnknown() {
		serviceOptions, err := data.ToAPIServiceOptions()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error converting service options",
				"Could not convert service options: "+err.Error(),
			)
			return
		}

		_, err = r.client.ServiceOptions.UpdateOptions(ctx, service.ID, serviceOptions)
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
				"Error updating service options",
				"Could not update service options: "+err.Error(),
			)
			return
		}
	}

	r.mapServiceToState(service, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.ServiceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := r.client.Services.GetByID(ctx, data.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Error Reading CacheFly Service",
				"Could not read service ID "+data.ID.ValueString()+": "+err.Error(),
			)
		}
		return
	}

	// Handle service options based on whether they are configured or imported
	if err := r.handleServiceOptionsRead(ctx, &data); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Service Options",
			"Could not read service options: "+err.Error(),
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
	var state models.ServiceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := api.UpdateServiceRequest{}

	if !data.Description.Equal(state.Description) {
		updateReq.Description = data.Description.ValueString()
	}

	if !data.AutoSSL.Equal(state.AutoSSL) {
		updateReq.AutoSSL = data.AutoSSL.ValueBool()
	}

	if !data.TLSProfile.Equal(state.TLSProfile) {
		updateReq.TLSProfile = data.TLSProfile.ValueString()
	}

	if !data.DeliveryRegion.Equal(state.DeliveryRegion) {
		updateReq.DeliveryRegion = data.DeliveryRegion.ValueString()
	}

	service, err := r.client.Services.UpdateServiceByID(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating CacheFly Service",
			"Could not update service, unexpected error: "+err.Error(),
		)
		return
	}

	if !data.Status.Equal(state.Status) {
		if data.Status.ValueString() == "ACTIVE" {
			service, err = r.client.Services.ActivateServiceByID(ctx, data.ID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Activating CacheFly Service",
					"Could not activate service: "+err.Error(),
				)
				return
			}
		} else {
			service, err = r.client.Services.DeactivateServiceByID(ctx, data.ID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Deactivating CacheFly Service",
					"Could not deactivate service: "+err.Error(),
				)
				return
			}
		}
	}

	r.mapServiceToState(service, &data)

	if !data.Options.Equal(state.Options) {
		// Get current state to compare with planned changes
		var currentState models.ServiceResourceModel
		resp.Diagnostics.Append(req.State.Get(ctx, &currentState)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Convert both current and planned options to API format for comparison
		currentOptions, err := currentState.ToAPIServiceOptions()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Converting Current Service Options",
				"Could not convert current service options: "+err.Error(),
			)
			return
		}

		plannedOptions, err := data.ToAPIServiceOptions()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Converting Planned Service Options",
				"Could not convert planned service options: "+err.Error(),
			)
			return
		}

		// Compare and create a map of only changed options
		changedOptions := make(api.ServiceOptions)

		// Check each planned option against current state
		for key, plannedValue := range plannedOptions {
			currentValue, exists := currentOptions[key]

			// Include option if it's new or value has changed
			if !exists || !r.compareOptionValues(currentValue, plannedValue) {
				changedOptions[key] = plannedValue
			}
		}

		if len(changedOptions) > 0 {
			_, err = r.client.ServiceOptions.UpdateOptions(ctx, data.ID.ValueString(), changedOptions)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Updating CacheFly Service Options",
					"Could not update service options: "+err.Error(),
				)
				return
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.ServiceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Services.DeactivateServiceByID(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deactivating CacheFly Service",
			"Could not deactivate service before deletion: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.AddWarning(
		"Resource deactivated, not deleted",
		"The backing service was deactivated because hard delete is not supported by the API.",
	)
}

func (r *ServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ModifyPlan suppresses no-op updates caused by type-only differences in the
// dynamic "options" attribute (e.g. a config-declared list(string) versus a
// tuple stored in state by older provider versions). Terraform
// explicitly permit a planned value that conflicts with configuration when it
// exactly matches the prior state, as a signal that the two are functionally
// equivalent; returning the prior state wholesale turns the plan into a no-op
// so no phantom update (and no server-side updatedAt bump) ever happens.
func (r *ServiceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Only in-place updates are eligible; skip create and destroy plans.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan, state models.ServiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Practitioner-owned attributes are never marked unknown by the
	// framework; an unknown here comes from unresolved config expressions,
	// i.e. a genuine pending change that must not be suppressed.
	if plan.Options.IsUnknown() || plan.TLSProfile.IsUnknown() || plan.DeliveryRegion.IsUnknown() {
		return
	}
	if !plan.TLSProfile.Equal(state.TLSProfile) || !plan.DeliveryRegion.Equal(state.DeliveryRegion) {
		return
	}
	if !plan.Name.Equal(state.Name) || !plan.UniqueName.Equal(state.UniqueName) {
		return
	}

	// Computed (and optional+computed) attributes: an unknown value is a
	// framework marking artifact of the suspected change, not a change by
	// itself; a known value that differs from state is a real change.
	for _, pair := range []struct{ planV, stateV attr.Value }{
		{plan.ID, state.ID},
		{plan.Description, state.Description},
		{plan.AutoSSL, state.AutoSSL},
		{plan.ConfigurationMode, state.ConfigurationMode},
		{plan.Status, state.Status},
		{plan.CreatedAt, state.CreatedAt},
		{plan.UpdatedAt, state.UpdatedAt},
	} {
		if pair.planV.IsUnknown() {
			continue
		}
		if !pair.planV.Equal(pair.stateV) {
			return
		}
	}

	// Options: semantic comparison. Same values in a different dynamic type
	// representation (list vs tuple, int vs float) are not a change. Nested
	// unknowns are dropped by ToAPIServiceOptions and therefore show up as a
	// difference, which correctly prevents suppression.
	planOptions, err := plan.ToAPIServiceOptions()
	if err != nil {
		return
	}
	stateOptions, err := state.ToAPIServiceOptions()
	if err != nil {
		return
	}
	if plan.Options.IsNull() != state.Options.IsNull() {
		return
	}
	if !optionValuesEqual(map[string]interface{}(planOptions), map[string]interface{}(stateOptions)) {
		return
	}

	// Nothing genuinely changes: return the prior state as the plan, which
	// core recognizes as a no-op.
	resp.Plan.Raw = req.State.Raw.Copy()
}

func (r *ServiceResource) handleServiceOptionsRead(ctx context.Context, data *models.ServiceResourceModel) error {
	serviceID := data.ID.ValueString()
	// if data.Options.IsNull() || data.Options.IsUnknown() {
	// 	if r.isPostImportRead(data) {
	// 		allOptions, err := r.client.ServiceOptions.GetOptions(ctx, serviceID)
	// 		if err != nil {
	// 			return fmt.Errorf("could not read service options for imported service: %w", err)
	// 		}

	// 		if err := r.setOptionsFromAPI(data, allOptions); err != nil {
	// 			return fmt.Errorf("could not convert service options: %w", err)
	// 		}

	// 		tflog.Debug(ctx, "Loaded all service options for imported resource", map[string]interface{}{
	// 			"service_id":    serviceID,

	// 			"options_count": len(allOptions),
	// 		})
	// 	}
	// 	return nil
	// }

	currentOptions, err := data.ToAPIServiceOptions()
	if err != nil {
		return fmt.Errorf("could not convert current options: %w", err)
	}

	allApiOptions, err := r.client.ServiceOptions.GetOptions(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("could not read current service options: %w", err)
	}

	managedOptions := make(api.ServiceOptions)
	for key := range currentOptions {
		if apiValue, exists := allApiOptions[key]; exists {
			// Apply recursive filtering to only include configured nested fields
			managedOptions[key] = r.filterNestedOptions(currentOptions[key], apiValue)
		}
	}

	// When the API reports no semantic change, keep the state value untouched.
	//
	// Rebuilding the dynamic options value from API JSON guesses types that
	// the config may have declared differently (JSON arrays carry no
	// list/tuple distinction and are rebuilt as tuples, while the config may
	// type them as lists). Such a type-only difference is invisible in
	// rendered plans but makes Terraform/OpenTofu report phantom drift on
	// every refresh and plan no-op in-place updates. The config's typing is
	// authoritative for this Optional attribute, so Read must preserve the
	// stored representation whenever the values are semantically equal.
	if optionValuesEqual(map[string]interface{}(currentOptions), map[string]interface{}(managedOptions)) {
		return nil
	}

	if err := r.setOptionsFromAPI(data, managedOptions); err != nil {
		return fmt.Errorf("could not convert managed options: %w", err)
	}

	return nil
}

// optionValuesEqual reports whether two option values represent the same
// data. Unlike compareOptionValues, it compares numbers numerically: values
// read back from state carry Go int (see ToAPIServiceOptions), while values
// decoded from API JSON carry float64, and those must compare as equal.
func optionValuesEqual(a, b interface{}) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}

	if aMap, ok := a.(map[string]interface{}); ok {
		bMap, ok := b.(map[string]interface{})
		if !ok || len(aMap) != len(bMap) {
			return false
		}
		for key, aVal := range aMap {
			bVal, exists := bMap[key]
			if !exists || !optionValuesEqual(aVal, bVal) {
				return false
			}
		}
		return true
	}

	if aSlice, ok := a.([]interface{}); ok {
		bSlice, ok := b.([]interface{})
		if !ok || len(aSlice) != len(bSlice) {
			return false
		}
		for i := range aSlice {
			if !optionValuesEqual(aSlice[i], bSlice[i]) {
				return false
			}
		}
		return true
	}

	if aNum, aOk := numericValue(a); aOk {
		bNum, bOk := numericValue(b)
		return bOk && aNum == bNum
	}

	return a == b
}

func numericValue(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case float64:
		return n, true
	}
	return 0, false
}

// isPostImportRead detects if this is a read immediately after import
// by checking if we have minimal state (indicating fresh import)
func (r *ServiceResource) isPostImportRead(data *models.ServiceResourceModel) bool {
	// After import, we typically only have ID set and other fields may be in default state
	// We can check if computed fields like Status, CreatedAt are not set yet
	return data.Status.IsNull() || data.CreatedAt.IsNull()
}

// setOptionsFromAPI converts API ServiceOptions directly to the ServiceModel's Options field
func (r *ServiceResource) setOptionsFromAPI(data *models.ServiceResourceModel, options api.ServiceOptions) error {
	// Convert map[string]interface{} to types.Dynamic
	if len(options) > 0 {
		// Create a map[string]attr.Value for the dynamic type
		elements := make(map[string]attr.Value)
		attrTypes := make(map[string]attr.Type)

		for key, value := range options {
			// Convert interface{} to appropriate types.Value based on type
			convertedValue, attrType := convertInterfaceToAttrValue(value)
			elements[key] = convertedValue
			attrTypes[key] = attrType
		}

		// Create the object value
		objValue, diags := types.ObjectValue(attrTypes, elements)
		if diags.HasError() {
			return fmt.Errorf("failed to convert options to object: %v", diags.Errors())
		}

		data.Options = types.DynamicValue(objValue)
	}

	return nil
}

// filterNestedOptions recursively filters API options to only include fields that were configured
func (r *ServiceResource) filterNestedOptions(current, api interface{}) interface{} {
	currentMap, currentIsMap := current.(map[string]interface{})
	apiMap, apiIsMap := api.(map[string]interface{})

	if currentIsMap && apiIsMap {
		filtered := make(map[string]interface{})
		for key := range currentMap {
			if apiValue, exists := apiMap[key]; exists {
				// Recursively filter nested objects
				filtered[key] = r.filterNestedOptions(currentMap[key], apiValue)
			}
		}
		return filtered
	}

	// For non-map values, just return the API value
	return api
}

// convertInterfaceToAttrValue converts interface{} to attr.Value and attr.Type
func convertInterfaceToAttrValue(value interface{}) (attr.Value, attr.Type) {
	switch v := value.(type) {
	case string:
		return types.StringValue(v), types.StringType
	case bool:
		return types.BoolValue(v), types.BoolType
	case int:
		return types.Int64Value(int64(v)), types.Int64Type
	case int64:
		return types.Int64Value(v), types.Int64Type
	case float64:
		return types.Float64Value(v), types.Float64Type
	case map[string]interface{}:
		nestedElements := make(map[string]attr.Value)
		nestedAttrTypes := make(map[string]attr.Type)

		for nestedKey, nestedValue := range v {
			nestedAttrValue, nestedAttrType := convertInterfaceToAttrValue(nestedValue)
			nestedElements[nestedKey] = nestedAttrValue
			nestedAttrTypes[nestedKey] = nestedAttrType
		}

		objValue, _ := types.ObjectValue(nestedAttrTypes, nestedElements)
		return objValue, types.ObjectType{AttrTypes: nestedAttrTypes}
	case []interface{}:
		if len(v) == 0 {
			tupleValue, _ := types.TupleValue([]attr.Type{}, []attr.Value{})
			return tupleValue, types.TupleType{ElemTypes: []attr.Type{}}
		}

		tupleElements := make([]attr.Value, len(v))
		elemTypes := make([]attr.Type, len(v))

		for i, item := range v {
			itemValue, itemType := convertInterfaceToAttrValue(item)
			tupleElements[i] = itemValue
			elemTypes[i] = itemType
		}

		tupleValue, _ := types.TupleValue(elemTypes, tupleElements)
		return tupleValue, types.TupleType{ElemTypes: elemTypes}
	default:
		return types.StringValue(fmt.Sprintf("%v", v)), types.StringType
	}
}

// compareOptionValues compares two option values to determine if they are equal
// This handles the various data types that service options can contain
func (r *ServiceResource) compareOptionValues(current, planned interface{}) bool {
	// Handle nil cases
	if current == nil && planned == nil {
		return true
	}
	if current == nil || planned == nil {
		return false
	}

	// Handle maps (complex options like reverseProxy, redirect, etc.)
	if currentMap, ok := current.(map[string]interface{}); ok {
		if plannedMap, ok := planned.(map[string]interface{}); ok {
			// Compare maps recursively
			if len(currentMap) != len(plannedMap) {
				return false
			}
			for key, currentVal := range currentMap {
				plannedVal, exists := plannedMap[key]
				if !exists || !r.compareOptionValues(currentVal, plannedVal) {
					return false
				}
			}
			return true
		}
		return false
	}

	// Handle slices/arrays
	if currentSlice, ok := current.([]interface{}); ok {
		if plannedSlice, ok := planned.([]interface{}); ok {
			if len(currentSlice) != len(plannedSlice) {
				return false
			}
			for i, currentVal := range currentSlice {
				if !r.compareOptionValues(currentVal, plannedSlice[i]) {
					return false
				}
			}
			return true
		}
		return false
	}

	// Handle primitive types (string, bool, int, float)
	return current == planned
}

// IMPORTANT: to use what the API returns, not what the user configured
func (r *ServiceResource) mapServiceToState(service *api.Service, data *models.ServiceResourceModel) {
	// Core fields - always use API response
	data.ID = types.StringValue(service.ID)
	data.Name = types.StringValue(service.Name)
	data.UniqueName = types.StringValue(service.UniqueName)
	data.Description = types.StringValue(service.Description)
	data.Status = types.StringValue(service.Status)
	data.CreatedAt = types.StringValue(service.CreatedAt)
	data.UpdatedAt = types.StringValue(service.UpdatedAt)

	data.AutoSSL = types.BoolValue(service.AutoSSL)
	data.ConfigurationMode = types.StringValue(service.ConfigurationMode)

	if service.ConfigurationMode == "" && !data.ConfigurationMode.IsNull() {
		// Keep the user's configuration if API doesn't return anything
	} else {
		// Use what the API returned
		data.ConfigurationMode = types.StringValue(service.ConfigurationMode)
	}

	// todo: (awet) TLSProfile and DeliveryRegion ,
}
