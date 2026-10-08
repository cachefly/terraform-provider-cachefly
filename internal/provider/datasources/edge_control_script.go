package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	_ datasource.DataSource              = &EdgeControlScriptDataSource{}
	_ datasource.DataSourceWithConfigure = &EdgeControlScriptDataSource{}
)

func NewEdgeControlScriptDataSource() datasource.DataSource {
	return &EdgeControlScriptDataSource{}
}

// EdgeControlScriptDataSource reads one edge control script version.
type EdgeControlScriptDataSource struct {
	client *cachefly.Client
}

func (d *EdgeControlScriptDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_edge_control_script"
}

func (d *EdgeControlScriptDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a published edge control script of a service: the active version, or the version set in `version`.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier in the form `service_id:kind:version`.",
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
			"version": schema.Int64Attribute{
				Description: "Version to read. Defaults to the active version.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"script": schema.StringAttribute{
				Description: "JavaScript source of the version.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the version (ACTIVE or DEACTIVATED).",
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
	}
}

func (d *EdgeControlScriptDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EdgeControlScriptDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.EdgeControlScriptDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()
	kind := api.EdgeControlScriptKind(data.Kind.ValueString())

	var script *api.EdgeControlScript
	var err error
	if !data.Version.IsNull() && !data.Version.IsUnknown() {
		script, err = d.client.EdgeControlScripts.GetVersion(ctx, serviceID, kind, int(data.Version.ValueInt64()))
	} else {
		script, err = d.client.EdgeControlScripts.GetActiveVersion(ctx, serviceID, kind)
		if isNotFoundError(err) {
			resp.Diagnostics.AddError(
				"No Active CacheFly Edge Control Script",
				fmt.Sprintf("Service %s has no active %s script. Set version to read a specific version.", serviceID, kind),
			)
			return
		}
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading CacheFly Edge Control Script",
			fmt.Sprintf("Could not read the %s script of service %s: %s", kind, serviceID, err),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s:%d", serviceID, kind, script.Version))
	data.Version = types.Int64Value(int64(script.Version))
	data.Script = types.StringValue(script.Script)
	data.Status = types.StringValue(script.Status)
	data.ScriptSize = types.Int64Value(int64(script.ScriptSize))
	data.Filename = types.StringValue(script.Filename)
	data.LastActivatedAt = types.StringValue(script.LastActivatedAt)
	data.CreatedAt = types.StringValue(script.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
