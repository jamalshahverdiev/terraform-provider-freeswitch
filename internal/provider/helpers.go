package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// tfMapToGo converts a Terraform map(string) into a Go map. Null/unknown -> {}.
func tfMapToGo(ctx context.Context, m types.Map) (map[string]string, diag.Diagnostics) {
	out := map[string]string{}
	if m.IsNull() || m.IsUnknown() {
		return out, nil
	}
	diags := m.ElementsAs(ctx, &out, false)
	return out, diags
}

// goMapToTF converts a Go map into a Terraform map(string). Empty -> empty map.
func goMapToTF(ctx context.Context, m map[string]string) (types.Map, diag.Diagnostics) {
	if m == nil {
		m = map[string]string{}
	}
	return types.MapValueFrom(ctx, types.StringType, m)
}

// goMapToTFPreserveNull keeps the prior (plan/state) null when the API returns
// an empty map — otherwise Terraform reports "inconsistent result after apply"
// for configs that omit the attribute.
func goMapToTFPreserveNull(ctx context.Context, m map[string]string, prior types.Map) (types.Map, diag.Diagnostics) {
	if len(m) == 0 && (prior.IsNull() || prior.IsUnknown()) {
		return types.MapNull(types.StringType), nil
	}
	return types.MapValueFrom(ctx, types.StringType, m)
}
