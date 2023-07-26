package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SAP/terraform-provider-btp/internal/btpcli"
	"github.com/SAP/terraform-provider-btp/internal/validation/uuidvalidator"
)

func newSubaccountEntitlementsDataSource() datasource.DataSource {
	return &subaccountEntitlementsDataSource{}
}

type subaccountEntitlementsDataSourceConfig struct {
	/* INPUT */
	SubaccountId types.String `tfsdk:"subaccount_id"`
	Id           types.String `tfsdk:"id"`
	/* OUTPUT */
	Values types.Map `tfsdk:"values"`
}

type subaccountEntitlementsDataSource struct {
	cli *btpcli.ClientFacade
}

func (ds *subaccountEntitlementsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_subaccount_entitlements", req.ProviderTypeName)
}

func (ds *subaccountEntitlementsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	ds.cli = req.ProviderData.(*btpcli.ClientFacade)
}

func (ds *subaccountEntitlementsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Gets all the entitlements and quota assignments for a subaccount.

To get all entitlements and quota assigned to a specific subaccount:
* You must be assigned to either the subaccount admin or subaccount viewer role.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{ // required by hashicorps terraform plugin testing framework
				DeprecationMessage:  "Use the `subaccount_id` attribute instead",
				MarkdownDescription: "The ID of the subaccount.",
				Computed:            true,
			}, "subaccount_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the subaccount.",
				Required:            true,
				Validators: []validator.String{
					uuidvalidator.ValidUUID(),
				},
			},
			"values": schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"service_name": schema.StringAttribute{
							MarkdownDescription: "The name of the entitled service.",
							Computed:            true,
						},
						"service_display_name": schema.StringAttribute{
							MarkdownDescription: "The display name of the entitled service.",
							Computed:            true,
						},
						"plan_name": schema.StringAttribute{
							MarkdownDescription: "The name of the entitled service plan.",
							Computed:            true,
						},
						"plan_display_name": schema.StringAttribute{
							MarkdownDescription: "The display name of the entitled service plan.",
							Computed:            true,
						},
						"plan_description": schema.StringAttribute{
							MarkdownDescription: "The description of the entitled service plan.",
							Computed:            true,
						},
						"quota_assigned": schema.Float64Attribute{
							MarkdownDescription: "The overall quota assigned.",
							Computed:            true,
						},
						"quota_remaining": schema.Float64Attribute{
							MarkdownDescription: "The quota, which is not used.",
							Computed:            true,
						},
						"category": schema.StringAttribute{
							MarkdownDescription: "The current state of the entitlement. Possible values are: \n " +
								getFormattedValueAsTableRow("value", "description") +
								getFormattedValueAsTableRow("---", "---") +
								getFormattedValueAsTableRow("`PLATFORM`", " A service required for using a specific platform; for example, Application Runtime is required for the Cloud Foundry platform.") +
								getFormattedValueAsTableRow("`SERVICE`", "A commercial or technical service. that has a numeric quota (amount) when entitled or assigned to a resource. When assigning entitlements of this type, use the 'amount' option.") +
								getFormattedValueAsTableRow("`ELASTIC_SERVICE`", "A commercial or technical service that has no numeric quota (amount) when entitled or assigned to a resource. Generally this type of service can be as many times as needed when enabled, but may in some cases be restricted by the service owner.") +
								getFormattedValueAsTableRow("`ELASTIC_LIMITED`", "An elastic service that can be enabled for only one subaccount per global account.") +
								getFormattedValueAsTableRow("`APPLICATION`", "A multitenant application to which consumers can subscribe. As opposed to applications defined as a 'QUOTA_BASED_APPLICATION', these applications do not have a numeric quota and are simply enabled or disabled as entitlements per subaccount.") +
								getFormattedValueAsTableRow("`QUOTA_BASED_APPLICATION`", "A multitenant application to which consumers can subscribe. As opposed to applications defined as 'APPLICATION', these applications have an numeric quota that limits consumer usage of the subscribed application per subaccount.") +
								getFormattedValueAsTableRow("`ENVIRONMENT`", " An environment service; for example, Cloud Foundry."),
							Computed: true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func (ds *subaccountEntitlementsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data subaccountEntitlementsDataSourceConfig

	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cliRes, _, err := ds.cli.Accounts.Entitlement.ListBySubaccount(ctx, data.SubaccountId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("API Error Reading Resource Entitlements (Subaccount)", fmt.Sprintf("%s", err))
		return
	}

	values := map[string]entitledService{}

	for _, service := range cliRes.EntitledServices {
		for _, servicePlan := range service.ServicePlans {
			values[fmt.Sprintf("%s:%s", service.Name, servicePlan.Name)] = entitledService{
				ServiceName:        types.StringValue(service.Name),
				ServiceDisplayName: types.StringValue(service.DisplayName),
				PlanName:           types.StringValue(servicePlan.Name),
				PlanDisplayName:    types.StringValue(servicePlan.DisplayName),
				PlanDescription:    types.StringValue(servicePlan.Description),
				QuotaAssigned:      types.Float64Value(servicePlan.Amount),
				QuotaRemaining:     types.Float64Value(servicePlan.RemainingAmount),
				Category:           types.StringValue(servicePlan.Category),
			}
		}
	}

	data.Id = data.SubaccountId
	data.Values, diags = types.MapValueFrom(ctx, types.ObjectType{AttrTypes: entitledServiceType()}, values)
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
