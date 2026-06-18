package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &plansDataSource{}

type plansDataSource struct{ client *Client }

func NewPlansDataSource() datasource.DataSource { return &plansDataSource{} }

type planModel struct {
	ID           types.String `tfsdk:"id"`
	Family       types.String `tfsdk:"family"`
	Label        types.String `tfsdk:"label"`
	CPU          types.Int64  `tfsdk:"cpu"`
	RAMGB        types.Int64  `tfsdk:"ram_gb"`
	DiskGB       types.Int64  `tfsdk:"disk_gb"`
	HourlyCents  types.Int64  `tfsdk:"hourly_cents"`
	MonthlyCents types.Int64  `tfsdk:"monthly_cents"`
}

type plansModel struct {
	RegionCode types.String `tfsdk:"region_code"`
	Plans      []planModel  `tfsdk:"plans"`
}

func (d *plansDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plans"
}

func (d *plansDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The compute plans (SKUs) available, optionally filtered to a region.",
		Attributes: map[string]schema.Attribute{
			"region_code": schema.StringAttribute{Optional: true, MarkdownDescription: "Filter to plans available in this region."},
			"plans": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Plan/SKU id."},
						"family":        schema.StringAttribute{Computed: true, MarkdownDescription: "Plan family, e.g. `STANDARD`."},
						"label":         schema.StringAttribute{Computed: true, MarkdownDescription: "Human-readable label."},
						"cpu":           schema.Int64Attribute{Computed: true, MarkdownDescription: "vCPU cores."},
						"ram_gb":        schema.Int64Attribute{Computed: true, MarkdownDescription: "Memory in GB."},
						"disk_gb":       schema.Int64Attribute{Computed: true, MarkdownDescription: "Disk in GB."},
						"hourly_cents":  schema.Int64Attribute{Computed: true, MarkdownDescription: "Hourly price in cents."},
						"monthly_cents": schema.Int64Attribute{Computed: true, MarkdownDescription: "Monthly price in cents."},
					},
				},
			},
		},
	}
}

func (d *plansDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("expected *Client, got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *plansDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg plansModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plans, err := d.client.ListPlans(ctx, cfg.RegionCode.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Could not list plans", err.Error())
		return
	}
	cfg.Plans = make([]planModel, 0, len(plans))
	for _, p := range plans {
		cfg.Plans = append(cfg.Plans, planModel{
			ID:           types.StringValue(p.ID),
			Family:       types.StringValue(p.Family),
			Label:        types.StringValue(p.Label),
			CPU:          types.Int64Value(p.CPU),
			RAMGB:        types.Int64Value(p.RAMGB),
			DiskGB:       types.Int64Value(p.DiskGB),
			HourlyCents:  types.Int64Value(p.HourlyCents),
			MonthlyCents: types.Int64Value(p.MonthlyCents),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}
