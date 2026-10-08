package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly"
	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource                   = &EdgeControlKVDataSource{}
	_ datasource.DataSourceWithConfigure      = &EdgeControlKVDataSource{}
	_ datasource.DataSourceWithValidateConfig = &EdgeControlKVDataSource{}
)

func NewEdgeControlKVDataSource() datasource.DataSource {
	return &EdgeControlKVDataSource{}
}

// EdgeControlKVDataSource reads an edge control key/value store.
type EdgeControlKVDataSource struct {
	client *cachefly.Client
}

func (d *EdgeControlKVDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_edge_control_kv"
}

func (d *EdgeControlKVDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an edge control key/value store: the account-level store, the store of one service, or, with " +
			"`merged`, the account store merged with the service store as the service's edge scripts see it.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "`account`, the service ID, or `service_id:merged`.",
				Computed:            true,
			},
			"service_id": schema.StringAttribute{
				Description: "ID of the service whose store is read. Leave unset to read the account-level store.",
				Optional:    true,
			},
			"merged": schema.BoolAttribute{
				MarkdownDescription: "Return the account store merged with the service store; service keys override account keys. Requires `service_id`.",
				Optional:            true,
			},
			"data": schema.DynamicAttribute{
				Description: "Keys and values of the store.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "When the store was created. Not set when `merged` is true.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "When the store was last updated. Not set when `merged` is true.",
				Computed:            true,
			},
		},
	}
}

func (d *EdgeControlKVDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*cachefly.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *cachefly.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// ValidateConfig requires service_id when merged is true.
func (d *EdgeControlKVDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data models.EdgeControlKVDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Merged.ValueBool() && data.ServiceID.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("merged"),
			"Missing service_id",
			"merged = true requires service_id: the merged view combines the account store with the store of a service.",
		)
	}
}

func (d *EdgeControlKVDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.EdgeControlKVDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()

	var values map[string]interface{}
	switch {
	case data.Merged.ValueBool():
		if serviceID == "" {
			resp.Diagnostics.AddAttributeError(path.Root("service_id"), "Missing service_id", "merged = true requires service_id.")
			return
		}
		merged, err := d.client.EdgeControlKV.GetMerged(ctx, serviceID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading CacheFly Edge Control KV",
				fmt.Sprintf("Could not read the merged edge control KV store of service %s: %s", serviceID, err),
			)
			return
		}
		values = merged
		data.ID = types.StringValue(serviceID + ":merged")
		data.CreatedAt = types.StringNull()
		data.UpdatedAt = types.StringNull()

	default:
		var kv *api.EdgeControlKV
		var err error
		if serviceID == "" {
			kv, err = d.client.EdgeControlKV.GetAccount(ctx)
			data.ID = types.StringValue("account")
		} else {
			kv, err = d.client.EdgeControlKV.GetService(ctx, serviceID)
			data.ID = types.StringValue(serviceID)
		}
		if err != nil {
			resp.Diagnostics.AddError("Error Reading CacheFly Edge Control KV", "Could not read the edge control KV store: "+err.Error())
			return
		}
		values = kv.Data
		data.CreatedAt = types.StringValue(kv.CreatedAt)
		data.UpdatedAt = types.StringValue(kv.UpdatedAt)
	}

	value, err := models.EdgeControlKVDataFromAPI(values)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading CacheFly Edge Control KV", err.Error())
		return
	}
	data.Data = value

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
