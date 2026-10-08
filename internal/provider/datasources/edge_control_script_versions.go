package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly"
	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &EdgeControlScriptVersionsDataSource{}
	_ datasource.DataSourceWithConfigure = &EdgeControlScriptVersionsDataSource{}
)

func NewEdgeControlScriptVersionsDataSource() datasource.DataSource {
	return &EdgeControlScriptVersionsDataSource{}
}

// EdgeControlScriptVersionsDataSource lists the published versions of an edge
// control script.
type EdgeControlScriptVersionsDataSource struct {
	client *cachefly.Client
}

func (d *EdgeControlScriptVersionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_edge_control_script_versions"
}

func (d *EdgeControlScriptVersionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all published versions of the edge control script of a service for one kind, newest first.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier in the form `service_id:kind`.",
				Computed:            true,
			},
			"service_id": schema.StringAttribute{
				Description: "ID of the service.",
				Required:    true,
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: "Kind of script: `REQUEST` or `RESPONSE`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(string(api.EdgeControlScriptKindRequest), string(api.EdgeControlScriptKindResponse)),
				},
			},
			"versions": schema.ListNestedAttribute{
				Description: "Published versions, newest first.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the version.",
							Computed:    true,
						},
						"version": schema.Int64Attribute{
							Description: "Version number.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the version (ACTIVE or DEACTIVATED).",
							Computed:    true,
						},
						"script": schema.StringAttribute{
							Description: "JavaScript source of the version.",
							Computed:    true,
						},
						"script_size": schema.Int64Attribute{
							Description: "Size of the script in bytes.",
							Computed:    true,
						},
						"filename": schema.StringAttribute{
							Description: "Name of the file the version is stored as.",
							Computed:    true,
						},
						"last_activated_at": schema.StringAttribute{
							Description: "When the version was last activated.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "When the version was published.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *EdgeControlScriptVersionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EdgeControlScriptVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.EdgeControlScriptVersionsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()
	kind := api.EdgeControlScriptKind(data.Kind.ValueString())

	versions, err := d.client.EdgeControlScripts.ListVersions(ctx, serviceID, kind)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading CacheFly Edge Control Script Versions",
			fmt.Sprintf("Could not list the %s script versions of service %s: %s", kind, serviceID, err),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", serviceID, kind))
	data.Versions = make([]models.EdgeControlScriptVersionModel, 0, len(versions))
	for i := range versions {
		data.Versions = append(data.Versions, models.NewEdgeControlScriptVersionModel(&versions[i]))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
