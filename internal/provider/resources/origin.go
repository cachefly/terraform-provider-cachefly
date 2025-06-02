// internal/provider/resources/origin.go
package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/cachefly/cachefly-go-sdk/pkg/cachefly"
	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"

	"github.com/avvvet/terraform-provider-cachefly/internal/provider/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &OriginResource{}
	_ resource.ResourceWithImportState = &OriginResource{}
)

func NewOriginResource() resource.Resource {
	return &OriginResource{}
}

// OriginResource defines the resource implementation.
type OriginResource struct {
	client *cachefly.Client
}

func (r *OriginResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_origin"
}

func (r *OriginResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CacheFly Origin resource. Manages backend servers that CacheFly fetches content from.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the origin.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Type of origin (e.g., 'http', 's3', 'gcs').",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the origin.",
				Optional:    true,
				Computed:    true,
			},
			"host": schema.StringAttribute{
				Description: "Hostname of the origin server.",
				Required:    true,
			},
			"scheme": schema.StringAttribute{
				Description: "Protocol scheme (HTTP, HTTPS, or FOLLOW).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("HTTPS"),
			},
			"cache_by_query_param": schema.BoolAttribute{
				Description: "Whether to cache content based on query parameters.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"gzip": schema.BoolAttribute{
				Description: "Whether to enable gzip compression.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"ttl": schema.Int64Attribute{
				Description: "Time to live (TTL) in seconds for cached content.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(86400), // 24 hours
			},
			"missed_ttl": schema.Int64Attribute{
				Description: "TTL in seconds for missed (404/error) responses.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(300), // 5 minutes
			},
			"connection_timeout": schema.Int64Attribute{
				Description: "Connection timeout in seconds.",
				Optional:    true,
				Computed:    true,
			},
			"time_to_first_byte_timeout": schema.Int64Attribute{
				Description: "Time to first byte timeout in seconds.",
				Optional:    true,
				Computed:    true,
			},

			// S3-specific attributes
			"access_key": schema.StringAttribute{
				Description: "S3 access key (for S3 origins).",
				Optional:    true,
				Sensitive:   true,
			},
			"secret_key": schema.StringAttribute{
				Description: "S3 secret key (for S3 origins).",
				Optional:    true,
				Sensitive:   true,
			},
			"region": schema.StringAttribute{
				Description: "S3 region (for S3 origins).",
				Optional:    true,
			},
			"signature_version": schema.StringAttribute{
				Description: "S3 signature version (for S3 origins).",
				Optional:    true,
			},

			// Computed attributes
			"created_at": schema.StringAttribute{
				Description: "When the origin was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the origin was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *OriginResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OriginResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.OriginResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build create request
	createReq := api.CreateOriginRequest{
		Type:     data.Type.ValueString(),
		Hostname: data.Host.ValueString(),
	}

	// Optional fields
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		createReq.Name = data.Name.ValueString()
	}
	if !data.Scheme.IsNull() && !data.Scheme.IsUnknown() {
		createReq.Scheme = data.Scheme.ValueString()
	}
	if !data.CacheByQueryParam.IsNull() && !data.CacheByQueryParam.IsUnknown() {
		createReq.CacheByQueryParam = data.CacheByQueryParam.ValueBool()
	}
	if !data.Gzip.IsNull() && !data.Gzip.IsUnknown() {
		createReq.Gzip = data.Gzip.ValueBool()
	}
	if !data.TTL.IsNull() && !data.TTL.IsUnknown() {
		createReq.TTL = int(data.TTL.ValueInt64())
	}
	if !data.MissedTTL.IsNull() && !data.MissedTTL.IsUnknown() {
		createReq.MissedTTL = int(data.MissedTTL.ValueInt64())
	}
	if !data.ConnectionTimeout.IsNull() && !data.ConnectionTimeout.IsUnknown() {
		createReq.ConnectionTimeout = int(data.ConnectionTimeout.ValueInt64())
	}
	if !data.TimeToFirstByteTimeout.IsNull() && !data.TimeToFirstByteTimeout.IsUnknown() {
		createReq.TimeToFirstByteTimeout = int(data.TimeToFirstByteTimeout.ValueInt64())
	}

	// S3-specific fields
	if !data.AccessKey.IsNull() && !data.AccessKey.IsUnknown() {
		createReq.AccessKey = data.AccessKey.ValueString()
	}
	if !data.SecretKey.IsNull() && !data.SecretKey.IsUnknown() {
		createReq.SecretKey = data.SecretKey.ValueString()
	}
	if !data.Region.IsNull() && !data.Region.IsUnknown() {
		createReq.Region = data.Region.ValueString()
	}
	if !data.SignatureVersion.IsNull() && !data.SignatureVersion.IsUnknown() {
		createReq.SignatureVersion = data.SignatureVersion.ValueString()
	}

	tflog.Debug(ctx, "Creating origin", map[string]interface{}{
		"type":     createReq.Type,
		"hostname": createReq.Hostname,
		"name":     createReq.Name,
	})

	origin, err := r.client.Origins.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating CacheFly Origin",
			"Could not create origin, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to state
	r.mapOriginToState(origin, &data)

	tflog.Debug(ctx, "Origin created successfully", map[string]interface{}{
		"origin_id": origin.ID,
		"hostname":  origin.Hostname,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OriginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.OriginResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading origin", map[string]interface{}{
		"origin_id": data.ID.ValueString(),
	})

	origin, err := r.client.Origins.GetByID(ctx, data.ID.ValueString(), "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading CacheFly Origin",
			"Could not read origin ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to state
	r.mapOriginToState(origin, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OriginResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.OriginResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update request with only changed fields
	updateReq := api.UpdateOriginRequest{}

	if !data.Type.IsNull() && !data.Type.IsUnknown() {
		updateReq.Type = data.Type.ValueString()
	}
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		updateReq.Name = data.Name.ValueString()
	}
	if !data.Host.IsNull() && !data.Host.IsUnknown() {
		updateReq.Hostname = data.Host.ValueString()
	}
	if !data.Scheme.IsNull() && !data.Scheme.IsUnknown() {
		updateReq.Scheme = data.Scheme.ValueString()
	}
	if !data.CacheByQueryParam.IsNull() && !data.CacheByQueryParam.IsUnknown() {
		updateReq.CacheByQueryParam = data.CacheByQueryParam.ValueBool()
	}
	if !data.Gzip.IsNull() && !data.Gzip.IsUnknown() {
		updateReq.Gzip = data.Gzip.ValueBool()
	}
	if !data.TTL.IsNull() && !data.TTL.IsUnknown() {
		updateReq.TTL = int(data.TTL.ValueInt64())
	}
	if !data.MissedTTL.IsNull() && !data.MissedTTL.IsUnknown() {
		updateReq.MissedTTL = int(data.MissedTTL.ValueInt64())
	}
	if !data.ConnectionTimeout.IsNull() && !data.ConnectionTimeout.IsUnknown() {
		updateReq.ConnectionTimeout = int(data.ConnectionTimeout.ValueInt64())
	}
	if !data.TimeToFirstByteTimeout.IsNull() && !data.TimeToFirstByteTimeout.IsUnknown() {
		updateReq.TimeToFirstByteTimeout = int(data.TimeToFirstByteTimeout.ValueInt64())
	}

	// S3-specific fields
	if !data.AccessKey.IsNull() && !data.AccessKey.IsUnknown() {
		updateReq.AccessKey = data.AccessKey.ValueString()
	}
	if !data.SecretKey.IsNull() && !data.SecretKey.IsUnknown() {
		updateReq.SecretKey = data.SecretKey.ValueString()
	}
	if !data.Region.IsNull() && !data.Region.IsUnknown() {
		updateReq.Region = data.Region.ValueString()
	}
	if !data.SignatureVersion.IsNull() && !data.SignatureVersion.IsUnknown() {
		updateReq.SignatureVersion = data.SignatureVersion.ValueString()
	}

	tflog.Debug(ctx, "Updating origin", map[string]interface{}{
		"origin_id": data.ID.ValueString(),
	})

	origin, err := r.client.Origins.UpdateByID(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating CacheFly Origin",
			"Could not update origin, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to state
	r.mapOriginToState(origin, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OriginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.OriginResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting origin", map[string]interface{}{
		"origin_id": data.ID.ValueString(),
	})

	err := r.client.Origins.Delete(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting CacheFly Origin",
			"Could not delete origin, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Origin deleted successfully")
}

