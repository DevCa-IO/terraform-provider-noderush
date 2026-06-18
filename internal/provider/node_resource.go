package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &nodeResource{}
	_ resource.ResourceWithImportState = &nodeResource{}
)

// How long Create waits for a node to finish provisioning before giving up.
const nodeProvisionTimeout = 15 * time.Minute

type nodeResource struct{ client *Client }

func NewNodeResource() resource.Resource { return &nodeResource{} }

type nodeModel struct {
	ID          types.String `tfsdk:"id"`
	Hostname    types.String `tfsdk:"hostname"`
	RegionCode  types.String `tfsdk:"region_code"`
	ImageID     types.String `tfsdk:"image_id"`
	CPU         types.Int64  `tfsdk:"cpu"`
	RAMGB       types.Int64  `tfsdk:"ram_gb"`
	DiskGB      types.Int64  `tfsdk:"disk_gb"`
	BillingMode types.String `tfsdk:"billing_mode"`
	SKUID       types.String `tfsdk:"sku_id"`
	CloudInit   types.String `tfsdk:"cloud_init"`
	SSHKeyIDs   types.List   `tfsdk:"ssh_key_ids"`
	IPv4        types.String `tfsdk:"ipv4"`
	IPv6        types.String `tfsdk:"ipv6"`
	Status      types.String `tfsdk:"status"`
}

func (r *nodeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node"
}

func (r *nodeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	strReplace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	intReplace := []planmodifier.Int64{int64planmodifier.RequiresReplace()}
	keepUnknown := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A NodeRush VPS node. Create blocks until the node finishes provisioning (status ONLINE). Every configuration attribute forces replacement; resize/rescale is not yet supported in-place.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Opaque node id.", PlanModifiers: keepUnknown},
			"hostname":    schema.StringAttribute{Required: true, MarkdownDescription: "Hostname (<= 63 chars, DNS-label safe).", PlanModifiers: strReplace},
			"region_code": schema.StringAttribute{Required: true, MarkdownDescription: "Region to deploy in (e.g. `fra`, `iad`).", PlanModifiers: strReplace},
			"image_id":    schema.StringAttribute{Required: true, MarkdownDescription: "OS image id (see the `noderush_images` data source).", PlanModifiers: strReplace},
			"cpu":         schema.Int64Attribute{Required: true, MarkdownDescription: "vCPU cores.", PlanModifiers: intReplace},
			"ram_gb":      schema.Int64Attribute{Required: true, MarkdownDescription: "Memory in GB.", PlanModifiers: intReplace},
			"disk_gb":     schema.Int64Attribute{Required: true, MarkdownDescription: "Disk in GB.", PlanModifiers: intReplace},
			"billing_mode": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("HOURLY"),
				MarkdownDescription: "`HOURLY` (default) or `MONTHLY`. Charged upfront at deploy.",
				PlanModifiers:       strReplace,
			},
			"sku_id":      schema.StringAttribute{Optional: true, MarkdownDescription: "Optional SKU/plan id to pin pricing (see `noderush_plans`).", PlanModifiers: strReplace},
			"cloud_init":  schema.StringAttribute{Optional: true, MarkdownDescription: "Optional cloud-init script run on first boot.", PlanModifiers: strReplace},
			"ssh_key_ids": schema.ListAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "SSH key ids to inject (Linux images). See `noderush_ssh_key`.", PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()}},
			"ipv4":        schema.StringAttribute{Computed: true, MarkdownDescription: "Allocated IPv4 address.", PlanModifiers: keepUnknown},
			"ipv6":        schema.StringAttribute{Computed: true, MarkdownDescription: "Allocated IPv6 address, if any.", PlanModifiers: keepUnknown},
			"status":      schema.StringAttribute{Computed: true, MarkdownDescription: "Lifecycle status (ONLINE once provisioned)."},
		},
	}
}

func (r *nodeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *nodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan nodeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := NodeCreate{
		Hostname:    plan.Hostname.ValueString(),
		RegionCode:  plan.RegionCode.ValueString(),
		ImageID:     plan.ImageID.ValueString(),
		CPU:         plan.CPU.ValueInt64(),
		RAMGB:       plan.RAMGB.ValueInt64(),
		DiskGB:      plan.DiskGB.ValueInt64(),
		BillingMode: plan.BillingMode.ValueString(),
		SKUID:       plan.SKUID.ValueString(),
		CloudInit:   plan.CloudInit.ValueString(),
	}
	if !plan.SSHKeyIDs.IsNull() && !plan.SSHKeyIDs.IsUnknown() {
		resp.Diagnostics.Append(plan.SSHKeyIDs.ElementsAs(ctx, &body.SSHKeyIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	node, err := r.client.CreateNode(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Could not create node", err.Error())
		return
	}

	// Block until the node reaches ONLINE (or fails / times out).
	node, err = r.waitForOnline(ctx, node.ID)
	if err != nil {
		resp.Diagnostics.AddError("Node did not come online", err.Error())
		return
	}
	r.apply(&plan, node)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *nodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state nodeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	node, err := r.client.GetNode(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Could not read node", err.Error())
		return
	}
	if node == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	r.apply(&state, node)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update never runs meaningfully: every attribute forces replacement. Present to
// satisfy the interface.
func (r *nodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan nodeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *nodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state nodeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteNode(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Could not delete node", err.Error())
	}
}

func (r *nodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// apply copies the API node's computed fields into the model.
func (r *nodeResource) apply(m *nodeModel, node *Node) {
	m.ID = types.StringValue(node.ID)
	m.Hostname = types.StringValue(node.Hostname)
	m.RegionCode = types.StringValue(node.RegionCode)
	m.ImageID = types.StringValue(node.ImageID)
	m.CPU = types.Int64Value(node.CPU)
	m.RAMGB = types.Int64Value(node.RAMGB)
	m.DiskGB = types.Int64Value(node.DiskGB)
	if node.BillingMode != "" {
		m.BillingMode = types.StringValue(node.BillingMode)
	}
	m.Status = types.StringValue(node.Status)
	m.IPv4 = strOrNull(node.IPv4)
	m.IPv6 = strOrNull(node.IPv6)
}

func strOrNull(s *string) types.String {
	if s == nil || *s == "" {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

// waitForOnline polls the node until it is ONLINE, fails (ERROR), is gone, or the
// timeout/context elapses.
func (r *nodeResource) waitForOnline(ctx context.Context, id string) (*Node, error) {
	deadline := time.Now().Add(nodeProvisionTimeout)
	for {
		node, err := r.client.GetNode(ctx, id)
		if err != nil {
			return nil, err
		}
		if node == nil {
			return nil, fmt.Errorf("node %s disappeared during provisioning", id)
		}
		switch node.Status {
		case "ONLINE":
			return node, nil
		case "ERROR":
			reason := "provisioning failed"
			if node.FailureReason != nil && *node.FailureReason != "" {
				reason = *node.FailureReason
			}
			return nil, fmt.Errorf("node %s entered ERROR: %s", id, reason)
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("node %s did not reach ONLINE within %s (last status %s)", id, nodeProvisionTimeout, node.Status)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}
}
