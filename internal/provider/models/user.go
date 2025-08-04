package models

import (
	"context"

	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UserModel represents the Terraform model for a CacheFly user
type UserModel struct {
	ID                     types.String `tfsdk:"id"`
	Username               types.String `tfsdk:"username"`
	Email                  types.String `tfsdk:"email"`
	FullName               types.String `tfsdk:"full_name"`
	Phone                  types.String `tfsdk:"phone"`
	Password               types.String `tfsdk:"password"`
	PasswordChangeRequired types.Bool   `tfsdk:"password_change_required"`
	Services               types.Set    `tfsdk:"services"`    // Set of service IDs
	Permissions            types.Set    `tfsdk:"permissions"` // Set of permission strings
	Status                 types.String `tfsdk:"status"`      // Computed field
	CreatedAt              types.String `tfsdk:"created_at"`  // Computed field
	UpdatedAt              types.String `tfsdk:"updated_at"`  // Computed field
}

// UsersDataSourceModel represents the Terraform model for the users data source
type UsersDataSourceModel struct {
	ID           types.String    `tfsdk:"id"`
	Search       types.String    `tfsdk:"search"`
	Offset       types.Int64     `tfsdk:"offset"`
	Limit        types.Int64     `tfsdk:"limit"`
	ResponseType types.String    `tfsdk:"response_type"`
	Users        []UserDataModel `tfsdk:"users"`
	TotalCount   types.Int64     `tfsdk:"total_count"`
}

// UserDataModel represents a user in the data source (simplified version of UserModel)
type UserDataModel struct {
	ID                     types.String `tfsdk:"id"`
	Username               types.String `tfsdk:"username"`
	Email                  types.String `tfsdk:"email"`
	FullName               types.String `tfsdk:"full_name"`
	Phone                  types.String `tfsdk:"phone"`
	PasswordChangeRequired types.Bool   `tfsdk:"password_change_required"`
	Services               types.Set    `tfsdk:"services"`
	Permissions            types.Set    `tfsdk:"permissions"`
	Status                 types.String `tfsdk:"status"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
}

// ToSDKCreateRequest converts the Terraform model to SDK CreateUserRequest
func (m *UserModel) ToSDKCreateRequest(ctx context.Context) *api.CreateUserRequest {
	// Convert Services set to string slice
	var services []string
	if !m.Services.IsNull() && !m.Services.IsUnknown() {
		serviceElements := make([]types.String, 0, len(m.Services.Elements()))
		m.Services.ElementsAs(ctx, &serviceElements, false)
		for _, elem := range serviceElements {
			services = append(services, elem.ValueString())
		}
	}

	// Convert Permissions set to string slice
	var permissions []string
	if !m.Permissions.IsNull() && !m.Permissions.IsUnknown() {
		permissionElements := make([]types.String, 0, len(m.Permissions.Elements()))
		m.Permissions.ElementsAs(ctx, &permissionElements, false)
		for _, elem := range permissionElements {
			permissions = append(permissions, elem.ValueString())
		}
	}

	return &api.CreateUserRequest{
		Username:               m.Username.ValueString(),
		Password:               m.Password.ValueString(),
		Email:                  m.Email.ValueString(),
		FullName:               m.FullName.ValueString(),
		Phone:                  m.Phone.ValueString(),
		PasswordChangeRequired: m.PasswordChangeRequired.ValueBoolPointer(),
		Services:               services,
		Permissions:            permissions,
	}
}

// ToSDKUpdateRequest converts the Terraform model to SDK UpdateUserRequest
func (m *UserModel) ToSDKUpdateRequest(ctx context.Context) *api.UpdateUserRequest {
	req := &api.UpdateUserRequest{}

	// Only set fields that have values (not null/unknown)
	if !m.Password.IsNull() && !m.Password.IsUnknown() && m.Password.ValueString() != "" {
		req.Password = m.Password.ValueString()
	}

	if !m.Email.IsNull() && !m.Email.IsUnknown() {
		req.Email = m.Email.ValueString()
	}

	if !m.FullName.IsNull() && !m.FullName.IsUnknown() {
		req.FullName = m.FullName.ValueString()
	}

	if !m.Phone.IsNull() && !m.Phone.IsUnknown() {
		req.Phone = m.Phone.ValueString()
	}

	if !m.PasswordChangeRequired.IsNull() && !m.PasswordChangeRequired.IsUnknown() {
		val := m.PasswordChangeRequired.ValueBool()
		req.PasswordChangeRequired = &val
	}

	// Convert Services set to string slice
	if !m.Services.IsNull() && !m.Services.IsUnknown() {
		var services []string
		serviceElements := make([]types.String, 0, len(m.Services.Elements()))
		m.Services.ElementsAs(ctx, &serviceElements, false)
		for _, elem := range serviceElements {
			services = append(services, elem.ValueString())
		}
		req.Services = services
	}

	// Convert Permissions set to string slice
	if !m.Permissions.IsNull() && !m.Permissions.IsUnknown() {
		var permissions []string
		permissionElements := make([]types.String, 0, len(m.Permissions.Elements()))
		m.Permissions.ElementsAs(ctx, &permissionElements, false)
		for _, elem := range permissionElements {
			permissions = append(permissions, elem.ValueString())
		}
		req.Permissions = permissions
	}

	return req
}

// FromSDKUser converts an SDK User to the Terraform model
func (m *UserModel) FromSDKUser(ctx context.Context, user *api.User) {
	m.ID = types.StringValue(user.ID)
	m.Username = types.StringValue(user.Username)
	m.Email = types.StringValue(user.Email)
	m.FullName = types.StringValue(user.FullName)
	m.Phone = types.StringPointerValue(user.Phone)
	m.Status = types.StringValue(user.Status)
	m.CreatedAt = types.StringValue(user.CreatedAt)
	m.UpdatedAt = types.StringValue(user.UpdatedAt)
	m.PasswordChangeRequired = types.BoolValue(user.PasswordChangeRequired)

	// Convert Services slice to set
	if len(user.Services) > 0 {
		serviceValues := make([]attr.Value, len(user.Services))
		for i, service := range user.Services {
			serviceValues[i] = types.StringValue(service)
		}
		m.Services = types.SetValueMust(types.StringType, serviceValues)
	} else {
		m.Services = types.SetValueMust(types.StringType, []attr.Value{})
	}

	// Convert Permissions slice to set
	if len(user.Permissions) > 0 {
		permissionValues := make([]attr.Value, len(user.Permissions))
		for i, permission := range user.Permissions {
			permissionValues[i] = types.StringValue(permission)
		}
		m.Permissions = types.SetValueMust(types.StringType, permissionValues)
	} else {
		m.Permissions = types.SetValueMust(types.StringType, []attr.Value{})
	}
}

// Helper function to convert []types.String to []attr.Value
func convertStringSliceToValues(stringSlice []types.String) []attr.Value {
	values := make([]attr.Value, len(stringSlice))
	for i, str := range stringSlice {
		values[i] = str
	}
	return values
}
