package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/cachefly/cachefly-go-sdk/pkg/cachefly"
	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &ServiceDataSource{}
	_ datasource.DataSourceWithConfigure = &ServiceDataSource{}
)

func NewServiceDataSource() datasource.DataSource {
	return &ServiceDataSource{}
}

// ServiceDataSource defines the data source implementation.
type ServiceDataSource struct {
	client *cachefly.Client
}

func (d *ServiceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (d *ServiceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about an existing CacheFly CDN service.",
		Attributes: map[string]schema.Attribute{
			// Input attributes (one of these is required)
			"id": schema.StringAttribute{
				Description: "The unique identifier of the service. Either 'id' or 'unique_name' must be specified.",
				Optional:    true,
			},
			"unique_name": schema.StringAttribute{
				Description: "The unique name of the service. Either 'id' or 'unique_name' must be specified.",
				Optional:    true,
			},

			// Get method options (optional)
			"response_type": schema.StringAttribute{
				Description: "The response type for the API call. Controls the level of detail returned.",
				Optional:    true,
			},
			"include_features": schema.BoolAttribute{
				Description: "Whether to include features in the response.",
				Optional:    true,
			},

			// Output attributes (computed)
			"name": schema.StringAttribute{
				Description: "The display name of the service.",
				Computed:    true,
			},
			"auto_ssl": schema.BoolAttribute{
				Description: "Whether SSL is automatically provisioned for this service.",
				Computed:    true,
			},
			"configuration_mode": schema.StringAttribute{
				Description: "The configuration mode of the service.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the service.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the service was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the service was last updated.",
				Computed:    true,
			},
		},
	}
}

func (d *ServiceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (d *ServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.ServiceDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either ID or UniqueName is provided
	hasID := !data.ID.IsNull() && !data.ID.IsUnknown()
	hasUniqueName := !data.UniqueName.IsNull() && !data.UniqueName.IsUnknown()

	if !hasID && !hasUniqueName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'unique_name' must be specified to look up a service.",
		)
		return
	}

	if hasID && hasUniqueName {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Only one of 'id' or 'unique_name' should be specified, not both.",
		)
		return
	}

	var service *api.Service
	var err error

	if hasID {

		serviceID := data.ID.ValueString()

		// Check if we should use Get() with options or GetByID()
		hasOptions := !data.ResponseType.IsNull() || !data.IncludeFeatures.IsNull()

		if hasOptions {
			// Use Get() method with options
			responseType := ""
			if !data.ResponseType.IsNull() {
				responseType = data.ResponseType.ValueString()
			}

			includeFeatures := false
			if !data.IncludeFeatures.IsNull() {
				includeFeatures = data.IncludeFeatures.ValueBool()
			}

			tflog.Debug(ctx, "Looking up service by ID with options", map[string]interface{}{
				"id":               serviceID,
				"response_type":    responseType,
				"include_features": includeFeatures,
			})

			service, err = d.client.Services.Get(ctx, serviceID, responseType, includeFeatures)
		} else {
			// Use simple GetByID() method
			tflog.Debug(ctx, "Looking up service by ID", map[string]interface{}{
				"id": serviceID,
			})

			service, err = d.client.Services.GetByID(ctx, serviceID)
		}

		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading CacheFly Service",
				"Could not read service by ID "+serviceID+": "+err.Error(),
			)
			return
		}

	} else {
		// Look up by unique name - we need to list services and filter
		uniqueName := data.UniqueName.ValueString()

		tflog.Debug(ctx, "Looking up service by unique name", map[string]interface{}{
			"unique_name": uniqueName,
		})

		// List all services and find the one with matching unique name
		listOptions := api.ListOptions{
			Limit: 100, // Get a large batch to avoid pagination issues
		}

		listResp, err := d.client.Services.List(ctx, listOptions)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing CacheFly Services",
				"Could not list services to find by unique name: "+err.Error(),
			)
			return
		}

		// Find service with matching unique name
		found := false
		for _, svc := range listResp.Services {
			if svc.UniqueName == uniqueName {
				service = &svc
				found = true
				break
			}
		}

		if !found {
			resp.Diagnostics.AddError(
				"Service Not Found",
				fmt.Sprintf("Could not find service with unique name: %s", uniqueName),
			)
			return
		}
	}

	// Map the service data to our model
	data.ID = types.StringValue(service.ID)
	data.UniqueName = types.StringValue(service.UniqueName)
	data.Name = types.StringValue(service.Name)
	data.AutoSSL = types.BoolValue(service.AutoSSL)
	data.ConfigurationMode = types.StringValue(service.ConfigurationMode)
	data.Status = types.StringValue(service.Status)
	data.CreatedAt = types.StringValue(service.CreatedAt)
	data.UpdatedAt = types.StringValue(service.UpdatedAt)

	tflog.Debug(ctx, "Successfully read service data", map[string]interface{}{
		"id":          service.ID,
		"unique_name": service.UniqueName,
		"status":      service.Status,
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
