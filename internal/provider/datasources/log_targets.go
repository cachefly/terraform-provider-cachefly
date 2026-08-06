// internal/provider/datasources/log_targets.go
package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly"
	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &LogTargetsDataSource{}

// NewLogTargetsDataSource is a helper constructor.
func NewLogTargetsDataSource() datasource.DataSource {
	return &LogTargetsDataSource{}
}

// LogTargetsDataSource defines the data source implementation.
type LogTargetsDataSource struct {
	client *cachefly.Client
}

func (d *LogTargetsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_log_targets"
}

func (d *LogTargetsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CacheFly Log Targets data source. List log targets for access and origin logs.",

		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Description: "Filter log targets by type ('S3_BUCKET' | 'GOOGLE_BUCKET' | 'AZURE_BLOB' | 'HTTP').",
				Optional:    true,
			},
			"offset": schema.Int64Attribute{
				Description: "Offset for pagination (default: 0).",
				Optional:    true,
			},
			"limit": schema.Int64Attribute{
				Description: "Limit for pagination (default: API default).",
				Optional:    true,
			},
			"response_type": schema.StringAttribute{
				Description: "Optional response type parameter for the API call.",
				Optional:    true,
			},
			"log_targets": schema.ListNestedAttribute{
				Description: "List of log targets.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the log target.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the log target.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Type of log target ('S3_BUCKET' | 'GOOGLE_BUCKET' | 'AZURE_BLOB' | 'HTTP' | 'MANUAL').",
							Computed:    true,
						},
						"format": schema.StringAttribute{
							Description: "Format of the shipped logs ('JSON' | 'NDJSON').",
							Computed:    true,
						},
						"compression": schema.StringAttribute{
							Description: "Compression of the shipped logs ('NONE' | 'GZIP' | 'ZSTD').",
							Computed:    true,
						},
						"sampling": schema.Int64Attribute{
							Description: "Percentage of logs to ship (0-100).",
							Computed:    true,
						},
						"endpoint": schema.StringAttribute{
							Description: "Endpoint URL (for S3 log targets).",
							Computed:    true,
						},
						"region": schema.StringAttribute{
							Description: "Region (for S3 log targets).",
							Computed:    true,
						},
						"bucket": schema.StringAttribute{
							Description: "Bucket name (for S3 or Google Cloud log targets).",
							Computed:    true,
						},
						"access_key": schema.StringAttribute{
							Description: "Access key (for S3 log targets).",
							Computed:    true,
							Sensitive:   true,
						},
						"secret_key": schema.StringAttribute{
							Description: "Secret key (for S3 log targets).",
							Computed:    true,
							Sensitive:   true,
						},
						"signature_version": schema.StringAttribute{
							Description: "Signature version (for S3 log targets).",
							Computed:    true,
						},
						"json_key": schema.StringAttribute{
							Description: "Service account JSON key (for Google Cloud log targets).",
							Computed:    true,
							Sensitive:   true,
						},
						"endpoint_protocol": schema.StringAttribute{
							Description: "Endpoint protocol ('HTTP' | 'HTTPS') for Azure Blob log targets.",
							Computed:    true,
						},
						"endpoint_suffix": schema.StringAttribute{
							Description: "Endpoint suffix (for Azure Blob log targets).",
							Computed:    true,
						},
						"account_name": schema.StringAttribute{
							Description: "Storage account name (for Azure Blob log targets).",
							Computed:    true,
						},
						"account_key": schema.StringAttribute{
							Description: "Storage account key (for Azure Blob log targets).",
							Computed:    true,
							Sensitive:   true,
						},
						"container_name": schema.StringAttribute{
							Description: "Blob container name (for Azure Blob log targets).",
							Computed:    true,
						},
						"prefix": schema.StringAttribute{
							Description: "Path prefix within the container (for Azure Blob log targets).",
							Computed:    true,
						},
						"uri": schema.StringAttribute{
							Description: "URI logs are shipped to (for HTTP log targets).",
							Computed:    true,
						},
						"method": schema.StringAttribute{
							Description: "HTTP method ('POST' | 'PUT') for HTTP log targets.",
							Computed:    true,
						},
						"auth": schema.StringAttribute{
							Description: "Authentication scheme ('NONE' | 'BASIC' | 'BEARER') for HTTP log targets.",
							Computed:    true,
						},
						"username": schema.StringAttribute{
							Description: "Username for BASIC authentication (for HTTP log targets).",
							Computed:    true,
						},
						"password": schema.StringAttribute{
							Description: "Password for BASIC authentication (for HTTP log targets).",
							Computed:    true,
							Sensitive:   true,
						},
						"token": schema.StringAttribute{
							Description: "Token for BEARER authentication (for HTTP log targets).",
							Computed:    true,
							Sensitive:   true,
						},
						"access_logs_services": schema.SetAttribute{
							Description: "List of service IDs with access logs enabled (when reported by the API).",
							Computed:    true,
							ElementType: types.StringType,
						},
						"origin_logs_services": schema.SetAttribute{
							Description: "List of service IDs with origin logs enabled (when reported by the API).",
							Computed:    true,
							ElementType: types.StringType,
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
				},
			},
		},
	}
}

