// internal/provider/resources/log_target.go
package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/cachefly/cachefly-go-sdk/pkg/cachefly"
	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// satisfy framework interfaces.
var (
	_ resource.Resource                = &LogTargetResource{}
	_ resource.ResourceWithImportState = &LogTargetResource{}
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
		MarkdownDescription: "CacheFly Log Target resource. Manages log target configurations for storing access and origin logs.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the log target.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the log target.",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of log target ('S3_BUCKET' | 'ELASTICSEARCH' | 'GOOGLE_BUCKET').",
				Required:    true,
			},
			"endpoint": schema.StringAttribute{
				Description: "Endpoint URL for the log target (for S3 log targets).",
				Optional:    true,
			},
			"region": schema.StringAttribute{
				Description: "Region for the log target (for S3 log targets).",
				Optional:    true,
			},
			"bucket": schema.StringAttribute{
				Description: "Bucket name (for S3 or Google Cloud log targets).",
				Optional:    true,
			},
			"access_key": schema.StringAttribute{
				Description: "Access key (for S3 log targets).",
				Optional:    true,
				Sensitive:   true,
			},
			"secret_key": schema.StringAttribute{
				Description: "Secret key (for S3 log targets).",
				Optional:    true,
				Sensitive:   true,
			},
			"signature_version": schema.StringAttribute{
				Description: "Signature version (for S3 log targets).",
				Optional:    true,
			},
			"json_key": schema.StringAttribute{
				Description: "JSON key (for Google Cloud log targets).",
				Optional:    true,
				Sensitive:   true,
			},
			"hosts": schema.ListAttribute{
				Description: "List of hosts (for Elasticsearch log targets).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"ssl": schema.BoolAttribute{
				Description: "Whether to use SSL/TLS.",
				Optional:    true,
			},
			"ssl_certificate_verification": schema.BoolAttribute{
				Description: "Whether to verify SSL certificates.",
				Optional:    true,
			},
			"index": schema.StringAttribute{
				Description: "Index name (for Elasticsearch log targets).",
				Optional:    true,
			},
			"user": schema.StringAttribute{
				Description: "Username for authentication.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for authentication.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_key": schema.StringAttribute{
				Description: "API key for authentication.",
				Optional:    true,
				Sensitive:   true,
			},
			"access_logs_services": schema.ListAttribute{
				Description: "List of service IDs to enable access logs for.",
				Optional:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				Computed:    true,
			},
			"origin_logs_services": schema.ListAttribute{
				Description: "List of service IDs to enable origin logs for.",
				Optional:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
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

	// Optional fields
	createReq.Name = data.Name.ValueStringPointer()
	createReq.Endpoint = data.Endpoint.ValueStringPointer()
	createReq.Region = data.Region.ValueStringPointer()
	createReq.Bucket = data.Bucket.ValueStringPointer()
	createReq.AccessKey = data.AccessKey.ValueStringPointer()
	createReq.SecretKey = data.SecretKey.ValueStringPointer()
	createReq.SignatureVersion = data.SignatureVersion.ValueStringPointer()
	createReq.JsonKey = data.JsonKey.ValueStringPointer()
	createReq.SSL = data.SSL.ValueBoolPointer()
	createReq.SSLCertificateVerification = data.SSLCertificateVerification.ValueBoolPointer()
	createReq.Index = data.Index.ValueStringPointer()
	createReq.User = data.User.ValueStringPointer()
	createReq.Password = data.Password.ValueStringPointer()
	createReq.ApiKey = data.ApiKey.ValueStringPointer()

	// Handle hosts list
	if !data.Hosts.IsNull() && !data.Hosts.IsUnknown() {
		var hosts []string
		resp.Diagnostics.Append(data.Hosts.ElementsAs(ctx, &hosts, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Hosts = &hosts
	}

	tflog.Debug(ctx, "Creating log target", map[string]interface{}{
		"name": createReq.Name,
		"type": createReq.Type,
	})

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

	// Map response to state
	r.mapLogTargetToState(logTarget, &data)

	// Print the data structure after mapping
	fmt.Printf("Log target data after creation: %+v\n", data)

	tflog.Debug(ctx, "Log target created successfully", map[string]interface{}{
		"log_target_id": logTarget.ID,
		"name":          logTarget.Name,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LogTargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.LogTargetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading log target", map[string]interface{}{
		"log_target_id": data.ID.ValueString(),
	})

	logTarget, err := r.client.LogTargets.GetByID(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading CacheFly Log Target",
			"Could not read log target ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to state
	r.mapLogTargetToState(logTarget, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LogTargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.LogTargetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update request with only changed fields
	updateReq := api.UpdateLogTargetRequest{}

	updateReq.Name = data.Name.ValueStringPointer()
	updateReq.Type = data.Type.ValueStringPointer()
	updateReq.Endpoint = data.Endpoint.ValueStringPointer()
	updateReq.Region = data.Region.ValueStringPointer()
	updateReq.Bucket = data.Bucket.ValueStringPointer()
	updateReq.AccessKey = data.AccessKey.ValueStringPointer()
	updateReq.SecretKey = data.SecretKey.ValueStringPointer()
	updateReq.SignatureVersion = data.SignatureVersion.ValueStringPointer()
	updateReq.JsonKey = data.JsonKey.ValueStringPointer()
	updateReq.SSL = data.SSL.ValueBoolPointer()
	updateReq.SSLCertificateVerification = data.SSLCertificateVerification.ValueBoolPointer()
	updateReq.Index = data.Index.ValueStringPointer()
	updateReq.User = data.User.ValueStringPointer()
	updateReq.Password = data.Password.ValueStringPointer()
	updateReq.ApiKey = data.ApiKey.ValueStringPointer()

	// Handle hosts list
	if !data.Hosts.IsNull() && !data.Hosts.IsUnknown() {
		var hosts []string
		resp.Diagnostics.Append(data.Hosts.ElementsAs(ctx, &hosts, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.Hosts = &hosts
	}

	tflog.Debug(ctx, "Updating log target", map[string]interface{}{
		"log_target_id": data.ID.ValueString(),
	})

	logTarget, err := r.client.LogTargets.UpdateByID(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating CacheFly Log Target",
			"Could not update log target, unexpected error: "+err.Error(),
		)
		return
	}

	// Get current state to compare with planned changes
	var currentState models.LogTargetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &currentState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var needToUpdateLogging = false
	var setLoggingRequest api.SetLoggingRequest

	// Check if AccessLogsServices has changed
	var plannedAccessLogsServices []string
	resp.Diagnostics.Append(data.AccessLogsServices.ElementsAs(ctx, &plannedAccessLogsServices, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current access logs services for comparison
	var currentAccessLogsServices []string
	if !currentState.AccessLogsServices.IsNull() && !currentState.AccessLogsServices.IsUnknown() {
		resp.Diagnostics.Append(currentState.AccessLogsServices.ElementsAs(ctx, &currentAccessLogsServices, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Compare the lists to see if they've changed
	if !r.slicesHaveEqualElements(currentAccessLogsServices, plannedAccessLogsServices) {
		needToUpdateLogging = true
		setLoggingRequest.AccessLogsServices = plannedAccessLogsServices
		tflog.Debug(ctx, "AccessLogsServices changed, will update logging", map[string]interface{}{
			"current_services": currentAccessLogsServices,
			"planned_services": plannedAccessLogsServices,
		})
	} else {
		tflog.Debug(ctx, "AccessLogsServices unchanged, skipping logging update")
	}

	// Check if OriginLogsServices has changed
	var plannedOriginLogsServices []string
	resp.Diagnostics.Append(data.OriginLogsServices.ElementsAs(ctx, &plannedOriginLogsServices, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current origin logs services for comparison
	var currentOriginLogsServices []string
	if !currentState.OriginLogsServices.IsNull() && !currentState.OriginLogsServices.IsUnknown() {
		resp.Diagnostics.Append(currentState.OriginLogsServices.ElementsAs(ctx, &currentOriginLogsServices, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Compare the lists to see if they've changed
	if !r.slicesHaveEqualElements(currentOriginLogsServices, plannedOriginLogsServices) {
		needToUpdateLogging = true
		setLoggingRequest.OriginLogsServices = plannedOriginLogsServices
		tflog.Debug(ctx, "OriginLogsServices changed, will update logging", map[string]interface{}{
			"current_services": currentOriginLogsServices,
			"planned_services": plannedOriginLogsServices,
		})
	} else {
		tflog.Debug(ctx, "OriginLogsServices unchanged, skipping logging update")
	}

	// Update logging if needed
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

	tflog.Debug(ctx, "Deleting log target", map[string]interface{}{
		"log_target_id": data.ID.ValueString(),
	})

	err := r.client.LogTargets.DeleteByID(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting CacheFly Log Target",
			"Could not delete log target, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Log target deleted successfully")
}

func (r *LogTargetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to map SDK LogTarget to Terraform state
func (r *LogTargetResource) mapLogTargetToState(logTarget *api.LogTarget, data *models.LogTargetResourceModel) {
	data.ID = types.StringValue(logTarget.ID)
	data.Name = types.StringPointerValue(logTarget.Name)
	data.Type = types.StringValue(logTarget.Type)
	data.CreatedAt = types.StringValue(logTarget.CreatedAt)
	data.UpdatedAt = types.StringValue(logTarget.UpdatedAt)

	// Handle optional string fields
	data.Endpoint = types.StringPointerValue(logTarget.Endpoint)
	data.Region = types.StringPointerValue(logTarget.Region)
	data.Bucket = types.StringPointerValue(logTarget.Bucket)
	data.AccessKey = types.StringPointerValue(logTarget.AccessKey)
	data.SecretKey = types.StringPointerValue(logTarget.SecretKey)
	data.SignatureVersion = types.StringPointerValue(logTarget.SignatureVersion)
	data.JsonKey = types.StringPointerValue(logTarget.JsonKey)
	data.Index = types.StringPointerValue(logTarget.Index)
	data.User = types.StringPointerValue(logTarget.User)
	data.Password = types.StringPointerValue(logTarget.Password)
	data.ApiKey = types.StringPointerValue(logTarget.ApiKey)

	// Handle boolean fields
	data.SSL = types.BoolPointerValue(logTarget.SSL)
	data.SSLCertificateVerification = types.BoolPointerValue(logTarget.SSLCertificateVerification)

	// Handle hosts list
	if logTarget.Hosts != nil && len(*logTarget.Hosts) > 0 {
		hostElements := make([]attr.Value, len(*logTarget.Hosts))
		for i, host := range *logTarget.Hosts {
			hostElements[i] = types.StringValue(host)
		}
		data.Hosts = types.ListValueMust(types.StringType, hostElements)
	}

	if logTarget.AccessLogsServices != nil && len(*logTarget.AccessLogsServices) > 0 {
		accessLogsServicesElements := make([]attr.Value, len(*logTarget.AccessLogsServices))
		for i, service := range *logTarget.AccessLogsServices {
			accessLogsServicesElements[i] = types.StringValue(service)
		}
		data.AccessLogsServices = types.ListValueMust(types.StringType, accessLogsServicesElements)
	} else {
		data.AccessLogsServices = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if logTarget.OriginLogsServices != nil && len(*logTarget.OriginLogsServices) > 0 {
		originLogsServicesElements := make([]attr.Value, len(*logTarget.OriginLogsServices))
		for i, service := range *logTarget.OriginLogsServices {
			originLogsServicesElements[i] = types.StringValue(service)
		}
		data.OriginLogsServices = types.ListValueMust(types.StringType, originLogsServicesElements)
	} else {
		data.OriginLogsServices = types.ListValueMust(types.StringType, []attr.Value{})
	}
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
