package resources

import (
	"context"
	"fmt"

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

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                   = &EdgeControlKVResource{}
	_ resource.ResourceWithConfigure      = &EdgeControlKVResource{}
	_ resource.ResourceWithImportState    = &EdgeControlKVResource{}
	_ resource.ResourceWithValidateConfig = &EdgeControlKVResource{}
)

// edgeControlAccountKVID is the ID of the account-level store.
const edgeControlAccountKVID = "account"

func NewEdgeControlKVResource() resource.Resource {
	return &EdgeControlKVResource{}
}

// EdgeControlKVResource manages the account-level or a service-level edge
// control key/value store.
type EdgeControlKVResource struct {
	client *cachefly.Client
}

func (r *EdgeControlKVResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_edge_control_kv"
}

func (r *EdgeControlKVResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an edge control key/value store read by edge control scripts: the account-level store, or the " +
			"store of one service when `service_id` is set. At the edge the two are merged, with service keys overriding " +
			"account keys. The whole store is replaced on every change, so keys that are not in `data` are removed, and " +
			"destroying the resource clears the store. Edge Control must be enabled for the account.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "`account` for the account-level store, otherwise the service ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_id": schema.StringAttribute{
				Description: "ID of the service whose store is managed. Leave unset to manage the account-level store.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data": schema.DynamicAttribute{
				MarkdownDescription: "Keys and values of the store, for example `{ region = \"eu\", max_age = 300, maintenance = false }`. " +
					"Values must be strings, numbers or booleans; nested objects and lists are not allowed. The account " +
					"limits the number of keys (10 by default), keys can be at most 64 characters and values at most 4096 bytes.",
				Required: true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the store was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "When the store was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *EdgeControlKVResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ValidateConfig rejects data that is not a flat object of scalar values.
func (r *EdgeControlKVResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data models.EdgeControlKVResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := models.EdgeControlKVDataToAPI(data.Data); err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("data"), "Invalid Edge Control KV Data", err.Error())
	}
}

func (r *EdgeControlKVResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.EdgeControlKVResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.replace(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error Creating CacheFly Edge Control KV", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EdgeControlKVResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.EdgeControlKVResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()

	var kv *api.EdgeControlKV
	var err error
	if serviceID == "" {
		kv, err = r.client.EdgeControlKV.GetAccount(ctx)
	} else {
		kv, err = r.client.EdgeControlKV.GetService(ctx, serviceID)
	}
	if err != nil {
		if serviceID != "" && isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading CacheFly Edge Control KV",
			fmt.Sprintf("Could not read the %s: %s", edgeControlKVScope(serviceID), err),
		)
		return
	}

	if err := mapEdgeControlKVToState(kv, &data); err != nil {
		resp.Diagnostics.AddError("Error Reading CacheFly Edge Control KV", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EdgeControlKVResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.EdgeControlKVResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.replace(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error Updating CacheFly Edge Control KV", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EdgeControlKVResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.EdgeControlKVResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()

	var err error
	if serviceID == "" {
		_, err = r.client.EdgeControlKV.ClearAccount(ctx)
	} else {
		_, err = r.client.EdgeControlKV.ClearService(ctx, serviceID)
	}
	if err != nil && !(serviceID != "" && isNotFoundError(err)) {
		resp.Diagnostics.AddError(
			"Error Clearing CacheFly Edge Control KV",
			fmt.Sprintf("Could not clear the %s: %s", edgeControlKVScope(serviceID), err),
		)
	}
}

// ImportState imports the account-level store with the ID "account", or a
// service-level store with the service ID.
func (r *EdgeControlKVResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be %q for the account-level store, or the service ID.", edgeControlAccountKVID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	if req.ID != edgeControlAccountKVID {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_id"), req.ID)...)
	}
}

// replace writes the planned data to the store and maps the result into data.
func (r *EdgeControlKVResource) replace(ctx context.Context, data *models.EdgeControlKVResourceModel) error {
	serviceID := data.ServiceID.ValueString()

	values, err := models.EdgeControlKVDataToAPI(data.Data)
	if err != nil {
		return fmt.Errorf("invalid data: %w", err)
	}

	var kv *api.EdgeControlKV
	if serviceID == "" {
		kv, err = r.client.EdgeControlKV.ReplaceAccount(ctx, values)
	} else {
		kv, err = r.client.EdgeControlKV.ReplaceService(ctx, serviceID, values)
	}
	if err != nil {
		return fmt.Errorf("could not replace the %s: %w", edgeControlKVScope(serviceID), err)
	}

	return mapEdgeControlKVToState(kv, data)
}

// mapEdgeControlKVToState copies kv into data. The current data value is kept
// when it holds the same content as the API's, because the API's JSON cannot
// tell objects from maps, and its numbers are float64.
func mapEdgeControlKVToState(kv *api.EdgeControlKV, data *models.EdgeControlKVResourceModel) error {
	serviceID := data.ServiceID.ValueString()
	if serviceID == "" {
		data.ID = types.StringValue(edgeControlAccountKVID)
	} else {
		data.ID = types.StringValue(serviceID)
	}
	data.CreatedAt = types.StringValue(kv.CreatedAt)
	data.UpdatedAt = types.StringValue(kv.UpdatedAt)

	if !data.Data.IsNull() && !data.Data.IsUnknown() {
		current, err := models.EdgeControlKVDataToAPI(data.Data)
		if err == nil && optionValuesEqual(current, kv.Data) {
			return nil
		}
	}

	value, err := models.EdgeControlKVDataFromAPI(kv.Data)
	if err != nil {
		return fmt.Errorf("could not convert the %s data: %w", edgeControlKVScope(serviceID), err)
	}
	data.Data = value
	return nil
}

func edgeControlKVScope(serviceID string) string {
	if serviceID == "" {
		return "account-level edge control KV store"
	}
	return fmt.Sprintf("edge control KV store of service %s", serviceID)
}
