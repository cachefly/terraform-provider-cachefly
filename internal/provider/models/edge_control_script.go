package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"
)

// EdgeControlScriptResourceModel is the Terraform model for the
// cachefly_edge_control_script resource.
type EdgeControlScriptResourceModel struct {
	ID              types.String `tfsdk:"id"`
	ServiceID       types.String `tfsdk:"service_id"`
	Kind            types.String `tfsdk:"kind"`
	Script          types.String `tfsdk:"script"`
	Activated       types.Bool   `tfsdk:"activated"`
	Version         types.Int64  `tfsdk:"version"`
	Status          types.String `tfsdk:"status"`
	ScriptSize      types.Int64  `tfsdk:"script_size"`
	Filename        types.String `tfsdk:"filename"`
	LastActivatedAt types.String `tfsdk:"last_activated_at"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

// EdgeControlScriptDataSourceModel is the Terraform model for the
// cachefly_edge_control_script data source.
type EdgeControlScriptDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	ServiceID       types.String `tfsdk:"service_id"`
	Kind            types.String `tfsdk:"kind"`
	Version         types.Int64  `tfsdk:"version"`
	Script          types.String `tfsdk:"script"`
	Status          types.String `tfsdk:"status"`
	ScriptSize      types.Int64  `tfsdk:"script_size"`
	Filename        types.String `tfsdk:"filename"`
	LastActivatedAt types.String `tfsdk:"last_activated_at"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

// EdgeControlScriptVersionsDataSourceModel is the Terraform model for the
// cachefly_edge_control_script_versions data source.
type EdgeControlScriptVersionsDataSourceModel struct {
	ID        types.String                    `tfsdk:"id"`
	ServiceID types.String                    `tfsdk:"service_id"`
	Kind      types.String                    `tfsdk:"kind"`
	Versions  []EdgeControlScriptVersionModel `tfsdk:"versions"`
}

// EdgeControlScriptVersionModel is one published edge control script version.
type EdgeControlScriptVersionModel struct {
	ID              types.String `tfsdk:"id"`
	Version         types.Int64  `tfsdk:"version"`
	Status          types.String `tfsdk:"status"`
	Script          types.String `tfsdk:"script"`
	ScriptSize      types.Int64  `tfsdk:"script_size"`
	Filename        types.String `tfsdk:"filename"`
	LastActivatedAt types.String `tfsdk:"last_activated_at"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

// NewEdgeControlScriptVersionModel converts a script version returned by the API.
func NewEdgeControlScriptVersionModel(script *api.EdgeControlScript) EdgeControlScriptVersionModel {
	return EdgeControlScriptVersionModel{
		ID:              types.StringValue(script.ID),
		Version:         types.Int64Value(int64(script.Version)),
		Status:          types.StringValue(script.Status),
		Script:          types.StringValue(script.Script),
		ScriptSize:      types.Int64Value(int64(script.ScriptSize)),
		Filename:        types.StringValue(script.Filename),
		LastActivatedAt: types.StringValue(script.LastActivatedAt),
		CreatedAt:       types.StringValue(script.CreatedAt),
	}
}
