package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &imagesDataSource{}

type imagesDataSource struct{ client *Client }

func NewImagesDataSource() datasource.DataSource { return &imagesDataSource{} }

type imageModel struct {
	ID        types.String `tfsdk:"id"`
	OS        types.String `tfsdk:"os"`
	Label     types.String `tfsdk:"label"`
	IsWindows types.Bool   `tfsdk:"is_windows"`
	Active    types.Bool   `tfsdk:"active"`
}

type imagesModel struct {
	Images []imageModel `tfsdk:"images"`
}

func (d *imagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_images"
}

func (d *imagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The OS images you can deploy a `noderush_node` from.",
		Attributes: map[string]schema.Attribute{
			"images": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Image id to pass to `noderush_node.image_id`."},
						"os":         schema.StringAttribute{Computed: true, MarkdownDescription: "OS family, e.g. `ubuntu`, `windows`."},
						"label":      schema.StringAttribute{Computed: true, MarkdownDescription: "Human-readable label."},
						"is_windows": schema.BoolAttribute{Computed: true, MarkdownDescription: "True for Windows images."},
						"active":     schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the image can be deployed."},
					},
				},
			},
		},
	}
}

func (d *imagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *imagesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	images, err := d.client.ListImages(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Could not list images", err.Error())
		return
	}
	state := imagesModel{Images: make([]imageModel, 0, len(images))}
	for _, im := range images {
		state.Images = append(state.Images, imageModel{
			ID:        types.StringValue(im.ID),
			OS:        types.StringValue(im.OS),
			Label:     types.StringValue(im.Label),
			IsWindows: types.BoolValue(im.IsWindows),
			Active:    types.BoolValue(im.Active),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
