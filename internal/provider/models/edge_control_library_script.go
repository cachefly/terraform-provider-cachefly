package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"
)

// EdgeControlLibraryScriptModel is the Terraform model for the
// cachefly_edge_control_library_script resource and data source, and for the
// items of the cachefly_edge_control_library_scripts data source.
type EdgeControlLibraryScriptModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Kind      types.String `tfsdk:"kind"`
	Code      types.String `tfsdk:"code"`
	Type      types.String `tfsdk:"type"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// NewEdgeControlLibraryScriptModel converts a library script returned by the API.
func NewEdgeControlLibraryScriptModel(script *api.EdgeControlLibraryScript) EdgeControlLibraryScriptModel {
	return EdgeControlLibraryScriptModel{
		ID:        types.StringValue(script.ID),
		Name:      types.StringValue(script.Name),
		Kind:      types.StringValue(script.Kind),
		Code:      types.StringValue(script.Code),
		Type:      types.StringValue(script.Type),
		CreatedAt: types.StringValue(script.CreatedAt),
		UpdatedAt: types.StringValue(script.UpdatedAt),
	}
}

// EdgeControlLibraryScriptsDataSourceModel is the Terraform model for the
// cachefly_edge_control_library_scripts data source.
type EdgeControlLibraryScriptsDataSourceModel struct {
	Type    types.String                    `tfsdk:"type"`
	Kind    types.String                    `tfsdk:"kind"`
	Search  types.String                    `tfsdk:"search"`
	Offset  types.Int32                     `tfsdk:"offset"`
	Limit   types.Int32                     `tfsdk:"limit"`
	Scripts []EdgeControlLibraryScriptModel `tfsdk:"scripts"`
}
