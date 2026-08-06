// internal/provider/resources/log_target.go
package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly"
	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// satisfy framework interfaces.
var (
	_ resource.Resource                   = &LogTargetResource{}
	_ resource.ResourceWithImportState    = &LogTargetResource{}
	_ resource.ResourceWithValidateConfig = &LogTargetResource{}
)

// Per-type field applicability, used to validate configurations before they
// reach the API (the API rejects fields that do not belong to the chosen
// type, but with less helpful error messages).
var (
	logTargetTypeSpecificFields = map[string][]string{
		"S3_BUCKET":     {"endpoint", "region", "bucket", "access_key", "secret_key", "signature_version"},
		"GOOGLE_BUCKET": {"bucket", "json_key"},
		"AZURE_BLOB":    {"endpoint_protocol", "endpoint_suffix", "account_name", "account_key", "container_name", "prefix"},
		"HTTP":          {"uri", "method", "auth", "username", "password", "token"},
	}
	logTargetRequiredFields = map[string][]string{
		"S3_BUCKET":     {"region", "bucket", "access_key", "secret_key"},
		"GOOGLE_BUCKET": {"bucket", "json_key"},
		"AZURE_BLOB":    {"account_name", "account_key", "container_name"},
		"HTTP":          {"uri"},
	}
)

func NewLogTargetResource() resource.Resource {
	return &LogTargetResource{}
}

// LogTargetResource defines the resource implementation.
type LogTargetResource struct {
	client *cachefly.Client
}

func (r *LogTargetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_log_target"
}

func (r *LogTargetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CacheFly Log Target resource. Manages log target configurations for shipping access and origin logs.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the log target.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the log target (minimum 2 characters).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(2),
				},
			},
			"type": schema.StringAttribute{
				Description: "Type of log target ('S3_BUCKET' | 'GOOGLE_BUCKET' | 'AZURE_BLOB' | 'HTTP'). Changing this forces a new log target to be created.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("S3_BUCKET", "GOOGLE_BUCKET", "AZURE_BLOB", "HTTP"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Common log delivery options.
			"format": schema.StringAttribute{
				Description: "Format of the shipped logs ('JSON' | 'NDJSON'). Defaults to 'JSON'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("JSON"),
				Validators: []validator.String{
					stringvalidator.OneOf("JSON", "NDJSON"),
				},
			},
			"compression": schema.StringAttribute{
				Description: "Compression of the shipped logs ('NONE' | 'GZIP' | 'ZSTD'). Defaults to 'NONE'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("NONE"),
				Validators: []validator.String{
					stringvalidator.OneOf("NONE", "GZIP", "ZSTD"),
				},
			},
			"sampling": schema.Int64Attribute{
				Description: "Percentage of logs to ship (0-100). Defaults to 100.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(100),
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
			},

			// S3_BUCKET fields.
			"endpoint": schema.StringAttribute{
				Description: "Endpoint URL (for S3 log targets).",
				Optional:    true,
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "Region (for S3 log targets).",
				Optional:    true,
				Computed:    true,
			},
			"bucket": schema.StringAttribute{
				Description: "Bucket name (for S3 or Google Cloud log targets).",
				Optional:    true,
				Computed:    true,
			},
			"access_key": schema.StringAttribute{
				Description: "Access key (for S3 log targets).",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},
			"secret_key": schema.StringAttribute{
				Description: "Secret key (for S3 log targets).",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},
			"signature_version": schema.StringAttribute{
				Description: "Signature version (for S3 log targets), e.g. 'v4'.",
				Optional:    true,
				Computed:    true,
			},

			// GOOGLE_BUCKET fields.
			"json_key": schema.StringAttribute{
				Description: "Service account JSON key (for Google Cloud log targets).",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},

			// AZURE_BLOB fields.
			"endpoint_protocol": schema.StringAttribute{
				Description: "Endpoint protocol ('HTTP' | 'HTTPS') for Azure Blob log targets. Defaults to 'HTTPS'.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("HTTP", "HTTPS"),
				},
			},
			"endpoint_suffix": schema.StringAttribute{
				Description: "Endpoint suffix (for Azure Blob log targets).",
				Optional:    true,
				Computed:    true,
			},
			"account_name": schema.StringAttribute{
				Description: "Storage account name (for Azure Blob log targets).",
				Optional:    true,
				Computed:    true,
			},
			"account_key": schema.StringAttribute{
				Description: "Storage account key (for Azure Blob log targets).",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},
			"container_name": schema.StringAttribute{
				Description: "Blob container name (for Azure Blob log targets).",
				Optional:    true,
				Computed:    true,
			},
			"prefix": schema.StringAttribute{
				Description: "Path prefix within the container (for Azure Blob log targets).",
				Optional:    true,
				Computed:    true,
			},

			// HTTP fields.
			"uri": schema.StringAttribute{
				Description: "URI logs are shipped to (for HTTP log targets).",
				Optional:    true,
				Computed:    true,
			},
			"method": schema.StringAttribute{
				Description: "HTTP method ('POST' | 'PUT') for HTTP log targets. Defaults to 'POST'.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("POST", "PUT"),
				},
			},
			"auth": schema.StringAttribute{
				Description: "Authentication scheme ('NONE' | 'BASIC' | 'BEARER') for HTTP log targets. Defaults to 'NONE'.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("NONE", "BASIC", "BEARER"),
				},
			},
			"username": schema.StringAttribute{
				Description: "Username for BASIC authentication (for HTTP log targets).",
				Optional:    true,
				Computed:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for BASIC authentication (for HTTP log targets).",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},
			"token": schema.StringAttribute{
				Description: "Token for BEARER authentication (for HTTP log targets).",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},

			"access_logs_services": schema.SetAttribute{
				Description: "List of service IDs to enable access logs for.",
				Optional:    true,
				ElementType: types.StringType,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				Computed:    true,
			},
			"origin_logs_services": schema.SetAttribute{
				Description: "List of service IDs to enable origin logs for.",
				Optional:    true,
				ElementType: types.StringType,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the log target was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the log target was last updated.",
				Computed:    true,
			},
		},
	}
}

