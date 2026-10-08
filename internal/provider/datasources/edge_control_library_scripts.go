package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly"
	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &EdgeControlLibraryScriptsDataSource{}
	_ datasource.DataSourceWithConfigure = &EdgeControlLibraryScriptsDataSource{}
)

func NewEdgeControlLibraryScriptsDataSource() datasource.DataSource {
	return &EdgeControlLibraryScriptsDataSource{}
}

// EdgeControlLibraryScriptsDataSource lists edge control library scripts.
type EdgeControlLibraryScriptsDataSource struct {
	client *cachefly.Client
}

func (d *EdgeControlLibraryScriptsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_edge_control_library_scripts"
}

func (d *EdgeControlLibraryScriptsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the edge control script library: `SYSTEM` scripts curated by CacheFly and the account's own `USER` scripts.",

		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: "Only return scripts of this ownership type: `USER` or `SYSTEM`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(api.EdgeControlLibraryTypeUser, api.EdgeControlLibraryTypeSystem),
				},
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: "Only return scripts of this kind: `REQUEST` or `RESPONSE`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(string(api.EdgeControlScriptKindRequest), string(api.EdgeControlScriptKindResponse)),
				},
			},
			"search": schema.StringAttribute{
				Description: "Only return scripts whose name matches this text (at least 2 characters).",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(2),
				},
			},
			"offset": schema.Int32Attribute{
				Description: "Number of scripts to skip (default: 0).",
				Optional:    true,
				Validators: []validator.Int32{
					int32validator.AtLeast(0),
				},
			},
			"limit": schema.Int32Attribute{
				Description: "Number of scripts fetched per request while reading all pages (default: 100, max: 1000).",
				Optional:    true,
				Validators: []validator.Int32{
					int32validator.Between(1, 1000),
				},
			},
			"scripts": schema.ListNestedAttribute{
				Description: "Library scripts.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the library script.",
							Computed:    true,
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
				},
			},
		},
	}
}

func (d *EdgeControlLibraryScriptsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EdgeControlLibraryScriptsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.EdgeControlLibraryScriptsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := api.ListEdgeControlLibraryOptions{
		Type:   data.Type.ValueString(),
		Kind:   api.EdgeControlScriptKind(data.Kind.ValueString()),
		Search: data.Search.ValueString(),
		Offset: int(data.Offset.ValueInt32()),
		Limit:  int(data.Limit.ValueInt32()),
	}
	if opts.Limit <= 0 {
		opts.Limit = 100
	}

	var all []api.EdgeControlLibraryScript
	for {
		page, err := d.client.EdgeControlLibrary.List(ctx, opts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading CacheFly Edge Control Library Scripts",
				"Could not list library scripts: "+err.Error(),
			)
			return
		}

		all = append(all, page.Scripts...)

		fetched := len(page.Scripts)
		opts.Offset += fetched
		if fetched < opts.Limit || (page.Meta.Count > 0 && opts.Offset >= page.Meta.Count) {
			break
		}
	}

	data.Scripts = make([]models.EdgeControlLibraryScriptModel, 0, len(all))
	for i := range all {
		data.Scripts = append(data.Scripts, models.NewEdgeControlLibraryScriptModel(&all[i]))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
