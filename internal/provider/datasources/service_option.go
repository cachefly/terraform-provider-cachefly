package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/avvvet/terraform-provider-cachefly/internal/provider/models"
	"github.com/cachefly/cachefly-go-sdk/pkg/cachefly"
)

var _ datasource.DataSource = &ServiceOptionsDataSource{}

func NewServiceOptionsDataSource() datasource.DataSource {
	return &ServiceOptionsDataSource{}
}

type ServiceOptionsDataSource struct {
	client *cachefly.Client
}

func (d *ServiceOptionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_options"
}

func (d *ServiceOptionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches service options configuration for a CacheFly service.",
		Attributes: map[string]schema.Attribute{
			"service_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the service to fetch options for.",
				Required:            true,
			},
			"ftp": schema.BoolAttribute{
				MarkdownDescription: "FTP access enabled for the service.",
				Computed:            true,
			},
			"cors": schema.BoolAttribute{
				MarkdownDescription: "CORS (Cross-Origin Resource Sharing) headers enabled.",
				Computed:            true,
			},
			"auto_redirect": schema.BoolAttribute{
				MarkdownDescription: "Automatic redirects enabled.",
				Computed:            true,
			},
			"brotli_compression": schema.BoolAttribute{
				MarkdownDescription: "Brotli compression enabled.",
				Computed:            true,
			},
			"brotli_support": schema.BoolAttribute{
				MarkdownDescription: "Brotli support enabled.",
				Computed:            true,
			},
			"livestreaming": schema.BoolAttribute{
				MarkdownDescription: "Livestreaming support enabled.",
				Computed:            true,
			},
			"nocache": schema.BoolAttribute{
				MarkdownDescription: "Caching disabled.",
				Computed:            true,
			},
			"cache_by_geo_country": schema.BoolAttribute{
				MarkdownDescription: "Cache by geographic country enabled.",
				Computed:            true,
			},
			"cache_by_region": schema.BoolAttribute{
				MarkdownDescription: "Cache by region enabled.",
				Computed:            true,
			},
			"cache_by_referer": schema.BoolAttribute{
				MarkdownDescription: "Cache by HTTP referer header enabled.",
				Computed:            true,
			},
			"normalize_query_string": schema.BoolAttribute{
				MarkdownDescription: "Query string normalization enabled.",
				Computed:            true,
			},
			"allow_retry": schema.BoolAttribute{
				MarkdownDescription: "Retry on origin failures enabled.",
				Computed:            true,
			},
			"link_preheat": schema.BoolAttribute{
				MarkdownDescription: "Link preheating enabled.",
				Computed:            true,
			},
			"edge_to_origin": schema.BoolAttribute{
				MarkdownDescription: "Edge to origin communication enabled.",
				Computed:            true,
			},
			"follow_redirect": schema.BoolAttribute{
				MarkdownDescription: "Follow redirects from origin enabled.",
				Computed:            true,
			},
			"purge_no_query": schema.BoolAttribute{
				MarkdownDescription: "Purge without query parameters enabled.",
				Computed:            true,
			},
			"force_orig_qstring": schema.BoolAttribute{
				MarkdownDescription: "Force original query string enabled.",
				Computed:            true,
			},
			"serve_stale": schema.BoolAttribute{
				MarkdownDescription: "Serve stale content when origin is unavailable.",
				Computed:            true,
			},
			"cache_post_requests": schema.BoolAttribute{
				MarkdownDescription: "Cache POST requests enabled.",
				Computed:            true,
			},
			"send_xff": schema.BoolAttribute{
				MarkdownDescription: "Send X-Forwarded-For header enabled.",
				Computed:            true,
			},
			"use_cf_doot_encoding": schema.BoolAttribute{
				MarkdownDescription: "CacheFly DoOT encoding enabled.",
				Computed:            true,
			},
			"skip_url_encoding": schema.BoolAttribute{
				MarkdownDescription: "Skip URL encoding enabled.",
				Computed:            true,
			},
			"trace": schema.BoolAttribute{
				MarkdownDescription: "Tracing enabled.",
				Computed:            true,
			},
			"use_slicer": schema.BoolAttribute{
				MarkdownDescription: "Slicer for large files enabled.",
				Computed:            true,
			},
			"protect_serve_key_enabled": schema.BoolAttribute{
				MarkdownDescription: "Protect serve key enabled.",
				Computed:            true,
			},
			"api_key_enabled": schema.BoolAttribute{
				MarkdownDescription: "API key authentication enabled.",
				Computed:            true,
			},
			"reverse_proxy": schema.SingleNestedAttribute{
				MarkdownDescription: "Reverse proxy configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Reverse proxy enabled.",
						Computed:            true,
					},
					"hostname": schema.StringAttribute{
						MarkdownDescription: "Hostname for reverse proxy.",
						Computed:            true,
					},
					"prepend": schema.StringAttribute{
						MarkdownDescription: "Path to prepend to requests.",
						Computed:            true,
					},
					"ttl": schema.Int64Attribute{
						MarkdownDescription: "TTL for reverse proxy cache.",
						Computed:            true,
					},
					"cache_by_query_param": schema.BoolAttribute{
						MarkdownDescription: "Cache by query parameters.",
						Computed:            true,
					},
					"origin_scheme": schema.StringAttribute{
						MarkdownDescription: "Origin scheme (http or https).",
						Computed:            true,
					},
					"use_robots_txt": schema.BoolAttribute{
						MarkdownDescription: "Use robots.txt from origin.",
						Computed:            true,
					},
					"mode": schema.StringAttribute{
						MarkdownDescription: "Reverse proxy mode.",
						Computed:            true,
					},
				},
			},
			"mime_types_overrides": schema.ListNestedAttribute{
				MarkdownDescription: "MIME type overrides for file extensions.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"extension": schema.StringAttribute{
							MarkdownDescription: "File extension (without dot).",
							Computed:            true,
						},
						"mime_type": schema.StringAttribute{
							MarkdownDescription: "MIME type used for this extension.",
							Computed:            true,
						},
					},
				},
			},
			"expiry_headers": schema.ListNestedAttribute{
				MarkdownDescription: "Expiry headers configuration for paths and extensions.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							MarkdownDescription: "Path pattern matched.",
							Computed:            true,
						},
						"extension": schema.StringAttribute{
							MarkdownDescription: "File extension matched.",
							Computed:            true,
						},
						"expiry_time": schema.Int64Attribute{
							MarkdownDescription: "Expiry time in seconds.",
							Computed:            true,
						},
					},
				},
			},
			"raw_logs": schema.SingleNestedAttribute{
				MarkdownDescription: "Raw logs configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Raw logs enabled.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Raw logs configuration value.",
						Computed:            true,
					},
				},
			},
			"error_ttl": schema.SingleNestedAttribute{
				MarkdownDescription: "Error TTL configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Error TTL enabled.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Error TTL value.",
						Computed:            true,
					},
				},
			},
			"ttfb_timeout": schema.SingleNestedAttribute{
				MarkdownDescription: "Time to first byte timeout configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "TTFB timeout enabled.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "TTFB timeout value.",
						Computed:            true,
					},
				},
			},
			"con_timeout": schema.SingleNestedAttribute{
				MarkdownDescription: "Connection timeout configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Connection timeout enabled.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Connection timeout value.",
						Computed:            true,
					},
				},
			},
			"shared_shield": schema.SingleNestedAttribute{
				MarkdownDescription: "Shared shield configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Shared shield enabled.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Shared shield value.",
						Computed:            true,
					},
				},
			},
			"bw_throttle": schema.SingleNestedAttribute{
				MarkdownDescription: "Bandwidth throttle configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Bandwidth throttle enabled.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Bandwidth throttle value.",
						Computed:            true,
					},
				},
			},
			"purge_mode": schema.SingleNestedAttribute{
				MarkdownDescription: "Purge mode configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Purge mode enabled.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Purge mode value.",
						Computed:            true,
					},
				},
			},
			"dir_purge_skip": schema.SingleNestedAttribute{
				MarkdownDescription: "Directory purge skip configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Directory purge skip enabled.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Directory purge skip value.",
						Computed:            true,
					},
				},
			},
			"skip_pserve_ext": schema.SingleNestedAttribute{
				MarkdownDescription: "Skip pserve extension configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Skip pserve extension enabled.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Skip pserve extension value.",
						Computed:            true,
					},
				},
			},
			"http_methods": schema.SingleNestedAttribute{
				MarkdownDescription: "HTTP methods configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "HTTP methods restriction enabled.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Allowed HTTP methods value.",
						Computed:            true,
					},
				},
			},
			"bw_throttle_query": schema.SingleNestedAttribute{
				MarkdownDescription: "Bandwidth throttle query configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Bandwidth throttle query enabled.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Bandwidth throttle query value.",
						Computed:            true,
					},
				},
			},
			"origin_host_header": schema.SingleNestedAttribute{
				MarkdownDescription: "Origin host header configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Origin host header override enabled.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Origin host header value.",
						Computed:            true,
					},
				},
			},
			"max_cons": schema.SingleNestedAttribute{
				MarkdownDescription: "Maximum connections configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Maximum connections limit enabled.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Maximum connections value.",
						Computed:            true,
					},
				},
			},
			"slice": schema.SingleNestedAttribute{
				MarkdownDescription: "Slice configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Slicing enabled.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Slice configuration value.",
						Computed:            true,
					},
				},
			},
			"redirect": schema.SingleNestedAttribute{
				MarkdownDescription: "Redirect configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Redirect enabled.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Redirect configuration value.",
						Computed:            true,
					},
				},
			},
			"skip_encoding_ext": schema.SingleNestedAttribute{
				MarkdownDescription: "Skip encoding extension configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Skip encoding extension enabled.",
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Skip encoding extension value.",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (d *ServiceOptionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceOptionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.ServiceOptionsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()
	if serviceID == "" {
		resp.Diagnostics.AddError(
			"Missing Service ID",
			"Service ID is required to fetch service options.",
		)
		return
	}

	tflog.Debug(ctx, "Reading service options", map[string]interface{}{
		"service_id": serviceID,
	})

	// Get service options from API
	serviceOptions, err := d.client.ServiceOptions.GetBasicOptions(ctx, serviceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading service options",
			fmt.Sprintf("Could not read service options for service %s: %s", serviceID, err),
		)
		return
	}

	// Convert SDK model to Terraform model
	data.FromSDKServiceOptions(ctx, serviceOptions)

	tflog.Debug(ctx, "Successfully read service options", map[string]interface{}{
		"service_id": serviceID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
