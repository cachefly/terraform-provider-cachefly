// internal/provider/datasources/log_targets.go
package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/cachefly/cachefly-go-sdk/pkg/cachefly"
	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"

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
				Description: "Filter log targets by type ('S3_BUCKET' | 'ELASTICSEARCH' | 'GOOGLE_BUCKET').",
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
							Description: "Type of log target.",
							Computed:    true,
						},
						"endpoint": schema.StringAttribute{
							Description: "Endpoint URL for the log target (for S3 log targets).",
							Computed:    true,
						},
						"region": schema.StringAttribute{
							Description: "Region for the log target (for S3 log targets).",
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
							Description: "JSON key (for Google Cloud log targets).",
							Computed:    true,
							Sensitive:   true,
						},
						"hosts": schema.SetAttribute{
							Description: "List of hosts (for Elasticsearch log targets).",
							Computed:    true,
							ElementType: types.StringType,
						},
						"ssl": schema.BoolAttribute{
							Description: "Whether to use SSL/TLS.",
							Computed:    true,
						},
						"ssl_certificate_verification": schema.BoolAttribute{
							Description: "Whether to verify SSL certificates.",
							Computed:    true,
						},
						"index": schema.StringAttribute{
							Description: "Index name (for Elasticsearch log targets).",
							Computed:    true,
						},
						"user": schema.StringAttribute{
							Description: "Username for authentication.",
							Computed:    true,
						},
						"password": schema.StringAttribute{
							Description: "Password for authentication.",
							Computed:    true,
							Sensitive:   true,
						},
						"api_key": schema.StringAttribute{
							Description: "API key for authentication.",
							Computed:    true,
							Sensitive:   true,
						},
						"access_logs_services": schema.SetAttribute{
							Description: "List of service IDs to enable access logs for.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"origin_logs_services": schema.SetAttribute{
							Description: "List of service IDs to enable origin logs for.",
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
		tflog.Info(ctx, "Log targets", map[string]interface{}{
			"log_targets": pageResp.LogTargets,
		})
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
		"id":                           types.StringType,
		"name":                         types.StringType,
		"type":                         types.StringType,
		"endpoint":                     types.StringType,
		"region":                       types.StringType,
		"bucket":                       types.StringType,
		"access_key":                   types.StringType,
		"secret_key":                   types.StringType,
		"signature_version":            types.StringType,
		"json_key":                     types.StringType,
		"hosts":                        types.SetType{ElemType: types.StringType},
		"ssl":                          types.BoolType,
		"ssl_certificate_verification": types.BoolType,
		"index":                        types.StringType,
		"user":                         types.StringType,
		"password":                     types.StringType,
		"api_key":                      types.StringType,
		"access_logs_services":         types.SetType{ElemType: types.StringType},
		"origin_logs_services":         types.SetType{ElemType: types.StringType},
		"created_at":                   types.StringType,
		"updated_at":                   types.StringType,
	}

	// Map response to Terraform values
	items := make([]attr.Value, len(allLogTargets))
	for i, lt := range allLogTargets {
		// Convert hosts slice pointer to set
		var hostsSet types.Set
		if lt.Hosts != nil && len(*lt.Hosts) > 0 {
			hostElements := make([]attr.Value, len(*lt.Hosts))
			for j, host := range *lt.Hosts {
				hostElements[j] = types.StringValue(host)
			}
			hostsSet, _ = types.SetValue(types.StringType, hostElements)
		} else {
			hostsSet = types.SetNull(types.StringType)
		}

		// Convert access logs services
		var accessLogsSet types.Set
		if lt.AccessLogsServices != nil && len(*lt.AccessLogsServices) > 0 {
			elems := make([]attr.Value, len(*lt.AccessLogsServices))
			for j, svc := range *lt.AccessLogsServices {
				elems[j] = types.StringValue(svc)
			}
			accessLogsSet, _ = types.SetValue(types.StringType, elems)
		} else {
			accessLogsSet = types.SetNull(types.StringType)
		}

		// Convert origin logs services
		var originLogsSet types.Set
		if lt.OriginLogsServices != nil && len(*lt.OriginLogsServices) > 0 {
			elems := make([]attr.Value, len(*lt.OriginLogsServices))
			for j, svc := range *lt.OriginLogsServices {
				elems[j] = types.StringValue(svc)
			}
			originLogsSet, _ = types.SetValue(types.StringType, elems)
		} else {
			originLogsSet = types.SetNull(types.StringType)
		}

		obj, _ := types.ObjectValue(
			objectAttrTypes,
			map[string]attr.Value{
				"id":                           types.StringValue(lt.ID),
				"name":                         types.StringPointerValue(lt.Name),
				"type":                         types.StringValue(lt.Type),
				"endpoint":                     types.StringPointerValue(lt.Endpoint),
				"region":                       types.StringPointerValue(lt.Region),
				"bucket":                       types.StringPointerValue(lt.Bucket),
				"access_key":                   types.StringPointerValue(lt.AccessKey),
				"secret_key":                   types.StringPointerValue(lt.SecretKey),
				"signature_version":            types.StringPointerValue(lt.SignatureVersion),
				"json_key":                     types.StringPointerValue(lt.JsonKey),
				"hosts":                        hostsSet,
				"ssl":                          types.BoolPointerValue(lt.SSL),
				"ssl_certificate_verification": types.BoolPointerValue(lt.SSLCertificateVerification),
				"index":                        types.StringPointerValue(lt.Index),
				"user":                         types.StringPointerValue(lt.User),
				"password":                     types.StringPointerValue(lt.Password),
				"api_key":                      types.StringPointerValue(lt.ApiKey),
				"access_logs_services":         accessLogsSet,
				"origin_logs_services":         originLogsSet,
				"created_at":                   types.StringValue(lt.CreatedAt),
				"updated_at":                   types.StringValue(lt.UpdatedAt),
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
