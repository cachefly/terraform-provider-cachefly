package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/cachefly/cachefly-go-sdk/pkg/cachefly"
	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// satisfy framework interfaces.
var (
	_ resource.Resource              = &ServiceOptionsResource{}
	_ resource.ResourceWithConfigure = &ServiceOptionsResource{}
)

func NewServiceOptionsResource() resource.Resource {
	return &ServiceOptionsResource{}
}

// ServiceOptionsResource defines the resource implementation.
type ServiceOptionsResource struct {
	client *cachefly.Client
}

func (r *ServiceOptionsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_options"
}

func (r *ServiceOptionsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages service options configuration for a CacheFly service.",
		Attributes: map[string]schema.Attribute{
			"service_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the service to configure options for.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ftp": schema.BoolAttribute{
				MarkdownDescription: "Enable FTP access for the service.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"cors": schema.BoolAttribute{
				MarkdownDescription: "Enable CORS (Cross-Origin Resource Sharing) headers.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"auto_redirect": schema.BoolAttribute{
				MarkdownDescription: "Enable automatic redirects.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"brotli_compression": schema.BoolAttribute{
				MarkdownDescription: "Enable Brotli compression.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"brotli_support": schema.BoolAttribute{
				MarkdownDescription: "Enable Brotli support.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"livestreaming": schema.BoolAttribute{
				MarkdownDescription: "Enable livestreaming support.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"nocache": schema.BoolAttribute{
				MarkdownDescription: "Disable caching.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"cache_by_geo_country": schema.BoolAttribute{
				MarkdownDescription: "Cache by geographic country.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"cache_by_region": schema.BoolAttribute{
				MarkdownDescription: "Cache by region.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"cache_by_referer": schema.BoolAttribute{
				MarkdownDescription: "Cache by HTTP referer header.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"normalize_query_string": schema.BoolAttribute{
				MarkdownDescription: "Normalize query string parameters.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"allow_retry": schema.BoolAttribute{
				MarkdownDescription: "Allow retry on origin failures.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"link_preheat": schema.BoolAttribute{
				MarkdownDescription: "Enable link preheating.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"edge_to_origin": schema.BoolAttribute{
				MarkdownDescription: "Enable edge to origin communication.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"follow_redirect": schema.BoolAttribute{
				MarkdownDescription: "Follow redirects from origin.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"purge_no_query": schema.BoolAttribute{
				MarkdownDescription: "Purge without query parameters.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"force_orig_qstring": schema.BoolAttribute{
				MarkdownDescription: "Force original query string.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"serve_stale": schema.BoolAttribute{
				MarkdownDescription: "Serve stale content when origin is unavailable.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"cache_post_requests": schema.BoolAttribute{
				MarkdownDescription: "Cache POST requests.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"send_xff": schema.BoolAttribute{
				MarkdownDescription: "Send X-Forwarded-For header.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"use_cf_doot_encoding": schema.BoolAttribute{
				MarkdownDescription: "Use CacheFly DoOT encoding.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"skip_url_encoding": schema.BoolAttribute{
				MarkdownDescription: "Skip URL encoding.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"trace": schema.BoolAttribute{
				MarkdownDescription: "Enable tracing.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"use_slicer": schema.BoolAttribute{
				MarkdownDescription: "Use slicer for large files.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"protect_serve_key_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable protect serve key.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"api_key_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable API key authentication.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"error_ttl": schema.SingleNestedAttribute{
				MarkdownDescription: "Origin Error TTL.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable origin error TTL.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"value": schema.Int64Attribute{
						MarkdownDescription: "TTL value",
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(60),
					},
				},
			},
			"con_timeout": schema.SingleNestedAttribute{
				MarkdownDescription: "Connect Timeout.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable Connect Timeout.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"value": schema.Int64Attribute{
						MarkdownDescription: "value",
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(3),
					},
				},
			},
			"max_cons": schema.SingleNestedAttribute{
				MarkdownDescription: "Maximum connections to origin.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable maximum connections limit.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"value": schema.Int64Attribute{
						MarkdownDescription: "Maximum number of connections.",
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(10),
					},
				},
			},
			"ttfb_timeout": schema.SingleNestedAttribute{
				MarkdownDescription: "Time to First Byte timeout from origin.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable TTFB timeout.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"value": schema.Int64Attribute{
						MarkdownDescription: "TTFB timeout value in seconds.",
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(3),
					},
				},
			},
			"origin_hostheader": schema.SingleNestedAttribute{
				MarkdownDescription: "Origin host header configuration.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable origin host header override.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"value": schema.ListAttribute{
						MarkdownDescription: "List of origin host header values.",
						ElementType:         types.StringType,
						Optional:            true,
						Computed:            true,
					},
				},
			},
			"shared_shield": schema.SingleNestedAttribute{
				MarkdownDescription: "Shared shield configuration.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable shared shield.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Shared shield location code. Must be one of: IAD, ORD, FRA, VIE.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						Validators: []validator.String{
							stringvalidator.OneOf("", "IAD", "ORD", "FRA", "VIE"),
						},
					},
				},
			},
			"purge_mode": schema.SingleNestedAttribute{
				MarkdownDescription: "Purge mode configuration.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable custom purge mode.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Purge mode value. This is computed when enabled is false.",
						Optional:            true,
						Computed:            true, // Always computed to allow API to set the value
						Default:             stringdefault.StaticString("2"),
					},
				},
			},
			"dir_purge_skip": schema.SingleNestedAttribute{
				MarkdownDescription: "Directory purge skip configuration.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable directory purge skip.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"value": schema.Int64Attribute{
						MarkdownDescription: "Directory purge skip value. This is computed when enabled is false.",
						Optional:            true,
						Computed:            true, // Always computed to allow API to set the value
						Default:             int64default.StaticInt64(0),
					},
				},
			},
			"skip_encoding_ext": schema.SingleNestedAttribute{
				Description: "Skip encoding for specified file extensions",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Enable skip encoding extensions",
						Optional:    true,
						Computed:    true,
					},
					"value": schema.ListAttribute{
						Description: "List of file extensions to skip encoding",
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
					},
				},
			},
			"redirect": schema.SingleNestedAttribute{
				Description: "Redirect configuration",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Enable redirect",
						Optional:    true,
						Computed:    true,
					},
					"value": schema.StringAttribute{
						Description: "Redirect URL",
						Optional:    true,
						Computed:    true,
					},
				},
			},
			"bw_throttle": schema.SingleNestedAttribute{
				Description: "Bandwidth throttle configuration",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Enable bandwidth throttling",
						Optional:    true,
						Computed:    true,
					},
					"value": schema.Int64Attribute{
						Description: "Bandwidth throttle value in bytes per second",
						Optional:    true,
						Computed:    true,
					},
				},
			},
			"expiry_headers": schema.ListNestedAttribute{
				Description: "Expiry headers configuration",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							Description: "Path for the expiry header rule",
							Required:    true,
						},
						"extension": schema.StringAttribute{
							Description: "File extension for the expiry header rule",
							Required:    true,
						},
						"expiry_time": schema.Int64Attribute{
							Description: "Expiry time in seconds",
							Required:    true,
						},
					},
				},
			},
			"skip_pserve_ext": schema.SingleNestedAttribute{
				Description: "Skip pserve for specified file extensions",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Enable skip pserve extensions",
						Optional:    true,
						Computed:    true,
					},
					"value": schema.ListAttribute{
						Description: "List of file extensions to skip pserve",
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
					},
				},
			},
			"http_methods": schema.SingleNestedAttribute{
				Description: "HTTP methods configuration",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Enable HTTP methods filtering",
						Optional:    true,
						Computed:    true,
					},
					"value": schema.SingleNestedAttribute{
						Description: "HTTP methods settings",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"get": schema.BoolAttribute{
								Description: "Allow GET requests",
								Optional:    true,
								Computed:    true,
							},
							"head": schema.BoolAttribute{
								Description: "Allow HEAD requests",
								Optional:    true,
								Computed:    true,
							},
							"options": schema.BoolAttribute{
								Description: "Allow OPTIONS requests",
								Optional:    true,
								Computed:    true,
							},
							"put": schema.BoolAttribute{
								Description: "Allow PUT requests",
								Optional:    true,
								Computed:    true,
							},
							"post": schema.BoolAttribute{
								Description: "Allow POST requests",
								Optional:    true,
								Computed:    true,
							},
							"patch": schema.BoolAttribute{
								Description: "Allow PATCH requests",
								Optional:    true,
								Computed:    true,
							},
							"delete": schema.BoolAttribute{
								Description: "Allow DELETE requests",
								Optional:    true,
								Computed:    true,
							},
						},
					},
				},
			},
			"reverse_proxy": schema.SingleNestedAttribute{
				MarkdownDescription: "Reverse proxy configuration.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable reverse proxy.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"hostname": schema.StringAttribute{
						MarkdownDescription: "Hostname for reverse proxy.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
					},
					"prepend": schema.StringAttribute{
						MarkdownDescription: "Path to prepend to requests.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
					},
					"ttl": schema.Int64Attribute{
						MarkdownDescription: "TTL for reverse proxy cache.",
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(0),
					},
					"cache_by_query_param": schema.BoolAttribute{
						MarkdownDescription: "Cache by query parameters.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"origin_scheme": schema.StringAttribute{
						MarkdownDescription: "Origin scheme (http or https).",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("http"),
					},
					"use_robots_txt": schema.BoolAttribute{
						MarkdownDescription: "Use robots.txt from origin.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
					"mode": schema.StringAttribute{
						MarkdownDescription: "Reverse proxy mode.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
					},
				},
			},
		},
	}
}