// ValidateConfig enforces the per-type field requirements of the log targets
// API: each type has its own set of required fields, and fields belonging to
// other types are rejected.
func (r *LogTargetResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data models.LogTargetResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Type.IsNull() || data.Type.IsUnknown() {
		return
	}
	targetType := data.Type.ValueString()

	typeSpecific, ok := logTargetTypeSpecificFields[targetType]
	if !ok {
		// Unknown type values are reported by the schema validator.
		return
	}

	configValues := map[string]attr.Value{
		"endpoint":          data.Endpoint,
		"region":            data.Region,
		"bucket":            data.Bucket,
		"access_key":        data.AccessKey,
		"secret_key":        data.SecretKey,
		"signature_version": data.SignatureVersion,
		"json_key":          data.JsonKey,
		"endpoint_protocol": data.EndpointProtocol,
		"endpoint_suffix":   data.EndpointSuffix,
		"account_name":      data.AccountName,
		"account_key":       data.AccountKey,
		"container_name":    data.ContainerName,
		"prefix":            data.Prefix,
		"uri":               data.Uri,
		"method":            data.Method,
		"auth":              data.Auth,
		"username":          data.Username,
		"password":          data.Password,
		"token":             data.Token,
	}

	allowed := make(map[string]bool, len(typeSpecific))
	for _, field := range typeSpecific {
		allowed[field] = true
	}

	// Reject fields that belong to other log target types.
	for field, value := range configValues {
		if allowed[field] || value.IsNull() || value.IsUnknown() {
			continue
		}
		resp.Diagnostics.AddAttributeError(
			path.Root(field),
			"Invalid Log Target Attribute",
			fmt.Sprintf("Attribute %q cannot be set for log targets of type %q.", field, targetType),
		)
	}

	// Require the fields the API mandates for the chosen type.
	for _, field := range logTargetRequiredFields[targetType] {
		if configValues[field].IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root(field),
				"Missing Log Target Attribute",
				fmt.Sprintf("Attribute %q is required for log targets of type %q.", field, targetType),
			)
		}
	}
}

