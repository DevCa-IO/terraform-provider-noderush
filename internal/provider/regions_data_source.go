package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &regionsDataSource{}

type regionsDataSource struct{ client *Client }

func NewRegionsDataSource() datasource.DataSource { return &regionsDataSource{} }

type regionModel struct {
	Code        types.String `tfsdk:"code"`
	Label       types.String `tfsdk:"label"`
	CountryCode types.String `tfsdk:"country_code"`
	Status      types.String `tfsdk:"status"`
}

type regionsModel struct {
	Regions []regionModel `tfsdk:"regions"`
}

func (d *regionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_regions"
}

func (d *regionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The set of NodeRush regions you can deploy into.",
		Attributes: map[string]schema.Attribute{
			"regions": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"code":         schema.StringAttribute{Computed: true, MarkdownDescription: "Region code, e.g. `fra`."},
						"label":        schema.StringAttribute{Computed: true, MarkdownDescription: "Human-readable label."},
						"country_code": schema.StringAttribute{Computed: true, MarkdownDescription: "ISO country code."},
						"status":       schema.StringAttribute{Computed: true, MarkdownDescription: "Operational status."},
					},
				},
			},
		},
	}
}

func (d *regionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *regionsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	regions, err := d.client.ListRegions(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Could not list regions", err.Error())
		return
	}
	state := regionsModel{Regions: make([]regionModel, 0, len(regions))}
	for _, r := range regions {
		state.Regions = append(state.Regions, regionModel{
			Code:        types.StringValue(r.Code),
			Label:       types.StringValue(r.Label),
			CountryCode: types.StringValue(r.CountryCode),
			Status:      types.StringValue(r.Status),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
