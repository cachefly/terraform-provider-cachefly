// internal/provider/datasources/service_domains.go
package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cachefly/cachefly-go-sdk/pkg/cachefly"
	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ServiceDomainsDataSource{}

func NewServiceDomainsDataSource() datasource.DataSource {
	return &ServiceDomainsDataSource{}
}

// ServiceDomainsDataSource defines the data source implementation.
type ServiceDomainsDataSource struct {
	client *cachefly.Client
}

func (d *ServiceDomainsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_domains"
}

func (d *ServiceDomainsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CacheFly Service Domains data source. List all domains attached to a CacheFly service.",

		Attributes: map[string]schema.Attribute{
			"service_id": schema.StringAttribute{
				Description: "The ID of the service to list domains for.",
				Required:    true,
			},
			"search": schema.StringAttribute{
				Description: "Optional search term to filter domains.",
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
			"domains": schema.ListNestedAttribute{
				Description: "List of domains for the service.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the service domain.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The domain name.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the domain.",
							Computed:    true,
						},
						"validation_mode": schema.StringAttribute{
							Description: "Domain validation mode.",
							Computed:    true,
						},
						"validation_target": schema.StringAttribute{
							Description: "The validation target.",
							Computed:    true,
						},
						"validation_status": schema.StringAttribute{
							Description: "The current validation status of the domain.",
							Computed:    true,
						},
						"certificates": schema.SetAttribute{
							Description: "List of certificate IDs associated with this domain.",
							ElementType: types.StringType,
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "When the domain was created.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "When the domain was last updated.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ServiceDomainsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceDomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.ServiceDomainsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := api.ListServiceDomainsOptions{
		Search:       data.Search.ValueString(),
		ResponseType: data.ResponseType.ValueString(),
	}

	if !data.Offset.IsNull() {
		opts.Offset = int(data.Offset.ValueInt64())
	}
	if !data.Limit.IsNull() {
		opts.Limit = int(data.Limit.ValueInt64())
	}

	// Get the domains
	domainsResp, err := d.client.ServiceDomains.List(ctx, data.ServiceID.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading CacheFly Service Domains",
			"Could not read service domains for service "+data.ServiceID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to data model
	domains := make([]attr.Value, len(domainsResp.Domains))
	for i, domain := range domainsResp.Domains {
		// Convert certificates to list
		var certsList types.Set
		if len(domain.Certificates) > 0 {
			certElements := make([]attr.Value, len(domain.Certificates))
			for j, cert := range domain.Certificates {
				certElements[j] = types.StringValue(cert)
			}
			certsList, _ = types.SetValue(types.StringType, certElements)
		} else {
			certsList = types.SetNull(types.StringType)
		}

		domainObj, _ := types.ObjectValue(
			map[string]attr.Type{
				"id":                types.StringType,
				"name":              types.StringType,
				"description":       types.StringType,
				"validation_mode":   types.StringType,
				"validation_target": types.StringType,
				"validation_status": types.StringType,
				"certificates":      types.SetType{ElemType: types.StringType},
				"created_at":        types.StringType,
				"updated_at":        types.StringType,
			},
			map[string]attr.Value{
				"id":                types.StringValue(domain.ID),
				"name":              types.StringValue(domain.Name),
				"description":       types.StringValue(domain.Description),
				"validation_mode":   types.StringValue(domain.ValidationMode),
				"validation_target": types.StringValue(domain.ValidationTarget),
				"validation_status": types.StringValue(domain.ValidationStatus),
				"certificates":      certsList,
				"created_at":        types.StringValue(domain.CreatedAt),
				"updated_at":        types.StringValue(domain.UpdatedAt),
			},
		)
		domains[i] = domainObj
	}

	domainsSet, diags := types.SetValue(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":                types.StringType,
				"name":              types.StringType,
				"description":       types.StringType,
				"validation_mode":   types.StringType,
				"validation_target": types.StringType,
				"validation_status": types.StringType,
				"certificates":      types.SetType{ElemType: types.StringType},
				"created_at":        types.StringType,
				"updated_at":        types.StringType,
			},
		},
		domains,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Domains = domainsSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
