package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly"
	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &EdgeControlScriptResource{}
	_ resource.ResourceWithConfigure   = &EdgeControlScriptResource{}
	_ resource.ResourceWithImportState = &EdgeControlScriptResource{}
)

func NewEdgeControlScriptResource() resource.Resource {
	return &EdgeControlScriptResource{}
}

// EdgeControlScriptResource manages the live edge control script of a service
// for one kind (REQUEST or RESPONSE).
type EdgeControlScriptResource struct {
	client *cachefly.Client
}

func (r *EdgeControlScriptResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_edge_control_script"
}

func (r *EdgeControlScriptResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the edge control script that runs for a service. Every change to `script` publishes a new, " +
			"immutable version and, unless `activated` is `false`, makes it the active version. Destroying the resource " +
			"deactivates the script; published versions are kept because the API cannot delete them. The service draft " +
			"edited in the portal is never changed. Edge Control must be enabled for the account.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier in the form `service_id:kind`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_id": schema.StringAttribute{
				Description: "ID of the service the script runs for.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: "When the script runs: `REQUEST` (when a request is received) or `RESPONSE` (before the response is returned).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(string(api.EdgeControlScriptKindRequest), string(api.EdgeControlScriptKindResponse)),
				},
			},
			"script": schema.StringAttribute{
				MarkdownDescription: "JavaScript source of the script. It must define a `handler` function that receives the event and returns it.",
				Required:            true,
			},
			"activated": schema.BoolAttribute{
				MarkdownDescription: "Whether the script is active. When `false`, no script runs for this service and kind. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"version": schema.Int64Attribute{
				Description: "Version number of the published script.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the published version (ACTIVE or DEACTIVATED).",
				Computed:    true,
			},
			"script_size": schema.Int64Attribute{
				Description: "Size of the script in bytes.",
				Computed:    true,
			},
			"filename": schema.StringAttribute{
				Description: "Name of the file the version is stored as.",
				Computed:    true,
			},
			"last_activated_at": schema.StringAttribute{
				Description: "When the version was last activated.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the version was published.",
				Computed:    true,
			},
		},
	}
}

func (r *EdgeControlScriptResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EdgeControlScriptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.EdgeControlScriptResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()
	kind := api.EdgeControlScriptKind(data.Kind.ValueString())

	script, err := r.publish(ctx, serviceID, kind, data.Script.ValueString(), data.Activated.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating CacheFly Edge Control Script",
			fmt.Sprintf("Could not publish the %s script of service %s: %s", kind, serviceID, err),
		)
		return
	}

	data.ID = types.StringValue(edgeControlScriptID(serviceID, kind))
	mapEdgeControlScriptToState(script, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EdgeControlScriptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.EdgeControlScriptResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()
	kind := api.EdgeControlScriptKind(data.Kind.ValueString())

	script, err := r.currentScript(ctx, serviceID, kind, data.Version)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading CacheFly Edge Control Script",
			fmt.Sprintf("Could not read the %s script of service %s: %s", kind, serviceID, err),
		)
		return
	}
	if script == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(edgeControlScriptID(serviceID, kind))
	mapEdgeControlScriptToState(script, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EdgeControlScriptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state models.EdgeControlScriptResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := plan.ServiceID.ValueString()
	kind := api.EdgeControlScriptKind(plan.Kind.ValueString())
	activated := plan.Activated.ValueBool()

	var script *api.EdgeControlScript
	var err error
	if !plan.Script.Equal(state.Script) {
		script, err = r.publish(ctx, serviceID, kind, plan.Script.ValueString(), activated)
	} else {
		script, err = r.setActivation(ctx, serviceID, kind, int(state.Version.ValueInt64()), activated)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating CacheFly Edge Control Script",
			fmt.Sprintf("Could not update the %s script of service %s: %s", kind, serviceID, err),
		)
		return
	}

	plan.ID = types.StringValue(edgeControlScriptID(serviceID, kind))
	mapEdgeControlScriptToState(script, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EdgeControlScriptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.EdgeControlScriptResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := data.ServiceID.ValueString()
	kind := api.EdgeControlScriptKind(data.Kind.ValueString())

	if err := r.client.EdgeControlScripts.Deactivate(ctx, serviceID, kind); err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError(
			"Error Deactivating CacheFly Edge Control Script",
			fmt.Sprintf("Could not deactivate the %s script of service %s: %s", kind, serviceID, err),
		)
	}
}

