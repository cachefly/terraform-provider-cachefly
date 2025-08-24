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

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ datasource.DataSource              = &UsersDataSource{}
	_ datasource.DataSourceWithConfigure = &UsersDataSource{}
)

// NewUsersDataSource is a helper function to simplify the provider implementation
func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

// UsersDataSource defines the data source implementation
type UsersDataSource struct {
	client *cachefly.Client
}

// Metadata returns the data source type name
func (d *UsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

// Schema defines the schema for the data source
func (d *UsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of CacheFly users with optional filtering.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder identifier for the data source.",
				Computed:    true,
			},
			"search": schema.StringAttribute{
				Description: "Search term to filter users by username, email, or full name.",
				Optional:    true,
			},
			"offset": schema.Int64Attribute{
				Description: "Number of users to skip for pagination.",
				Optional:    true,
			},
			"limit": schema.Int64Attribute{
				Description: "Maximum number of users to return.",
				Optional:    true,
			},
			"response_type": schema.StringAttribute{
				Description: "Response type for the API request.",
				Optional:    true,
			},
			"users": schema.ListNestedAttribute{
				Description: "List of users matching the criteria.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the user.",
							Computed:    true,
						},
						"username": schema.StringAttribute{
							Description: "Username of the user.",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "Email address of the user.",
							Computed:    true,
						},
						"full_name": schema.StringAttribute{
							Description: "Full name of the user.",
							Computed:    true,
						},
						"phone": schema.StringAttribute{
							Description: "Phone number of the user.",
							Computed:    true,
						},
						"password_change_required": schema.BoolAttribute{
							Description: "Whether the user must change password on next login.",
							Computed:    true,
						},
						"services": schema.SetAttribute{
							Description: "Set of service IDs the user has access to.",
							ElementType: types.StringType,
							Computed:    true,
						},
						"permissions": schema.SetAttribute{
							Description: "Set of permissions granted to the user.",
							ElementType: types.StringType,
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the user account.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the user was created.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "Timestamp when the user was last updated.",
							Computed:    true,
						},
					},
				},
			},
			"total_count": schema.Int64Attribute{
				Description: "Total number of users available (before pagination).",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *UsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the Terraform state with the latest data
func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.UsersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := api.ListUsersOptions{}

	if !data.Search.IsNull() && !data.Search.IsUnknown() {
		opts.Search = data.Search.ValueString()
	}
	if !data.Offset.IsNull() && !data.Offset.IsUnknown() {
		opts.Offset = int(data.Offset.ValueInt64())
	}
	if !data.ResponseType.IsNull() && !data.ResponseType.IsUnknown() {
		opts.ResponseType = data.ResponseType.ValueString()
	}

	if !data.Limit.IsNull() && !data.Limit.IsUnknown() {
		opts.Limit = int(data.Limit.ValueInt64())
	} else {
		opts.Limit = 100
	}

	var allUsers []api.User
	totalCount := 0
	for {
		pageResp, err := d.client.Users.List(ctx, opts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading CacheFly Users",
				"Could not read users list: "+err.Error(),
			)
			return
		}

		if totalCount == 0 {
			totalCount = pageResp.Meta.Count
		}

		allUsers = append(allUsers, pageResp.Users...)

		fetched := len(pageResp.Users)
		opts.Offset += fetched

		if fetched < opts.Limit || pageResp.Meta.Count > 0 && opts.Offset == pageResp.Meta.Count {
			break
		}
	}

	usersResp := &api.ListUsersResponse{Meta: api.MetaInfo{Count: totalCount}, Users: allUsers}
	d.mapUsersToState(usersResp, &data)

	// Set computed ID
	data.ID = types.StringValue("users")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper function to map SDK users response to Terraform state
func (d *UsersDataSource) mapUsersToState(usersResp *api.ListUsersResponse, data *models.UsersDataSourceModel) {
	// Set total count
	data.TotalCount = types.Int64Value(int64(usersResp.Meta.Count))

	// Map users
	userModels := make([]models.UserDataModel, len(usersResp.Users))
	for i, user := range usersResp.Users {
		userModels[i] = models.UserDataModel{
			ID:                     types.StringValue(user.ID),
			Username:               types.StringValue(user.Username),
			Email:                  types.StringValue(user.Email),
			FullName:               types.StringValue(user.FullName),
			Phone:                  types.StringPointerValue(user.Phone),
			PasswordChangeRequired: types.BoolValue(user.PasswordChangeRequired),
			Status:                 types.StringValue(user.Status),
			CreatedAt:              types.StringValue(user.CreatedAt),
			UpdatedAt:              types.StringValue(user.UpdatedAt),
		}

		// Convert Services slice to set
		if len(user.Services) > 0 {
			serviceValues := make([]types.String, len(user.Services))
			for j, service := range user.Services {
				serviceValues[j] = types.StringValue(service)
			}
			userModels[i].Services = types.SetValueMust(types.StringType, convertStringSliceToValues(serviceValues))
		} else {
			userModels[i].Services = types.SetValueMust(types.StringType, []attr.Value{})
		}

		// Convert Permissions slice to set
		if len(user.Permissions) > 0 {
			permissionValues := make([]types.String, len(user.Permissions))
			for j, permission := range user.Permissions {
				permissionValues[j] = types.StringValue(permission)
			}
			userModels[i].Permissions = types.SetValueMust(types.StringType, convertStringSliceToValues(permissionValues))
		} else {
			userModels[i].Permissions = types.SetValueMust(types.StringType, []attr.Value{})
		}
	}

	data.Users = userModels
}

// Helper function to convert []types.String to []types.Value
func convertStringSliceToValues(stringSlice []types.String) []attr.Value {
	values := make([]attr.Value, len(stringSlice))
	for i, str := range stringSlice {
		values[i] = str
	}
	return values
}