func (r *OriginResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to map SDK Origin to Terraform state
func (r *OriginResource) mapOriginToState(origin *api.Origin, data *models.OriginResourceModel) {
	data.ID = types.StringValue(origin.ID)
	data.Type = types.StringValue(origin.Type)
	data.Name = types.StringValue(origin.Name)
	data.Host = types.StringValue(origin.Hostname)
	data.Scheme = types.StringValue(origin.Scheme)
	data.CacheByQueryParam = types.BoolValue(origin.CacheByQueryParam)
	data.Gzip = types.BoolValue(origin.Gzip)
	data.TTL = types.Int64Value(int64(origin.TTL))
	data.MissedTTL = types.Int64Value(int64(origin.MissedTTL))
	data.CreatedAt = types.StringValue(origin.CreatedAt)
	data.UpdatedAt = types.StringValue(origin.UpdatedAt)

	// Handle optional timeout fields
	if origin.ConnectionTimeout > 0 {
		data.ConnectionTimeout = types.Int64Value(int64(origin.ConnectionTimeout))
	} else {
		data.ConnectionTimeout = types.Int64Null()
	}
	if origin.TimeToFirstByteTimeout > 0 {
		data.TimeToFirstByteTimeout = types.Int64Value(int64(origin.TimeToFirstByteTimeout))
	} else {
		data.TimeToFirstByteTimeout = types.Int64Null()
	}

	// Handle S3-specific fields
	if origin.AccessKey != "" {
		data.AccessKey = types.StringValue(origin.AccessKey)
	} else {
		data.AccessKey = types.StringNull()
	}
	if origin.SecretKey != "" {
		data.SecretKey = types.StringValue(origin.SecretKey)
	} else {
		data.SecretKey = types.StringNull()
	}
	if origin.Region != "" {
		data.Region = types.StringValue(origin.Region)
	} else {
		data.Region = types.StringNull()
	}
	if origin.SignatureVersion != "" {
		data.SignatureVersion = types.StringValue(origin.SignatureVersion)
	} else {
		data.SignatureVersion = types.StringNull()
	}
}
