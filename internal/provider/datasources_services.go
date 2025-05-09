package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/avvvet/cachefly-sdk-go/pkg/cachefly"
	"github.com/avvvet/cachefly-sdk-go/pkg/cachefly/api"
)

type servicesDataSource struct {
	client *cachefly.Client
}

func (d *servicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_services"
}

func (d *servicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a list of CacheFly services.",
		Attributes: map[string]schema.Attribute{
			"response_type": schema.StringAttribute{
				Optional:    true,
				Description: "The response type for the services list (e.g., 'shallow'). Defaults to 'shallow'.",
			},
			"include_features": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to include features in the response. Defaults to false.",
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Description: "Filter services by status (e.g., 'ACTIVE').",
			},
			"offset": schema.Int64Attribute{
				Optional:    true,
				Description: "The offset for pagination. Defaults to 0.",
			},
			"limit": schema.Int64Attribute{
				Optional:    true,
				Description: "The maximum number of services to return. Defaults to 100.",
			},
			"services": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of services retrieved from CacheFly.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the service.",
						},
						"updated_at": schema.StringAttribute{
							Computed:    true,
							Description: "The timestamp when the service was last updated.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "The timestamp when the service was created.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the service.",
						},
						"unique_name": schema.StringAttribute{
							Computed:    true,
							Description: "The unique name of the service.",
						},
						"auto_ssl": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether auto SSL is enabled for the service.",
						},
						"configuration_mode": schema.StringAttribute{
							Computed:    true,
							Description: "The configuration mode of the service.",
						},
						"status": schema.StringAttribute{
							Computed:    true,
							Description: "The status of the service.",
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *servicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		ResponseType    types.String `tfsdk:"response_type"`
		IncludeFeatures types.Bool   `tfsdk:"include_features"`
		Status          types.String `tfsdk:"status"`
		Offset          types.Int64  `tfsdk:"offset"`
		Limit           types.Int64  `tfsdk:"limit"`
		Services        []struct {
			ID                types.String `tfsdk:"id"`
			UpdatedAt         types.String `tfsdk:"updated_at"`
			CreatedAt         types.String `tfsdk:"created_at"`
			Name              types.String `tfsdk:"name"`
			UniqueName        types.String `tfsdk:"unique_name"`
			AutoSSL           types.Bool   `tfsdk:"auto_ssl"`
			ConfigurationMode types.String `tfsdk:"configuration_mode"`
			Status            types.String `tfsdk:"status"`
		} `tfsdk:"services"`
	}

	// Read the configuration
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default values if not provided
	listOptions := api.ListOptions{
		ResponseType:    "shallow",
		IncludeFeatures: false,
		Offset:          0,
		Limit:           100,
	}
	if !config.ResponseType.IsNull() {
		listOptions.ResponseType = config.ResponseType.ValueString()
	}
	if !config.IncludeFeatures.IsNull() {
		listOptions.IncludeFeatures = config.IncludeFeatures.ValueBool()
	}
	if !config.Status.IsNull() {
		listOptions.Status = config.Status.ValueString()
	}
	if !config.Offset.IsNull() {
		listOptions.Offset = int(config.Offset.ValueInt64())
	}
	if !config.Limit.IsNull() {
		listOptions.Limit = int(config.Limit.ValueInt64())
	}

	// Call the CacheFly API
	response, err := d.client.Services.List(ctx, listOptions)
	if err != nil {
		resp.Diagnostics.AddError("API Error", "Failed to list services: "+err.Error())
		return
	}

	// Map the API response to the Terraform state
	config.Services = make([]struct {
		ID                types.String `tfsdk:"id"`
		UpdatedAt         types.String `tfsdk:"updated_at"`
		CreatedAt         types.String `tfsdk:"created_at"`
		Name              types.String `tfsdk:"name"`
		UniqueName        types.String `tfsdk:"unique_name"`
		AutoSSL           types.Bool   `tfsdk:"auto_ssl"`
		ConfigurationMode types.String `tfsdk:"configuration_mode"`
		Status            types.String `tfsdk:"status"`
	}, len(response.Services))

	for i, service := range response.Services {
		config.Services[i].ID = types.StringValue(service.ID)
		config.Services[i].UpdatedAt = types.StringValue(service.UpdatedAt)
		config.Services[i].CreatedAt = types.StringValue(service.CreatedAt)
		config.Services[i].Name = types.StringValue(service.Name)
		config.Services[i].UniqueName = types.StringValue(service.UniqueName)
		config.Services[i].AutoSSL = types.BoolValue(service.AutoSSL)
		config.Services[i].ConfigurationMode = types.StringValue(service.ConfigurationMode)
		config.Services[i].Status = types.StringValue(service.Status)
	}

	// Set the state
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