func (d *LogTargetsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LogTargetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.LogTargetsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build list options (we will paginate to fetch all pages)
	opts := api.ListLogTargetsOptions{
		Type:         data.Type.ValueString(),
		ResponseType: data.ResponseType.ValueString(),
	}

	// Starting offset
	if !data.Offset.IsNull() {
		opts.Offset = int(data.Offset.ValueInt64())
	}

	// Per-page limit; if not provided, use a sane default
	if !data.Limit.IsNull() {
		opts.Limit = int(data.Limit.ValueInt64())
	}
	if opts.Limit <= 0 {
		opts.Limit = 100
	}

	// Accumulate all pages
	var allLogTargets []api.LogTarget
	for {
		pageResp, err := d.client.LogTargets.List(ctx, opts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading CacheFly Log Targets",
				"Could not read log targets: "+err.Error(),
			)
			return
		}

		allLogTargets = append(allLogTargets, pageResp.LogTargets...)

		// Advance offset; break when we've fetched all
		fetched := len(pageResp.LogTargets)
		total := pageResp.Meta.Count
		opts.Offset += fetched

		if fetched == 0 || opts.Offset >= total {
			break
		}
	}

	// Prepare attribute type map for each object in the list
	objectAttrTypes := map[string]attr.Type{
		"id":                   types.StringType,
		"name":                 types.StringType,
		"type":                 types.StringType,
		"format":               types.StringType,
		"compression":          types.StringType,
		"sampling":             types.Int64Type,
		"endpoint":             types.StringType,
		"region":               types.StringType,
		"bucket":               types.StringType,
		"access_key":           types.StringType,
		"secret_key":           types.StringType,
		"signature_version":    types.StringType,
		"json_key":             types.StringType,
		"endpoint_protocol":    types.StringType,
		"endpoint_suffix":      types.StringType,
		"account_name":         types.StringType,
		"account_key":          types.StringType,
		"container_name":       types.StringType,
		"prefix":               types.StringType,
		"uri":                  types.StringType,
		"method":               types.StringType,
		"auth":                 types.StringType,
		"username":             types.StringType,
		"password":             types.StringType,
		"token":                types.StringType,
		"access_logs_services": types.SetType{ElemType: types.StringType},
		"origin_logs_services": types.SetType{ElemType: types.StringType},
		"created_at":           types.StringType,
		"updated_at":           types.StringType,
	}

	// Map response to Terraform values
	items := make([]attr.Value, len(allLogTargets))
	for i, lt := range allLogTargets {
		// Convert sampling to Int64
		var samplingValue types.Int64
		if lt.Sampling != nil {
			samplingValue = types.Int64Value(int64(*lt.Sampling))
		} else {
			samplingValue = types.Int64Null()
		}

		// Convert services lists (null when the API does not report them)
		accessLogsSet := servicesSetOrNull(lt.AccessLogsServices)
		originLogsSet := servicesSetOrNull(lt.OriginLogsServices)

		obj, _ := types.ObjectValue(
			objectAttrTypes,
			map[string]attr.Value{
				"id":                   types.StringValue(lt.ID),
				"name":                 types.StringPointerValue(lt.Name),
				"type":                 types.StringValue(lt.Type),
				"format":               types.StringPointerValue(lt.Format),
				"compression":          types.StringPointerValue(lt.Compression),
				"sampling":             samplingValue,
				"endpoint":             types.StringPointerValue(lt.Endpoint),
				"region":               types.StringPointerValue(lt.Region),
				"bucket":               types.StringPointerValue(lt.Bucket),
				"access_key":           types.StringPointerValue(lt.AccessKey),
				"secret_key":           types.StringPointerValue(lt.SecretKey),
				"signature_version":    types.StringPointerValue(lt.SignatureVersion),
				"json_key":             types.StringPointerValue(lt.JsonKey),
				"endpoint_protocol":    types.StringPointerValue(lt.EndpointProtocol),
				"endpoint_suffix":      types.StringPointerValue(lt.EndpointSuffix),
				"account_name":         types.StringPointerValue(lt.AccountName),
				"account_key":          types.StringPointerValue(lt.AccountKey),
				"container_name":       types.StringPointerValue(lt.ContainerName),
				"prefix":               types.StringPointerValue(lt.Prefix),
				"uri":                  types.StringPointerValue(lt.Uri),
				"method":               types.StringPointerValue(lt.Method),
				"auth":                 types.StringPointerValue(lt.Auth),
				"username":             types.StringPointerValue(lt.Username),
				"password":             types.StringPointerValue(lt.Password),
				"token":                types.StringPointerValue(lt.Token),
				"access_logs_services": accessLogsSet,
				"origin_logs_services": originLogsSet,
				"created_at":           types.StringValue(lt.CreatedAt),
				"updated_at":           types.StringValue(lt.UpdatedAt),
			},
		)
		items[i] = obj
	}

	listValue, diags := types.ListValue(
		types.ObjectType{AttrTypes: objectAttrTypes},
		items,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.LogTargets = listValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// servicesSetOrNull converts a services list to a set value, or a null set
// when the API does not report the list.
func servicesSetOrNull(services *[]string) types.Set {
	if services == nil {
		return types.SetNull(types.StringType)
	}
	elements := make([]attr.Value, len(*services))
	for i, service := range *services {
		elements[i] = types.StringValue(service)
	}
	return types.SetValueMust(types.StringType, elements)
}
