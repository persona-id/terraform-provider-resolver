package provider

import (
	"context"

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

var _ resource.ResourceWithModifyPlan = (*MapResource)(nil)

func NewMapResource() resource.Resource {
	return &MapResource{}
}

type PlanOrState interface {
	Set(context.Context, interface{}) diag.Diagnostics
}

type MapResource struct{}

func (r *MapResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model mapModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	model.ID = types.StringValue("-")

	r.modify(ctx, model, &resp.Diagnostics, &resp.State, true)
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r *MapResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *MapResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_map"
}

func (r *MapResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Will be when the resource is being deleted.
	if req.Plan.Raw.IsNull() {
		return
	}

	var model mapModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.modify(ctx, model, &resp.Diagnostics, &resp.Plan, false)
}

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r *MapResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *MapResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attempts to resolve a map when possible instead of the entire map being unknown at plan.",

		Attributes: map[string]schema.Attribute{
			"keys": schema.ListAttribute{
				Description: "The list of keys, must be in same order as values.",
				ElementType: types.StringType,
				Required:    true,
			},
			"result_keys": schema.ListAttribute{
				Description: "The list of keys that should be in the result, must be a subset of keys.",
				ElementType: types.StringType,
				Required:    true,
			},
			"values": schema.ListAttribute{
				Description: "The list of values, must be in same order as keys.",
				ElementType: types.StringType,
				Required:    true,
			},

			// Computed
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "A static value used internally by Terraform, this should not be referenced in configurations.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"result": schema.MapAttribute{
				Computed:    true,
				Description: "The resolved mapping. If a result_key is unknown, this will be unknown.",
				ElementType: types.StringType,
			},
		},
	}
}

func (r *MapResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model mapModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.modify(ctx, model, &resp.Diagnostics, &resp.State, true)
}

func (r *MapResource) modify(ctx context.Context, model mapModel, diagnostics *diag.Diagnostics, state PlanOrState, errorOnUnresolved bool) {
	keys := make([]basetypes.StringValue, len(model.Keys.Elements()))
	diagnostics.Append(model.Keys.ElementsAs(ctx, &keys, false)...)
	if diagnostics.HasError() {
		return
	}

	resultKeys := make([]basetypes.StringValue, len(model.ResultKeys.Elements()))
	diagnostics.Append(model.ResultKeys.ElementsAs(ctx, &resultKeys, false)...)
	if diagnostics.HasError() {
		return
	}

	values := make([]basetypes.StringValue, len(model.Values.Elements()))
	diagnostics.Append(model.Values.ElementsAs(ctx, &values, false)...)
	if diagnostics.HasError() {
		return
	}

	if len(keys) > len(values) {
		diagnostics.AddAttributeError(path.Root("keys"), "Key count is higher than the number of values", "")
		diagnostics.AddAttributeError(path.Root("values"), "Value count is lower than the number of keys", "")
		return
	} else if len(keys) < len(values) {
		diagnostics.AddAttributeError(path.Root("keys"), "Key count is lower than the number of values", "")
		diagnostics.AddAttributeError(path.Root("values"), "Value count is higher than the number of keys", "")
		return
	} else if len(resultKeys) > len(keys) {
		diagnostics.AddAttributeError(path.Root("result_keys"), "Result key count is higher than the number of keys", "")
		return
	}

	model.Result = resolveMap(keys, resultKeys, values)

	if errorOnUnresolved {
		if model.Result.IsNull() || model.Result.IsUnknown() {
			diagnostics.AddError("Unable to resolve some result_keys, is it a subset of keys?", "")
			return
		}
	}

	diagnostics.Append(state.Set(ctx, model)...)
}

type mapModel struct {
	ID         types.String `tfsdk:"id"`
	Keys       types.List   `tfsdk:"keys"`
	Result     types.Map    `tfsdk:"result"`
	ResultKeys types.List   `tfsdk:"result_keys"`
	Values     types.List   `tfsdk:"values"`
}

func resolveMap(keys, resultKeys, values []basetypes.StringValue) basetypes.MapValue {
	keyValueMapping := make(map[string]string)
	keyValueUnknown := make(map[string]bool)
	keysUnknown := 0
	resultKeyMapping := make(map[string]bool)
	resultKeysUnknown := 0

	for i := 0; i < len(keys); i++ {
		if keys[i].IsUnknown() {
			keysUnknown += 1
			continue
		}

		if values[i].IsUnknown() {
			keyValueUnknown[keys[i].ValueString()] = true
		} else {
			keyValueMapping[keys[i].ValueString()] = values[i].ValueString()
		}
	}

	for _, resultKey := range resultKeys {
		if resultKey.IsUnknown() {
			return basetypes.NewMapUnknown(basetypes.StringType{})
		}

		resultKeyMapping[resultKey.ValueString()] = true
	}

	finalMapping := make(map[string]attr.Value)

	for resultKey, _ := range resultKeyMapping {
		if value, ok := keyValueMapping[resultKey]; ok {
			finalMapping[resultKey] = basetypes.NewStringValue(value)
		} else if _, ok := keyValueUnknown[resultKey]; ok {
			finalMapping[resultKey] = basetypes.NewStringUnknown()
		} else {
			resultKeysUnknown += 1
		}
	}

	if resultKeysUnknown > 0 {
		if resultKeysUnknown <= keysUnknown {
			return basetypes.NewMapUnknown(basetypes.StringType{})
		} else {
			return basetypes.NewMapNull(basetypes.StringType{})
		}
	}

	return basetypes.NewMapValueMust(types.StringType, finalMapping)
}
