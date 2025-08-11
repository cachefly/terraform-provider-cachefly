// internal/provider/datasources/service_domain.go
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
var _ datasource.DataSource = &ServiceDomainDataSource{}

func NewServiceDomainDataSource() datasource.DataSource {
	return &ServiceDomainDataSource{}
}

// ServiceDomainDataSource defines the data source implementation.
type ServiceDomainDataSource struct {
	client *cachefly.Client
}

func (d *ServiceDomainDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_domain"
}

func (d *ServiceDomainDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CacheFly Service Domain data source. Look up a specific domain attached to a CacheFly service.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the service domain.",
				Required:    true,
			},
			"service_id": schema.StringAttribute{
				Description: "The ID of the service this domain belongs to.",
				Required:    true,
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
			"response_type": schema.StringAttribute{
				Description: "Optional response type parameter for the API call.",
				Optional:    true,
			},
		},
	}
}

func (d *ServiceDomainDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.ServiceDomainDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := d.client.ServiceDomains.GetByID(
		ctx,
		data.ServiceID.ValueString(),
		data.ID.ValueString(),
		data.ResponseType.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading CacheFly Service Domain",
			"Could not read service domain ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to data model
	d.mapDomainToDataSource(domain, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper function to map SDK ServiceDomain to data source model
func (d *ServiceDomainDataSource) mapDomainToDataSource(domain *api.ServiceDomain, data *models.ServiceDomainDataSourceModel) {
	data.ID = types.StringValue(domain.ID)
	data.ServiceID = types.StringValue(domain.Service)
	data.Name = types.StringValue(domain.Name)
	data.Description = types.StringValue(domain.Description)
	data.ValidationMode = types.StringValue(domain.ValidationMode)
	data.ValidationTarget = types.StringValue(domain.ValidationTarget)
	data.ValidationStatus = types.StringValue(domain.ValidationStatus)
	data.CreatedAt = types.StringValue(domain.CreatedAt)
	data.UpdatedAt = types.StringValue(domain.UpdatedAt)

	// Convert certificates slice to Terraform set
	if len(domain.Certificates) > 0 {
		certElements := make([]attr.Value, len(domain.Certificates))
		for i, cert := range domain.Certificates {
			certElements[i] = types.StringValue(cert)
		}
		data.Certificates, _ = types.SetValue(types.StringType, certElements)
	} else {
		data.Certificates = types.SetNull(types.StringType)
	}
}
