// internal/provider/datasources/delivery_regions.go
package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cachefly/cachefly-go-sdk/pkg/cachefly"
	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DeliveryRegionsDataSource{}

// NewDeliveryRegionsDataSource creates a new data source instance.
func NewDeliveryRegionsDataSource() datasource.DataSource {
	return &DeliveryRegionsDataSource{}
}

// DeliveryRegionsDataSource implements the data source for listing delivery regions.
type DeliveryRegionsDataSource struct {
	client *cachefly.Client
}

// deliveryRegionsDataSourceModel represents query inputs and results.
// Kept local to this file since the shape is simple.
type deliveryRegionsDataSourceModel struct {
	Offset types.Int64 `tfsdk:"offset"`
	Limit  types.Int64 `tfsdk:"limit"`

	// Results
	Regions types.List `tfsdk:"regions"`
}

func (d *DeliveryRegionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_delivery_regions"
}

func (d *DeliveryRegionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CacheFly Delivery Regions data source. List available delivery regions.",

		Attributes: map[string]schema.Attribute{
			"offset": schema.Int64Attribute{
				Description: "Offset for pagination (default: 0).",
				Optional:    true,
			},
			"limit": schema.Int64Attribute{
				Description: "Limit for pagination (default: API default).",
				Optional:    true,
			},
			"regions": schema.ListNestedAttribute{
				Description: "List of delivery regions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the delivery region.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the delivery region.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the delivery region.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *DeliveryRegionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DeliveryRegionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data deliveryRegionsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build list options
	opts := api.ListDeliveryRegionsOptions{}

	if !data.Offset.IsNull() {
		opts.Offset = int(data.Offset.ValueInt64())
	}
	if !data.Limit.IsNull() {
		opts.Limit = int(data.Limit.ValueInt64())
	}
	// Set a sane default page size if not provided
	if opts.Limit <= 0 {
		opts.Limit = 100
	}

	// Fetch and accumulate all pages
	var allRegions []api.DeliveryRegion
	for {
		pageResp, err := d.client.DeliveryRegions.List(ctx, opts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading CacheFly Delivery Regions",
				"Could not read delivery regions: "+err.Error(),
			)
			return
		}

		allRegions = append(allRegions, pageResp.Regions...)

		fetched := len(pageResp.Regions)
		total := pageResp.Meta.Count
		opts.Offset += fetched

		if fetched == 0 || (total > 0 && opts.Offset >= total) {
			break
		}
	}

	// Prepare attribute type map for each object in the list
	objectAttrTypes := map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"description": types.StringType,
	}

	// Map response to Terraform values
	items := make([]attr.Value, len(allRegions))
	for i, region := range allRegions {
		obj, _ := types.ObjectValue(
			objectAttrTypes,
			map[string]attr.Value{
				"id":          types.StringValue(region.ID),
				"name":        types.StringValue(region.Name),
				"description": types.StringValue(region.Description),
			},
		)
		items[i] = obj
	}

	listValue, diags := types.ListValue(
		types.ObjectType{AttrTypes: objectAttrTypes},
		items,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Regions = listValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
