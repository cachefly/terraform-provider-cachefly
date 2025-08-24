package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly"
	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &UserResource{}
	_ resource.ResourceWithImportState = &UserResource{}
)

// NewUserResource is a helper function to simplify the provider implementation
func NewUserResource() resource.Resource {
	return &UserResource{}
}

// UserResource defines the resource implementation
type UserResource struct {
	client *cachefly.Client
}

// Metadata returns the resource type name
func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the resource
func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CacheFly User resource. Manages user accounts in the CacheFly platform.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the user.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Description: "Username for the user account.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				Description: "Email address of the user.",
				Required:    true,
			},
			"full_name": schema.StringAttribute{
				Description: "Full name of the user.",
				Optional:    true,
			},
			"phone": schema.StringAttribute{
				Description: "Phone number of the user.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for the user account.",
				Required:    true,
				Sensitive:   true,
			},
			"password_change_required": schema.BoolAttribute{
				Description: "Whether the user must change password on next login.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"services": schema.SetAttribute{
				Description: "Set of service IDs the user has access to.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"permissions": schema.SetAttribute{
				Description: "Set of permissions granted to the user.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			// Computed attributes
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
	}
}

// Configure adds the provider configured client to the resource
func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state
func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.UserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build create request
	createReq := api.CreateUserRequest{
		Username:               data.Username.ValueString(),
		Password:               data.Password.ValueString(),
		PasswordChangeRequired: data.PasswordChangeRequired.ValueBoolPointer(),
		Email:                  data.Email.ValueString(),
		FullName:               data.FullName.ValueString(),
	}

	// Optional fields
	if !data.Phone.IsNull() && !data.Phone.IsUnknown() {
		createReq.Phone = data.Phone.ValueString()
	}

	// Convert Services set to string slice
	if !data.Services.IsNull() && !data.Services.IsUnknown() {
		var services []string
		serviceElements := make([]types.String, 0, len(data.Services.Elements()))
		data.Services.ElementsAs(ctx, &serviceElements, false)
		for _, elem := range serviceElements {
			services = append(services, elem.ValueString())
		}
		createReq.Services = services
	}

	// Convert Permissions set to string slice
	if !data.Permissions.IsNull() && !data.Permissions.IsUnknown() {
		var permissions []string
		permissionElements := make([]types.String, 0, len(data.Permissions.Elements()))
		data.Permissions.ElementsAs(ctx, &permissionElements, false)
		for _, elem := range permissionElements {
			permissions = append(permissions, elem.ValueString())
		}
		createReq.Permissions = permissions
	}

	user, err := r.client.Users.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating CacheFly User",
			"Could not create user, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to state
	r.mapUserToState(user, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data
func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.UserModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.ID.ValueString()

	user, err := r.client.Users.GetByID(ctx, userID, "")
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Error Reading CacheFly User",
				"Could not read user with ID "+userID+": "+err.Error(),
			)
		}
		return
	}

	// Map response to state
	r.mapUserToState(user, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success
func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.UserModel
	var state models.UserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.ID.ValueString()

	updateReq := api.UpdateUserRequest{}

	if !data.Password.Equal(state.Password) {
		updateReq.Password = data.Password.ValueString()
	}
	if !data.Email.Equal(state.Email) {
		updateReq.Email = data.Email.ValueString()
	}
	if !data.FullName.Equal(state.FullName) {
		updateReq.FullName = data.FullName.ValueString()
	}
	if !data.Phone.Equal(state.Phone) {
		updateReq.Phone = data.Phone.ValueString()
	}

	updateReq.PasswordChangeRequired = data.PasswordChangeRequired.ValueBoolPointer()

	if !data.Services.Equal(state.Services) {
		var services []string
		serviceElements := make([]types.String, 0, len(data.Services.Elements()))
		data.Services.ElementsAs(ctx, &serviceElements, false)
		for _, elem := range serviceElements {
			services = append(services, elem.ValueString())
		}
		updateReq.Services = services
	}

	if !data.Permissions.Equal(state.Permissions) {
		var permissions []string
		permissionElements := make([]types.String, 0, len(data.Permissions.Elements()))
		data.Permissions.ElementsAs(ctx, &permissionElements, false)
		for _, elem := range permissionElements {
			permissions = append(permissions, elem.ValueString())
		}
		updateReq.Permissions = permissions
	}

	user, err := r.client.Users.UpdateByID(ctx, userID, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating CacheFly User",
			"Could not update user with ID "+userID+": "+err.Error(),
		)
		return
	}

	r.mapUserToState(user, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success
func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.UserModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.ID.ValueString()

	err := r.client.Users.DeleteByID(ctx, userID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting CacheFly User",
			"Could not delete user with ID "+userID+": "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform state
func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to map SDK User to Terraform state
func (r *UserResource) mapUserToState(user *api.User, data *models.UserModel) {
	data.ID = types.StringValue(user.ID)
	data.Username = types.StringValue(user.Username)
	data.Email = types.StringValue(user.Email)
	data.FullName = types.StringValue(user.FullName)
	data.Phone = types.StringPointerValue(user.Phone)
	data.Status = types.StringValue(user.Status)
	data.CreatedAt = types.StringValue(user.CreatedAt)
	data.UpdatedAt = types.StringValue(user.UpdatedAt)
	data.PasswordChangeRequired = types.BoolValue(user.PasswordChangeRequired)

	// Convert Services slice to set
	if len(user.Services) > 0 {
		serviceValues := make([]types.String, len(user.Services))
		for i, service := range user.Services {
			serviceValues[i] = types.StringValue(service)
		}
		data.Services = types.SetValueMust(types.StringType, convertStringSliceToValues(serviceValues))
	} else {
		data.Services = types.SetValueMust(types.StringType, []attr.Value{})
	}

	// Convert Permissions slice to set
	if len(user.Permissions) > 0 {
		permissionValues := make([]types.String, len(user.Permissions))
		for i, permission := range user.Permissions {
			permissionValues[i] = types.StringValue(permission)
		}
		data.Permissions = types.SetValueMust(types.StringType, convertStringSliceToValues(permissionValues))
	} else {
		data.Permissions = types.SetValueMust(types.StringType, []attr.Value{})
	}
}

// Helper function to convert []types.String to []types.Value
func convertStringSliceToValues(stringSlice []types.String) []attr.Value {
	values := make([]attr.Value, len(stringSlice))
	for i, str := range stringSlice {
		values[i] = str
	}
	return values
}