func (r *ServiceOptionsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServiceOptionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.ServiceOptionsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()

	tflog.Debug(ctx, "Creating service options", map[string]interface{}{
		"service_id": serviceID,
	})

	// Convert to SDK model
	opts := data.ToSDKServiceOptions(ctx)

	// Create/Update service options via API
	updatedOpts, err := r.client.ServiceOptions.SaveBasicOptions(ctx, serviceID, *opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating CacheFly Service Options",
			"Could not create service options, unexpected error: "+err.Error(),
		)
		return
	}

	// Handle protect serve key if requested to be enabled
	if data.ProtectServeKeyEnabled.ValueBool() {
		tflog.Debug(ctx, "Enabling protect serve key during create", map[string]interface{}{
			"service_id": serviceID,
		})

		_, err := r.client.ServiceOptions.RecreateProtectServeKey(ctx, serviceID, "")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error enabling ProtectServe key during create",
				"Could not enable ProtectServe key: "+err.Error(),
			)
			return
		}

		// Re-read the options to get the updated state after enabling protect serve
		updatedOpts, err = r.client.ServiceOptions.GetBasicOptions(ctx, serviceID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading service options after enabling protect serve",
				"Could not read service options: "+err.Error(),
			)
			return
		}

		tflog.Debug(ctx, "Protect serve key enabled successfully during create")
	}

	// Convert response back to model
	data.FromSDKServiceOptions(ctx, updatedOpts)

	tflog.Debug(ctx, "Service options created successfully", map[string]interface{}{
		"service_id": serviceID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceOptionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.ServiceOptionsModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()
	tflog.Debug(ctx, "Reading service options", map[string]interface{}{
		"service_id": serviceID,
	})

	// Get service options from API
	opts, err := r.client.ServiceOptions.GetBasicOptions(ctx, serviceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading CacheFly Service Options",
			"Could not read service options for service ID "+serviceID+": "+err.Error(),
		)
		return
	}

	// Convert response back to model
	data.FromSDKServiceOptions(ctx, opts)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceOptionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.ServiceOptionsModel
	var state models.ServiceOptionsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()
	currentEnabled := state.ProtectServeKeyEnabled.ValueBool()
	requestedEnabled := data.ProtectServeKeyEnabled.ValueBool()

	// Convert to SDK model
	opts := data.ToSDKServiceOptions(ctx)

	// Update service options via API first
	updatedOpts, err := r.client.ServiceOptions.SaveBasicOptions(ctx, serviceID, *opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating CacheFly Service Options",
			"Could not update service options, unexpected error: "+err.Error(),
		)
		return
	}

	// Handle protect serve key changes
	if currentEnabled != requestedEnabled {
		if requestedEnabled {
			// Enabling protect serve key (false -> true)
			_, err := r.client.ServiceOptions.RecreateProtectServeKey(ctx, serviceID, "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error enabling ProtectServe key",
					"Could not enable ProtectServe key: "+err.Error(),
				)
				return
			}
		} else {
			// Disabling protect serve key (true -> false)
			err := r.client.ServiceOptions.DeleteProtectServeKey(ctx, serviceID)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error disabling ProtectServe key",
					"Could not disable ProtectServe key: "+err.Error(),
				)
				return
			}
		}

		// Re-read the options to get the updated state
		updatedOpts, err = r.client.ServiceOptions.GetBasicOptions(ctx, serviceID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading updated service options",
				"Could not read service options after protect serve change: "+err.Error(),
			)
			return
		}
	}

	// Convert response back to model
	data.FromSDKServiceOptions(ctx, updatedOpts)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceOptionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.ServiceOptionsModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()

	tflog.Debug(ctx, "Deleting service options (resetting to defaults)", map[string]interface{}{
		"service_id": serviceID,
	})

	// First, disable protect serve key if it's enabled
	if data.ProtectServeKeyEnabled.ValueBool() {
		tflog.Debug(ctx, "Disabling protect serve key before reset", map[string]interface{}{
			"service_id": serviceID,
		})

		err := r.client.ServiceOptions.DeleteProtectServeKey(ctx, serviceID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error disabling ProtectServe key during delete",
				"Could not disable ProtectServe key: "+err.Error(),
			)
			return
		}

		tflog.Debug(ctx, "Protect serve key disabled successfully")
	}

	// Reset service options to defaults
	defaultOpts := api.ServiceOptions{
		// Set all boolean options to false (defaults)
		FTP:                    false,
		CORS:                   false,
		AutoRedirect:           false,
		BrotliCompression:      false,
		BrotliSupport:          false,
		Livestreaming:          false,
		NoCache:                false,
		CacheByGeoCountry:      false,
		CacheByRegion:          false,
		CacheByReferer:         false,
		NormalizeQueryString:   false,
		AllowRetry:             false,
		LinkPreheat:            false,
		EdgeToOrigin:           false,
		FollowRedirect:         false,
		PurgeNoQuery:           false,
		ForceOrigQString:       false,
		ServeStale:             false,
		CachePostRequests:      false,
		SendXFF:                false,
		UseCFDooTEncoding:      false,
		SkipURLEncoding:        false,
		Trace:                  false,
		UseSlicer:              false,
		ProtectServeKeyEnabled: false,
		APIKeyEnabled:          false,

		// Default reverse proxy config
		ReverseProxy: api.ReverseProxyConfig{
			Enabled:           false,
			Hostname:          "",
			Prepend:           "",
			TTL:               0,
			CacheByQueryParam: false,
			OriginScheme:      "http",
			UseRobotsTXT:      false,
			Mode:              "",
		},

		// Initialize empty arrays
		MimeTypesOverrides: make([]api.MimeTypeOverride, 0),
		ExpiryHeaders:      make([]api.ExpiryHeader, 0),

		// Default option values
		RawLogs:          api.Option{Enabled: false, Value: ""},
		ErrorTTL:         api.Option{Enabled: false, Value: ""},
		TTFBTimeout:      api.Option{Enabled: false, Value: ""},
		ConTimeout:       api.Option{Enabled: false, Value: ""},
		SharedShield:     api.Option{Enabled: false, Value: ""},
		BWThrottle:       api.Option{Enabled: false, Value: ""},
		PurgeMode:        api.Option{Enabled: false, Value: ""},
		DirPurgeSkip:     api.Option{Enabled: false, Value: ""},
		SkipPserveExt:    api.Option{Enabled: false, Value: ""},
		HTTPMethods:      api.Option{Enabled: false, Value: ""},
		BWThrottleQuery:  api.Option{Enabled: false, Value: ""},
		OriginHostHeader: api.Option{Enabled: false, Value: ""},
		MaxCons:          api.Option{Enabled: false, Value: ""},
		Slice:            api.Option{Enabled: false, Value: ""},
		Redirect:         api.Option{Enabled: false, Value: ""},
		SkipEncodingExt:  api.Option{Enabled: false, Value: ""},
	}

	_, err := r.client.ServiceOptions.SaveBasicOptions(ctx, serviceID, defaultOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting CacheFly Service Options",
			"Could not reset service options to defaults, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Service options reset to defaults successfully")
}
