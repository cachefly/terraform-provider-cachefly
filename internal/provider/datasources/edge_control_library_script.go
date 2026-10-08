package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &EdgeControlLibraryScriptDataSource{}
	_ datasource.DataSourceWithConfigure = &EdgeControlLibraryScriptDataSource{}
)

func NewEdgeControlLibraryScriptDataSource() datasource.DataSource {
	return &EdgeControlLibraryScriptDataSource{}
}

// EdgeControlLibraryScriptDataSource reads one edge control library script.
type EdgeControlLibraryScriptDataSource struct {
	client *cachefly.Client
}

func (d *EdgeControlLibraryScriptDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_edge_control_library_script"
}

func (d *EdgeControlLibraryScriptDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a script from the edge control script library: a `SYSTEM` script curated by CacheFly or one of " +
			"the account's own `USER` scripts. Use `code` to deploy it with `cachefly_edge_control_script`.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the library script.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the library script.",
				Computed:    true,
			},
			"kind": schema.StringAttribute{
				Description: "Kind of script (REQUEST or RESPONSE).",
				Computed:    true,
			},
			"code": schema.StringAttribute{
				Description: "JavaScript source of the script.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Ownership type of the script (USER or SYSTEM).",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the library script was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the library script was last updated.",
				Computed:    true,
			},
		},
	}
}

func (d *EdgeControlLibraryScriptDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EdgeControlLibraryScriptDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.EdgeControlLibraryScriptModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	script, err := d.client.EdgeControlLibrary.GetByID(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading CacheFly Edge Control Library Script",
			"Could not read library script ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	data = models.NewEdgeControlLibraryScriptModel(script)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