// ImportState imports the script of a service by "service_id:kind".
func (r *EdgeControlScriptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	serviceID, rawKind, found := strings.Cut(req.ID, ":")
	kind := api.EdgeControlScriptKind(strings.ToUpper(rawKind))
	if !found || serviceID == "" || (kind != api.EdgeControlScriptKindRequest && kind != api.EdgeControlScriptKindResponse) {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in the format 'service_id:kind', where kind is REQUEST or RESPONSE.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), edgeControlScriptID(serviceID, kind))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_id"), serviceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("kind"), string(kind))...)
}

// publish creates a new version with the given script, then activates it or
// leaves no version of the kind active.
func (r *EdgeControlScriptResource) publish(ctx context.Context, serviceID string, kind api.EdgeControlScriptKind, script string, activated bool) (*api.EdgeControlScript, error) {
	created, err := r.client.EdgeControlScripts.CreateVersion(ctx, serviceID, kind, script)
	if err != nil {
		return nil, fmt.Errorf("could not publish a new version: %w", err)
	}

	return r.setActivation(ctx, serviceID, kind, created.Version, activated)
}

// setActivation makes version the active version of the kind, or deactivates
// every version of the kind, and returns the version as it is afterwards.
func (r *EdgeControlScriptResource) setActivation(ctx context.Context, serviceID string, kind api.EdgeControlScriptKind, version int, activated bool) (*api.EdgeControlScript, error) {
	if activated {
		active, err := r.client.EdgeControlScripts.ActivateVersion(ctx, serviceID, kind, version)
		if err != nil {
			return nil, fmt.Errorf("could not activate version %d: %w", version, err)
		}
		return active, nil
	}

	if err := r.client.EdgeControlScripts.Deactivate(ctx, serviceID, kind); err != nil {
		return nil, fmt.Errorf("could not deactivate the script: %w", err)
	}

	current, err := r.client.EdgeControlScripts.GetVersion(ctx, serviceID, kind, version)
	if err != nil {
		return nil, fmt.Errorf("could not read version %d: %w", version, err)
	}
	return current, nil
}

// currentScript returns the active version. When no version is active it
// returns the managed version, or the newest version when the managed one is
// unknown (after import). It returns nil when no version exists or the service
// cannot be found.
func (r *EdgeControlScriptResource) currentScript(ctx context.Context, serviceID string, kind api.EdgeControlScriptKind, managedVersion types.Int64) (*api.EdgeControlScript, error) {
	active, err := r.client.EdgeControlScripts.GetActiveVersion(ctx, serviceID, kind)
	if err == nil {
		return active, nil
	}
	if !isNotFoundError(err) {
		return nil, err
	}

	if !managedVersion.IsNull() && !managedVersion.IsUnknown() && managedVersion.ValueInt64() > 0 {
		version, err := r.client.EdgeControlScripts.GetVersion(ctx, serviceID, kind, int(managedVersion.ValueInt64()))
		if err == nil {
			return version, nil
		}
		if !isNotFoundError(err) {
			return nil, err
		}
	}

	versions, err := r.client.EdgeControlScripts.ListVersions(ctx, serviceID, kind)
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(versions) == 0 {
		return nil, nil
	}
	return &versions[0], nil
}

func edgeControlScriptID(serviceID string, kind api.EdgeControlScriptKind) string {
	return serviceID + ":" + string(kind)
}

func mapEdgeControlScriptToState(script *api.EdgeControlScript, data *models.EdgeControlScriptResourceModel) {
	data.Script = types.StringValue(script.Script)
	data.Activated = types.BoolValue(script.Status == api.EdgeControlScriptStatusActive)
	data.Version = types.Int64Value(int64(script.Version))
	data.Status = types.StringValue(script.Status)
	data.ScriptSize = types.Int64Value(int64(script.ScriptSize))
	data.Filename = types.StringValue(script.Filename)
	data.LastActivatedAt = types.StringValue(script.LastActivatedAt)
	data.CreatedAt = types.StringValue(script.CreatedAt)
}
