// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp-demoapp/hashicups-client-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &orderResource{}
	_ resource.ResourceWithConfigure   = &orderResource{}
	_ resource.ResourceWithImportState = &orderResource{}
)

var IRRELEVANT_INT int = 42

// NewOrderResource is a helper function to simplify the provider implementation.
func NewOrderResource() resource.Resource {
	return &orderResource{}
}

// orderResource is the resource implementation.
type orderResource struct {
	client *hashicups.Client
}

// orderResourceModel maps the resource schema data.
type orderResourceModel struct {
	ID    types.String     `tfsdk:"id"`
	Items []orderItemModel `tfsdk:"item"`
}

// orderItemModel maps order item data.
type orderItemModel struct {
	FlagA types.Bool `tfsdk:"flag_a"`
	FlagB types.Bool `tfsdk:"flag_b"`
}

type theModifier struct{}

func (*theModifier) Description(context.Context) string {
	return "Workaround modifier for setting default values for flags inside set nested objects"

}

func (m *theModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (*theModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	var diags diag.Diagnostics

	configElements := req.ConfigValue.Elements()
	planElements := make([]attr.Value, len(configElements))
	for i, element := range configElements {
		objElement := element.(basetypes.ObjectValue)
		attrs := objElement.Attributes()
		var flagA attr.Value = types.BoolValue(false)
		if !attrs["flag_a"].IsUnknown() && !attrs["flag_a"].IsNull() {
			flagA = attrs["flag_a"]
		}
		var flagB attr.Value = types.BoolValue(false)
		if !attrs["flag_b"].IsUnknown() && !attrs["flag_b"].IsNull() {
			flagB = attrs["flag_b"]
		}
		planElements[i], diags = types.ObjectValue(
			objElement.AttributeTypes(ctx),
			map[string]attr.Value{
				"flag_a": flagA,
				"flag_b": flagB,
			},
		)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}
	}
	resp.PlanValue, diags = basetypes.NewSetValue(req.ConfigValue.ElementType(ctx), planElements)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
}

var myModifier planmodifier.Set = &theModifier{}

// Metadata returns the resource type name.
func (r *orderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_order"
}

// Schema defines the schema for the resource.
func (r *orderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an order.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric identifier of the order.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"item": schema.SetNestedBlock{
				PlanModifiers: []planmodifier.Set{
					myModifier,
				},
				// Description: "List of items in the order.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"flag_a": schema.BoolAttribute{
							Optional: true,
							Computed: true,
						},
						"flag_b": schema.BoolAttribute{
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *orderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*hashicups.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Create a new resource.
func (r *orderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan orderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(IRRELEVANT_INT))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *orderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state orderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *orderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan orderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *orderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state orderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *orderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
