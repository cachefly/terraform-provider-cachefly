// internal/provider/resources/origin.go
package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly"
	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// satisfy framework interfaces.
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
				Description: "Type of origin ('WEB', 'GEO', 'FAILOVER', 'S3_BUCKET').",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the origin.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
				Description: "Hostname of the origin server.",
				Optional:    true,
			},
			"scheme": schema.StringAttribute{
				Description: "Protocol scheme (HTTP, HTTPS, or FOLLOW).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cache_by_query_param": schema.BoolAttribute{
				Description: "Whether to cache content based on query parameters.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"gzip": schema.BoolAttribute{
				Description: "Whether to enable gzip compression.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"ttl": schema.Int32Attribute{
				Description: "Time to live (TTL) in seconds for cached content.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"missed_ttl": schema.Int32Attribute{
				Description: "TTL in seconds for missed (404/error) responses.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"connection_timeout": schema.Int32Attribute{
				Description: "Connection timeout in seconds.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"time_to_first_byte_timeout": schema.Int32Attribute{
				Description: "Time to first byte timeout in seconds.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
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
		Type: data.Type.ValueString(),
	}

	if !data.Hostname.IsUnknown() {
		if data.Type.ValueString() == "WEB" {
			createReq.Hostname = data.Hostname.ValueStringPointer()
		} else {
			createReq.Host = data.Hostname.ValueStringPointer()
		}
	}
	if !data.Name.IsUnknown() {
		createReq.Name = data.Name.ValueStringPointer()
	}
	if !data.Scheme.IsUnknown() {
		createReq.Scheme = data.Scheme.ValueStringPointer()
	}
	if !data.CacheByQueryParam.IsUnknown() {
		createReq.CacheByQueryParam = data.CacheByQueryParam.ValueBoolPointer()
	}
	if !data.Gzip.IsUnknown() {
		createReq.Gzip = data.Gzip.ValueBoolPointer()
	}
	if !data.TTL.IsUnknown() {
		createReq.TTL = data.TTL.ValueInt32Pointer()
	}
	if !data.MissedTTL.IsUnknown() {
		createReq.MissedTTL = data.MissedTTL.ValueInt32Pointer()
	}
	if !data.ConnectionTimeout.IsUnknown() {
		createReq.ConnectionTimeout = data.ConnectionTimeout.ValueInt32Pointer()
	}
	if !data.TimeToFirstByteTimeout.IsUnknown() {
		createReq.TimeToFirstByteTimeout = data.TimeToFirstByteTimeout.ValueInt32Pointer()
	}
	if !data.AccessKey.IsUnknown() {
		createReq.AccessKey = data.AccessKey.ValueStringPointer()
	}
	if !data.SecretKey.IsUnknown() {
		createReq.SecretKey = data.SecretKey.ValueStringPointer()
	}
	if !data.Region.IsUnknown() {
		createReq.Region = data.Region.ValueStringPointer()
	}
	if !data.SignatureVersion.IsUnknown() {
		createReq.SignatureVersion = data.SignatureVersion.ValueStringPointer()
	}

	origin, err := r.client.Origins.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating CacheFly Origin",
			"Could not create origin, unexpected error: "+err.Error(),
		)
		return
	}

	r.mapOriginToState(origin, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OriginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.OriginResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	origin, err := r.client.Origins.GetByID(ctx, data.ID.ValueString(), "")
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Error Reading CacheFly Origin",
				"Could not read origin ID "+data.ID.ValueString()+": "+err.Error(),
			)
		}
		return
	}

	// Map response to state
	r.mapOriginToState(origin, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OriginResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.OriginResourceModel
	var state models.OriginResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := api.UpdateOriginRequest{}

	if !data.Type.Equal(state.Type) {
		updateReq.Type = data.Type.ValueStringPointer()
	}
	if !data.Name.Equal(state.Name) {
		updateReq.Name = data.Name.ValueStringPointer()
	}
	if !data.Scheme.Equal(state.Scheme) {
		updateReq.Scheme = data.Scheme.ValueStringPointer()
	}
	if !data.CacheByQueryParam.Equal(state.CacheByQueryParam) {
		updateReq.CacheByQueryParam = data.CacheByQueryParam.ValueBoolPointer()
	}
	if !data.Gzip.Equal(state.Gzip) {
		updateReq.Gzip = data.Gzip.ValueBoolPointer()
	}
	if !data.TTL.Equal(state.TTL) {
		updateReq.TTL = data.TTL.ValueInt32Pointer()
	}
	if !data.MissedTTL.Equal(state.MissedTTL) {
		updateReq.MissedTTL = data.MissedTTL.ValueInt32Pointer()
	}
	if !data.ConnectionTimeout.Equal(state.ConnectionTimeout) {
		updateReq.ConnectionTimeout = data.ConnectionTimeout.ValueInt32Pointer()
	}
	if !data.TimeToFirstByteTimeout.Equal(state.TimeToFirstByteTimeout) {
		updateReq.TimeToFirstByteTimeout = data.TimeToFirstByteTimeout.ValueInt32Pointer()
	}
	if !data.AccessKey.Equal(state.AccessKey) {
		updateReq.AccessKey = data.AccessKey.ValueStringPointer()
	}
	if !data.SecretKey.Equal(state.SecretKey) {
		updateReq.SecretKey = data.SecretKey.ValueStringPointer()
	}
	if !data.Region.Equal(state.Region) {
		updateReq.Region = data.Region.ValueStringPointer()
	}
	if !data.SignatureVersion.Equal(state.SignatureVersion) {
		updateReq.SignatureVersion = data.SignatureVersion.ValueStringPointer()
	}

	if !data.Hostname.Equal(state.Hostname) {
		if data.Type.ValueString() == "WEB" {
			updateReq.Hostname = data.Hostname.ValueStringPointer()
		} else {
			updateReq.Host = data.Hostname.ValueStringPointer()
		}
	}

	origin, err := r.client.Origins.UpdateByID(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating CacheFly Origin",
			"Could not update origin, unexpected error: "+err.Error(),
		)
		return
	}

	r.mapOriginToState(origin, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OriginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.OriginResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Origins.Delete(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting CacheFly Origin",
			"Could not delete origin, unexpected error: "+err.Error(),
		)
		return
	}

}

func (r *OriginResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to map SDK Origin to Terraform state
func (r *OriginResource) mapOriginToState(origin *api.Origin, data *models.OriginResourceModel) {
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
