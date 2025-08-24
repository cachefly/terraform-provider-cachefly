// internal/provider/datasources/origin.go
package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly"
	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &OriginDataSource{}

func NewOriginDataSource() datasource.DataSource {
	return &OriginDataSource{}
}

// OriginDataSource defines the data source implementation.
type OriginDataSource struct {
	client *cachefly.Client
}

func (d *OriginDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_origin"
}

func (d *OriginDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CacheFly Origin data source. Look up a specific origin server configuration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the origin.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of origin.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the origin.",
				Computed:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "Hostname of the origin server.",
				Computed:    true,
			},
			"scheme": schema.StringAttribute{
				Description: "Protocol scheme (http or https).",
				Computed:    true,
			},
			"cache_by_query_param": schema.BoolAttribute{
				Description: "Whether to cache content based on query parameters.",
				Computed:    true,
			},
			"gzip": schema.BoolAttribute{
				Description: "Whether gzip compression is enabled.",
				Computed:    true,
			},
			"ttl": schema.Int32Attribute{
				Description: "Time to live (TTL) in seconds for cached content.",
				Computed:    true,
			},
			"missed_ttl": schema.Int32Attribute{
				Description: "TTL in seconds for missed (404/error) responses.",
				Computed:    true,
			},
			"connection_timeout": schema.Int32Attribute{
				Description: "Connection timeout in seconds.",
				Computed:    true,
			},
			"time_to_first_byte_timeout": schema.Int32Attribute{
				Description: "Time to first byte timeout in seconds.",
				Computed:    true,
			},

			// S3-specific attributes
			"access_key": schema.StringAttribute{
				Description: "S3 access key (for S3 origins).",
				Computed:    true,
				Sensitive:   true,
			},
			"secret_key": schema.StringAttribute{
				Description: "S3 secret key (for S3 origins).",
				Computed:    true,
				Sensitive:   true,
			},
			"region": schema.StringAttribute{
				Description: "S3 region (for S3 origins).",
				Computed:    true,
			},
			"signature_version": schema.StringAttribute{
				Description: "S3 signature version (for S3 origins).",
				Computed:    true,
			},

			// Computed timestamps
			"created_at": schema.StringAttribute{
				Description: "When the origin was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the origin was last updated.",
				Computed:    true,
			},

			// Optional query parameters
			"response_type": schema.StringAttribute{
				Description: "Optional response type parameter for the API call.",
				Optional:    true,
			},
		},
	}
}

func (d *OriginDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OriginDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.OriginDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	origin, err := d.client.Origins.GetByID(
		ctx,
		data.ID.ValueString(),
		data.ResponseType.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading CacheFly Origin",
			"Could not read origin ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to data model
	d.mapOriginToDataSource(origin, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper function to map SDK Origin to data source model
func (d *OriginDataSource) mapOriginToDataSource(origin *api.Origin, data *models.OriginDataSourceModel) {
	data.ID = types.StringValue(origin.ID)
	data.Type = types.StringValue(origin.Type)
	data.Name = types.StringPointerValue(origin.Name)
	data.Scheme = types.StringPointerValue(origin.Scheme)
	data.CacheByQueryParam = types.BoolPointerValue(origin.CacheByQueryParam)
	data.Gzip = types.BoolPointerValue(origin.Gzip)
	data.TTL = types.Int32PointerValue(origin.TTL)
	data.MissedTTL = types.Int32PointerValue(origin.MissedTTL)
	data.CreatedAt = types.StringValue(origin.CreatedAt)
	data.UpdatedAt = types.StringValue(origin.UpdatedAt)

	if origin.Type == "WEB" {
		data.Hostname = types.StringPointerValue(origin.Hostname)
	} else {
		data.Hostname = types.StringPointerValue(origin.Host)
	}

	data.ConnectionTimeout = types.Int32PointerValue(origin.ConnectionTimeout)
	data.TimeToFirstByteTimeout = types.Int32PointerValue(origin.TimeToFirstByteTimeout)

	data.AccessKey = types.StringPointerValue(origin.AccessKey)
	data.SecretKey = types.StringPointerValue(origin.SecretKey)
	data.Region = types.StringPointerValue(origin.Region)
	data.SignatureVersion = types.StringPointerValue(origin.SignatureVersion)
}
