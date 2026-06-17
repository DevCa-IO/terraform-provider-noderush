package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &volumeResource{}
	_ resource.ResourceWithImportState = &volumeResource{}
)

type volumeResource struct{ client *Client }

func NewVolumeResource() resource.Resource { return &volumeResource{} }

type volumeModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	RegionCode types.String `tfsdk:"region_code"`
	SizeGB     types.Int64  `tfsdk:"size_gb"`
	Status     types.String `tfsdk:"status"`
}

func (r *volumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (r *volumeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A NodeRush block storage volume. Billed per GB-month. Resize is grow-only.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Opaque volume id.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":        schema.StringAttribute{Required: true, MarkdownDescription: "Display name (changing it forces a new volume).", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"region_code": schema.StringAttribute{Required: true, MarkdownDescription: "Region the volume lives in (e.g. `fra`, `iad`). Immutable.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"size_gb":     schema.Int64Attribute{Required: true, MarkdownDescription: "Size in GB. Can be increased in place; decreasing forces a new volume."},
			"status":      schema.StringAttribute{Computed: true, MarkdownDescription: "Lifecycle status.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func (r *volumeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("expected *Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *volumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan volumeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vol, err := r.client.CreateVolume(ctx, plan.Name.ValueString(), plan.RegionCode.ValueString(), plan.SizeGB.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Could not create volume", err.Error())
		return
	}
	plan.ID = types.StringValue(vol.ID)
	plan.Status = types.StringValue(vol.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *volumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state volumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vol, err := r.client.GetVolume(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Could not read volume", err.Error())
		return
	}
	if vol == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.Name = types.StringValue(vol.Name)
	state.RegionCode = types.StringValue(vol.RegionCode)
	state.SizeGB = types.Int64Value(vol.SizeGB)
	state.Status = types.StringValue(vol.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *volumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state volumeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only size_gb can change in place (grow-only enforced by the API).
	if plan.SizeGB.ValueInt64() != state.SizeGB.ValueInt64() {
		if err := r.client.ResizeVolume(ctx, state.ID.ValueString(), plan.SizeGB.ValueInt64()); err != nil {
			resp.Diagnostics.AddError("Could not resize volume", err.Error())
			return
		}
	}
	plan.ID = state.ID
	plan.Status = state.Status
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *volumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state volumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteVolume(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Could not delete volume", err.Error())
	}
}

func (r *volumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
