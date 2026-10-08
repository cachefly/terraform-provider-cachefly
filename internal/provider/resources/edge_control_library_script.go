package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly"
	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &EdgeControlLibraryScriptResource{}
	_ resource.ResourceWithConfigure   = &EdgeControlLibraryScriptResource{}
	_ resource.ResourceWithImportState = &EdgeControlLibraryScriptResource{}
)

func NewEdgeControlLibraryScriptResource() resource.Resource {
	return &EdgeControlLibraryScriptResource{}
}

// EdgeControlLibraryScriptResource manages a USER script in the edge control
// script library.
type EdgeControlLibraryScriptResource struct {
	client *cachefly.Client
}

func (r *EdgeControlLibraryScriptResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_edge_control_library_script"
}

func (r *EdgeControlLibraryScriptResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a reusable script in the account's edge control script library. Library scripts do not run " +
			"by themselves; use their code in `cachefly_edge_control_script` or start a service draft from them in the portal. " +
			"Edge Control must be enabled for the account.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the library script.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the library script.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: "Kind of script: `REQUEST` or `RESPONSE`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(string(api.EdgeControlScriptKindRequest), string(api.EdgeControlScriptKindResponse)),
				},
			},
			"code": schema.StringAttribute{
				MarkdownDescription: "JavaScript source of the script. It must define a `handler` function.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Ownership type of the script. Always `USER` for scripts managed by Terraform.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "When the library script was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "When the library script was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *EdgeControlLibraryScriptResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EdgeControlLibraryScriptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.EdgeControlLibraryScriptModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	script, err := r.client.EdgeControlLibrary.Create(ctx, edgeControlLibraryScriptRequest(&data))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating CacheFly Edge Control Library Script",
			"Could not create library script, unexpected error: "+err.Error(),
		)
		return
	}

	data = models.NewEdgeControlLibraryScriptModel(script)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EdgeControlLibraryScriptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.EdgeControlLibraryScriptModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	script, err := r.client.EdgeControlLibrary.GetByID(ctx, data.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading CacheFly Edge Control Library Script",
			"Could not read library script ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	data = models.NewEdgeControlLibraryScriptModel(script)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EdgeControlLibraryScriptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.EdgeControlLibraryScriptModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	script, err := r.client.EdgeControlLibrary.Update(ctx, data.ID.ValueString(), edgeControlLibraryScriptRequest(&data))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating CacheFly Edge Control Library Script",
			"Could not update library script ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	data = models.NewEdgeControlLibraryScriptModel(script)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EdgeControlLibraryScriptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.EdgeControlLibraryScriptModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.EdgeControlLibrary.Delete(ctx, data.ID.ValueString()); err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError(
			"Error Deleting CacheFly Edge Control Library Script",
			"Could not delete library script ID "+data.ID.ValueString()+": "+err.Error(),
		)
	}
}

func (r *EdgeControlLibraryScriptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func edgeControlLibraryScriptRequest(data *models.EdgeControlLibraryScriptModel) api.EdgeControlLibraryScriptRequest {
	return api.EdgeControlLibraryScriptRequest{
		Name: data.Name.ValueString(),
		Kind: api.EdgeControlScriptKind(data.Kind.ValueString()),
		Code: data.Code.ValueString(),
	}
}
