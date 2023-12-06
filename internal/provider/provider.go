// Copyright (c) Persona
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure Resolver satisfies various provider interfaces.
var _ provider.Provider = &Resolver{}

// Resolver defines the provider implementation.
type Resolver struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Resolver{
			version: version,
		}
	}
}

func (p *Resolver) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *Resolver) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func (p *Resolver) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "resolver"
	resp.Version = p.version
}

func (p *Resolver) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMapResource,
	}
}

func (p *Resolver) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform provider provides a resource that provides a resolution between keys and values when a subset is unknown to prevent unnessary plan diffs that are no-ops at apply.",
	}
}