func (r *LogTargetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LogTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.LogTargetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build create request
	createReq := api.CreateLogTargetRequest{
		Type: data.Type.ValueString(),
	}

	setString := func(value types.String, target **string) {
		if !value.IsNull() && !value.IsUnknown() {
			*target = value.ValueStringPointer()
		}
	}

	setString(data.Name, &createReq.Name)
	setString(data.Format, &createReq.Format)
	setString(data.Compression, &createReq.Compression)
	setString(data.Endpoint, &createReq.Endpoint)
	setString(data.Region, &createReq.Region)
	setString(data.Bucket, &createReq.Bucket)
	setString(data.AccessKey, &createReq.AccessKey)
	setString(data.SecretKey, &createReq.SecretKey)
	setString(data.SignatureVersion, &createReq.SignatureVersion)
	setString(data.JsonKey, &createReq.JsonKey)
	setString(data.EndpointProtocol, &createReq.EndpointProtocol)
	setString(data.EndpointSuffix, &createReq.EndpointSuffix)
	setString(data.AccountName, &createReq.AccountName)
	setString(data.AccountKey, &createReq.AccountKey)
	setString(data.ContainerName, &createReq.ContainerName)
	setString(data.Prefix, &createReq.Prefix)
	setString(data.Uri, &createReq.Uri)
	setString(data.Method, &createReq.Method)
	setString(data.Auth, &createReq.Auth)
	setString(data.Username, &createReq.Username)
	setString(data.Password, &createReq.Password)
	setString(data.Token, &createReq.Token)

	if !data.Sampling.IsNull() && !data.Sampling.IsUnknown() {
		sampling := int(data.Sampling.ValueInt64())
		createReq.Sampling = &sampling
	}

	logTarget, err := r.client.LogTargets.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating CacheFly Log Target",
			"Could not create log target, unexpected error: "+err.Error(),
		)
		return
	}

	var needToUpdateLogging = false
	var setLoggingRequest api.SetLoggingRequest
	if !data.AccessLogsServices.IsNull() && !data.AccessLogsServices.IsUnknown() {
		needToUpdateLogging = true
		var accessLogsServices []string
		resp.Diagnostics.Append(data.AccessLogsServices.ElementsAs(ctx, &accessLogsServices, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		setLoggingRequest.AccessLogsServices = accessLogsServices
	}

	if !data.OriginLogsServices.IsNull() && !data.OriginLogsServices.IsUnknown() {
		needToUpdateLogging = true
		var originLogsServices []string
		resp.Diagnostics.Append(data.OriginLogsServices.ElementsAs(ctx, &originLogsServices, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		setLoggingRequest.OriginLogsServices = originLogsServices
	}

	if needToUpdateLogging {
		var err error
		logTarget, err = r.client.LogTargets.SetLogging(ctx, logTarget.ID, setLoggingRequest)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Enabling Logging",
				"Could not enable logging, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.mapLogTargetToState(logTarget, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LogTargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.LogTargetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logTarget, err := r.client.LogTargets.GetByID(ctx, data.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "API error 404") {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Error Reading CacheFly Log Target",
				"Could not read log target ID "+data.ID.ValueString()+": "+err.Error(),
			)
		}
		return
	}

	// Map response to state
	r.mapLogTargetToState(logTarget, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LogTargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.LogTargetResourceModel
	var state models.LogTargetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update request with only changed fields. Type changes force a
	// replacement, so type is never part of an update.
	updateReq := api.UpdateLogTargetRequest{}

	setChangedString := func(planned, current types.String, target **string) {
		if !planned.Equal(current) && !planned.IsUnknown() {
			*target = planned.ValueStringPointer()
		}
	}

	setChangedString(data.Name, state.Name, &updateReq.Name)
	setChangedString(data.Format, state.Format, &updateReq.Format)
	setChangedString(data.Compression, state.Compression, &updateReq.Compression)
	setChangedString(data.Endpoint, state.Endpoint, &updateReq.Endpoint)
	setChangedString(data.Region, state.Region, &updateReq.Region)
	setChangedString(data.Bucket, state.Bucket, &updateReq.Bucket)
	setChangedString(data.AccessKey, state.AccessKey, &updateReq.AccessKey)
	setChangedString(data.SecretKey, state.SecretKey, &updateReq.SecretKey)
	setChangedString(data.SignatureVersion, state.SignatureVersion, &updateReq.SignatureVersion)
	setChangedString(data.JsonKey, state.JsonKey, &updateReq.JsonKey)
	setChangedString(data.EndpointProtocol, state.EndpointProtocol, &updateReq.EndpointProtocol)
	setChangedString(data.EndpointSuffix, state.EndpointSuffix, &updateReq.EndpointSuffix)
	setChangedString(data.AccountName, state.AccountName, &updateReq.AccountName)
	setChangedString(data.AccountKey, state.AccountKey, &updateReq.AccountKey)
	setChangedString(data.ContainerName, state.ContainerName, &updateReq.ContainerName)
	setChangedString(data.Prefix, state.Prefix, &updateReq.Prefix)
	setChangedString(data.Uri, state.Uri, &updateReq.Uri)
	setChangedString(data.Method, state.Method, &updateReq.Method)
	setChangedString(data.Auth, state.Auth, &updateReq.Auth)
	setChangedString(data.Username, state.Username, &updateReq.Username)
	setChangedString(data.Password, state.Password, &updateReq.Password)
	setChangedString(data.Token, state.Token, &updateReq.Token)

	if !data.Sampling.Equal(state.Sampling) && !data.Sampling.IsNull() && !data.Sampling.IsUnknown() {
		sampling := int(data.Sampling.ValueInt64())
		updateReq.Sampling = &sampling
	}

	logTarget, err := r.client.LogTargets.UpdateByID(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating CacheFly Log Target",
			"Could not update log target, unexpected error: "+err.Error(),
		)
		return
	}

	var needToUpdateLogging = false
	var setLoggingRequest api.SetLoggingRequest

	var plannedAccessLogsServices []string
	resp.Diagnostics.Append(data.AccessLogsServices.ElementsAs(ctx, &plannedAccessLogsServices, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var currentAccessLogsServices []string
	if !state.AccessLogsServices.IsNull() && !state.AccessLogsServices.IsUnknown() {
		resp.Diagnostics.Append(state.AccessLogsServices.ElementsAs(ctx, &currentAccessLogsServices, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !r.slicesHaveEqualElements(currentAccessLogsServices, plannedAccessLogsServices) {
		needToUpdateLogging = true
		setLoggingRequest.AccessLogsServices = plannedAccessLogsServices
	}

	var plannedOriginLogsServices []string
	resp.Diagnostics.Append(data.OriginLogsServices.ElementsAs(ctx, &plannedOriginLogsServices, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var currentOriginLogsServices []string
	if !state.OriginLogsServices.IsNull() && !state.OriginLogsServices.IsUnknown() {
		resp.Diagnostics.Append(state.OriginLogsServices.ElementsAs(ctx, &currentOriginLogsServices, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !r.slicesHaveEqualElements(currentOriginLogsServices, plannedOriginLogsServices) {
		needToUpdateLogging = true
		setLoggingRequest.OriginLogsServices = plannedOriginLogsServices
	}

	if needToUpdateLogging {
		var err error
		logTarget, err = r.client.LogTargets.SetLogging(ctx, data.ID.ValueString(), setLoggingRequest)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Enabling Logging",
				"Could not enable logging, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.mapLogTargetToState(logTarget, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LogTargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.LogTargetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	setLoggingRequest := api.SetLoggingRequest{
		AccessLogsServices: []string{},
		OriginLogsServices: []string{},
	}

	if _, err := r.client.LogTargets.SetLogging(ctx, data.ID.ValueString(), setLoggingRequest); err != nil {
		resp.Diagnostics.AddError(
			"Error Disabling Logging for Log Target",
			"Could not disable logging prior to deletion: "+err.Error(),
		)
		return
	}

	err := r.client.LogTargets.DeleteByID(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting CacheFly Log Target",
			"Could not delete log target, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *LogTargetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// stringFromAPI returns the value reported by the API. When the API omits the
// field (e.g. secrets that are not echoed back, or fields that do not apply
// to the log target's type), the prior known value is preserved so that
// apply results stay consistent with the plan.
func stringFromAPI(apiValue *string, prior types.String) types.String {
	if apiValue != nil {
		return types.StringValue(*apiValue)
	}
	if prior.IsUnknown() {
		return types.StringNull()
	}
	return prior
}

// int64FromAPI is the int equivalent of stringFromAPI.
func int64FromAPI(apiValue *int, prior types.Int64) types.Int64 {
	if apiValue != nil {
		return types.Int64Value(int64(*apiValue))
	}
	if prior.IsUnknown() {
		return types.Int64Null()
	}
	return prior
}

// servicesSetFromAPI converts a services list reported by the API to a set.
// When the API does not report the list, the prior known value is preserved
// (the current API does not document these lists in its responses).
func servicesSetFromAPI(apiValue *[]string, prior types.Set) types.Set {
	if apiValue != nil {
		elements := make([]attr.Value, len(*apiValue))
		for i, service := range *apiValue {
			elements[i] = types.StringValue(service)
		}
		return types.SetValueMust(types.StringType, elements)
	}
	if prior.IsNull() || prior.IsUnknown() {
		return types.SetValueMust(types.StringType, []attr.Value{})
	}
	return prior
}

// Helper function to map SDK LogTarget to Terraform state
func (r *LogTargetResource) mapLogTargetToState(logTarget *api.LogTarget, data *models.LogTargetResourceModel) {
	data.ID = types.StringValue(logTarget.ID)
	data.Type = types.StringValue(logTarget.Type)
	data.CreatedAt = types.StringValue(logTarget.CreatedAt)
	data.UpdatedAt = types.StringValue(logTarget.UpdatedAt)

	data.Name = stringFromAPI(logTarget.Name, data.Name)

	// Common log delivery options.
	data.Format = stringFromAPI(logTarget.Format, data.Format)
	data.Compression = stringFromAPI(logTarget.Compression, data.Compression)
	data.Sampling = int64FromAPI(logTarget.Sampling, data.Sampling)

	// S3_BUCKET fields.
	data.Endpoint = stringFromAPI(logTarget.Endpoint, data.Endpoint)
	data.Region = stringFromAPI(logTarget.Region, data.Region)
	data.Bucket = stringFromAPI(logTarget.Bucket, data.Bucket)
	data.AccessKey = stringFromAPI(logTarget.AccessKey, data.AccessKey)
	data.SecretKey = stringFromAPI(logTarget.SecretKey, data.SecretKey)
	data.SignatureVersion = stringFromAPI(logTarget.SignatureVersion, data.SignatureVersion)

	// GOOGLE_BUCKET fields.
	data.JsonKey = stringFromAPI(logTarget.JsonKey, data.JsonKey)

	// AZURE_BLOB fields.
	data.EndpointProtocol = stringFromAPI(logTarget.EndpointProtocol, data.EndpointProtocol)
	data.EndpointSuffix = stringFromAPI(logTarget.EndpointSuffix, data.EndpointSuffix)
	data.AccountName = stringFromAPI(logTarget.AccountName, data.AccountName)
	data.AccountKey = stringFromAPI(logTarget.AccountKey, data.AccountKey)
	data.ContainerName = stringFromAPI(logTarget.ContainerName, data.ContainerName)
	data.Prefix = stringFromAPI(logTarget.Prefix, data.Prefix)

	// HTTP fields.
	data.Uri = stringFromAPI(logTarget.Uri, data.Uri)
	data.Method = stringFromAPI(logTarget.Method, data.Method)
	data.Auth = stringFromAPI(logTarget.Auth, data.Auth)
	data.Username = stringFromAPI(logTarget.Username, data.Username)
	data.Password = stringFromAPI(logTarget.Password, data.Password)
	data.Token = stringFromAPI(logTarget.Token, data.Token)

	// Services logging lists.
	data.AccessLogsServices = servicesSetFromAPI(logTarget.AccessLogsServices, data.AccessLogsServices)
	data.OriginLogsServices = servicesSetFromAPI(logTarget.OriginLogsServices, data.OriginLogsServices)
}

// slicesHaveEqualElements compares two string slices for equality
// Returns true if they are equal, false otherwise
func (r *LogTargetResource) slicesHaveEqualElements(current, planned []string) bool {
	if len(current) != len(planned) {
		return false
	}

	// Create maps to count occurrences for order-independent comparison
	currentMap := make(map[string]int)
	plannedMap := make(map[string]int)

	for _, item := range current {
		currentMap[item]++
	}

	for _, item := range planned {
		plannedMap[item]++
	}

	// Compare the maps
	if len(currentMap) != len(plannedMap) {
		return false
	}

	for key, currentCount := range currentMap {
		plannedCount, exists := plannedMap[key]
		if !exists || currentCount != plannedCount {
			return false
		}
	}

	return true
}
