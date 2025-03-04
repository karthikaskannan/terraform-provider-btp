package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SAP/terraform-provider-btp/internal/btpcli"
	"github.com/SAP/terraform-provider-btp/internal/validation/uuidvalidator"
)

func newDirectoryRoleCollectionAssignmentResource() resource.Resource {
	return &directoryRoleCollectionAssignmentResource{}
}

type directoryRoleCollectionAssignmentType struct {
	Id                 types.String `tfsdk:"id"`
	DirectoryId        types.String `tfsdk:"directory_id"`
	RoleCollectionName types.String `tfsdk:"role_collection_name"`
	Username           types.String `tfsdk:"user_name"`
	Groupname          types.String `tfsdk:"group_name"`
	AttributeName      types.String `tfsdk:"attribute_name"`
	AttributeValue     types.String `tfsdk:"attribute_value"`
	Origin             types.String `tfsdk:"origin"`
}

type directoryRoleCollectionAssignmentResource struct {
	cli *btpcli.ClientFacade
}

func (rs *directoryRoleCollectionAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_directory_role_collection_assignment", req.ProviderTypeName)
}

func (rs *directoryRoleCollectionAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	rs.cli = req.ProviderData.(*btpcli.ClientFacade)
}

func (rs *directoryRoleCollectionAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Assigns a user to a role collection on a directory level.`,
		Attributes: map[string]schema.Attribute{
			"directory_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the directory.",
				Required:            true,
				Validators: []validator.String{
					uuidvalidator.ValidUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_collection_name": schema.StringAttribute{
				MarkdownDescription: "The name of the role collection.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"id": schema.StringAttribute{ // required by hashicorps terraform plugin testing framework
				DeprecationMessage:  "Use the `directory_id` and `role_collection_name` attributes instead",
				MarkdownDescription: "The combined unique ID of the role collection.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_name": schema.StringAttribute{
				MarkdownDescription: "The username of the user to assign.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("user_name"), path.MatchRoot("group_name"), path.MatchRoot("attribute_name")),
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"group_name": schema.StringAttribute{
				MarkdownDescription: "The name of the group to assign.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.AlsoRequires(path.MatchRoot("origin")),
				},
			},
			"attribute_name": schema.StringAttribute{
				MarkdownDescription: "The name of the attribute to assign.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("attribute_value")),
					stringvalidator.AlsoRequires(path.MatchRoot("origin")),
					stringvalidator.LengthAtLeast(1),
				},
			},
			"attribute_value": schema.StringAttribute{
				MarkdownDescription: "The value of the attribute to assign.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("attribute_name")),
					stringvalidator.LengthAtLeast(1),
				},
			},
			"origin": schema.StringAttribute{
				MarkdownDescription: "The identity provider that hosts the user or a group. Only needed for custom identity provider.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("ldap"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (rs *directoryRoleCollectionAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state directoryRoleCollectionAssignmentType

	diags := req.State.Get(ctx, &state)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// This resource is not supposed to be read by definition. However nothing the user can do about that, hence no error message is raised via resp.Diagnostics.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (rs *directoryRoleCollectionAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan directoryRoleCollectionAssignmentType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	if !plan.Username.IsNull() {
		// assign user
		_, _, err = rs.cli.Security.RoleCollection.AssignUserByDirectory(ctx, plan.DirectoryId.ValueString(), plan.RoleCollectionName.ValueString(), plan.Username.ValueString(), plan.Origin.ValueString())
	} else if !plan.Groupname.IsNull() {
		// assign group
		_, _, err = rs.cli.Security.RoleCollection.AssignGroupByDirectory(ctx, plan.DirectoryId.ValueString(), plan.RoleCollectionName.ValueString(), plan.Groupname.ValueString(), plan.Origin.ValueString())
	} else {
		// assign attribute
		_, _, err = rs.cli.Security.RoleCollection.AssignAttributeByDirectory(ctx, plan.DirectoryId.ValueString(), plan.RoleCollectionName.ValueString(), plan.AttributeName.ValueString(), plan.AttributeValue.ValueString(), plan.Origin.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError("API Error Creating Resource Role Collection Assignment (Directory)", fmt.Sprintf("%s", err))
		return
	}

	// Setting ID of state - required by hashicorps terraform plugin testing framework for Create. See issue https://github.com/hashicorp/terraform-plugin-testing/issues/84
	plan.Id = types.StringValue(fmt.Sprintf("%s,%s,%s", plan.DirectoryId.ValueString(), plan.RoleCollectionName.ValueString(), plan.Username.ValueString()))

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (rs *directoryRoleCollectionAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan directoryRoleCollectionAssignmentType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// since all the attributes are marked to be replaced in case of update, this should never be reached.
	resp.Diagnostics.AddError("API Error Updating Resource Role Collection Assignment (Directory)", "This resource is not supposed to be updated")

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (rs *directoryRoleCollectionAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state directoryRoleCollectionAssignmentType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	if !state.Username.IsNull() {
		// unassign user
		_, _, err = rs.cli.Security.RoleCollection.UnassignUserByDirectory(ctx, state.DirectoryId.ValueString(), state.RoleCollectionName.ValueString(), state.Username.ValueString(), state.Origin.ValueString())
	} else if !state.Groupname.IsNull() {
		// unassign group
		_, _, err = rs.cli.Security.RoleCollection.UnassignGroupByDirectory(ctx, state.DirectoryId.ValueString(), state.RoleCollectionName.ValueString(), state.Groupname.ValueString(), state.Origin.ValueString())
	} else {
		// unassign attribute
		_, _, err = rs.cli.Security.RoleCollection.UnassignAttributeByDirectory(ctx, state.DirectoryId.ValueString(), state.RoleCollectionName.ValueString(), state.AttributeName.ValueString(), state.AttributeValue.ValueString(), state.Origin.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError("API Error Deleting Resource Role Collection Assignment (Directory)", fmt.Sprintf("%s", err))
		return
	}
}

func (rs *directoryRoleCollectionAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError(
		"Import Not Supported",
		"Import is not supported for this resource. Use the resource directory_role_collection instead.",
	)
}
