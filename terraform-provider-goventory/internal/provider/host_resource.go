package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// same host struct saved in the DB
type goventoryHost struct {
	ID        string `json:"id"`
	Hostname  string `json:"hostname"`
	IPAddress string `json:"ip_address"`
	HostGroup string `json:"host_group"`
}

// resource implementation
type hostResource struct {
	client    *http.Client
	serverURL string
}

// hostResourceModel maps the resource schema data.
type hostResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Hostname  types.String `tfsdk:"hostname"`
	IPAddress types.String `tfsdk:"ip_address"`
	HostGroup types.String `tfsdk:"host_group"`
}

// factory function that returns new host rsrc
func NewHostResource() resource.Resource {
	return &hostResource{}
}

// rsrc metadata
func (r *hostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

// Schema defines the resource's schema.
func (r *hostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a host in GoVentory.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Host's UUID.",
				Computed:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "Name of the host.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ip_address": schema.StringAttribute{
				Description: "Host's IP address.",
				Optional:    true,
			},
			"host_group": schema.StringAttribute{
				Description: "Group the host belongs to.",
				Required:    true,
			},
		},
	}
}

// configure the resource
func (r *hostResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected map[string]interface{}, got: %%T. Please report this issue to the devs.", req.ProviderData),
		)
		return
	}

	r.client = providerData["client"].(*http.Client)
	r.serverURL = providerData["server_url"].(string)
}

// create rsrc
func (r *hostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hostResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := goventoryHost{
		Hostname:  plan.Hostname.ValueString(),
		IPAddress: plan.IPAddress.ValueString(),
		HostGroup: plan.HostGroup.ValueString(),
	}

	body, err := json.Marshal(host)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal host data", err.Error())
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/v1/host", r.serverURL), bytes.NewBuffer(body))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create HTTP request", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create host", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError("Failed to create host", fmt.Sprintf("unexpected status code: %d", httpResp.StatusCode))
		return
	}

	var createdHost goventoryHost
	err = json.NewDecoder(httpResp.Body).Decode(&createdHost)
	if err != nil {
		resp.Diagnostics.AddError("Failed to decode response body", err.Error())
		return
	}

	plan.ID = types.StringValue(createdHost.ID)
	plan.Hostname = types.StringValue(createdHost.Hostname)
	plan.IPAddress = types.StringValue(createdHost.IPAddress)
	plan.HostGroup = types.StringValue(createdHost.HostGroup)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// read the rsrc
func (r *hostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state hostResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/host/%s", r.serverURL, state.Hostname.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create HTTP request", err.Error())
		return
	}

	httpResp, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read host", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Failed to read host", fmt.Sprintf("unexpected status code: %d", httpResp.StatusCode))
		return
	}

	var host goventoryHost
	err = json.NewDecoder(httpResp.Body).Decode(&host)
	if err != nil {
		resp.Diagnostics.AddError("Failed to decode response body", err.Error())
		return
	}

	state.ID = types.StringValue(host.ID)
	state.Hostname = types.StringValue(host.Hostname)
	state.IPAddress = types.StringValue(host.IPAddress)
	state.HostGroup = types.StringValue(host.HostGroup)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// update the rsrc
func (r *hostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan hostResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := goventoryHost{
		IPAddress: plan.IPAddress.ValueString(),
		HostGroup: plan.HostGroup.ValueString(),
	}

	body, err := json.Marshal(host)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal host data", err.Error())
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/api/v1/host/%s", r.serverURL, plan.Hostname.ValueString()), bytes.NewBuffer(body))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create HTTP request", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update host", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Failed to update host", fmt.Sprintf("unexpected status code: %d", httpResp.StatusCode))
		return
	}

	var updatedHost goventoryHost
	err = json.NewDecoder(httpResp.Body).Decode(&updatedHost)
	if err != nil {
		resp.Diagnostics.AddError("Failed to decode response body", err.Error())
		return
	}

	plan.ID = types.StringValue(updatedHost.ID)
	plan.Hostname = types.StringValue(updatedHost.Hostname)
	plan.IPAddress = types.StringValue(updatedHost.IPAddress)
	plan.HostGroup = types.StringValue(updatedHost.HostGroup)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// delete the rsrc
func (r *hostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state hostResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/api/v1/host/%s", r.serverURL, state.Hostname.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create HTTP request", err.Error())
		return
	}

	httpResp, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete host", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError("Failed to delete host", fmt.Sprintf("unexpected status code: %d", httpResp.StatusCode))
		return
	}

	resp.State.RemoveResource(ctx)
}