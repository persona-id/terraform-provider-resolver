// Copyright (c) Persona
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceMap(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"resolver": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: `
				resource "resolver_map" "test" {
					keys        = ["a", "b", "c"]
					result_keys = ["a", "c"]
					values      = ["1", "2", "3"]
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("resolver_map.test", "keys.#", "3"),
					resource.TestCheckResourceAttr("resolver_map.test", "keys.0", "a"),
					resource.TestCheckResourceAttr("resolver_map.test", "keys.1", "b"),
					resource.TestCheckResourceAttr("resolver_map.test", "keys.2", "c"),
					resource.TestCheckResourceAttr("resolver_map.test", "result_keys.#", "2"),
					resource.TestCheckResourceAttr("resolver_map.test", "result_keys.0", "a"),
					resource.TestCheckResourceAttr("resolver_map.test", "result_keys.1", "c"),
					resource.TestCheckResourceAttr("resolver_map.test", "values.#", "3"),
					resource.TestCheckResourceAttr("resolver_map.test", "values.0", "1"),
					resource.TestCheckResourceAttr("resolver_map.test", "values.1", "2"),
					resource.TestCheckResourceAttr("resolver_map.test", "values.2", "3"),
					resource.TestCheckResourceAttr("resolver_map.test", "result.%", "2"),
					resource.TestCheckResourceAttr("resolver_map.test", "result.a", "1"),
					resource.TestCheckResourceAttr("resolver_map.test", "result.c", "3"),
				),
			},
		},
	})
}

func TestAccResourceMapMoreKeysThanValues(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ErrorCheck: func(err error) error {
			return err
		},
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"resolver": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: `
				resource "resolver_map" "test" {
					keys        = ["a", "b", "c"]
					result_keys = ["a", "c"]
					values      = ["1", "2"]
				}
				`,

				ExpectError: regexp.MustCompile(`(?s)(Key count is higher than the number of values).*(Value count is lower than the number of keys)`),
			},
		},
	})
}

func TestAccResourceMapFewerKeysThanValues(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ErrorCheck: func(err error) error {
			return err
		},
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"resolver": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: `
				resource "resolver_map" "test" {
					keys        = ["a", "b"]
					result_keys = ["a", "c"]
					values      = ["1", "2", "3"]
				}
				`,

				ExpectError: regexp.MustCompile(`(?s)(Key count is lower than the number of values).*(Value count is higher than the number of keys)`),
			},
		},
	})
}

func TestAccResourceMapTooManyResultKeys(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ErrorCheck: func(err error) error {
			return err
		},
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"resolver": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: `
				resource "resolver_map" "test" {
					keys        = ["a", "b"]
					result_keys = ["a", "b", "c"]
					values      = ["1", "2"]
				}
				`,

				ExpectError: regexp.MustCompile(`(Result key count is higher than the number of keys)`),
			},
		},
	})
}

func TestInternalResolveMap(t *testing.T) {
	var tests = []struct {
		keys, resultKeys, values []basetypes.StringValue
		expectedResult           basetypes.MapValue
	}{
		// basic cases
		{
			keys: []basetypes.StringValue{
				basetypes.NewStringValue("a"),
				basetypes.NewStringValue("b"),
				basetypes.NewStringValue("c"),
			},
			resultKeys: []basetypes.StringValue{
				basetypes.NewStringValue("a"),
				basetypes.NewStringValue("c"),
			},
			values: []basetypes.StringValue{
				basetypes.NewStringValue("1"),
				basetypes.NewStringValue("2"),
				basetypes.NewStringValue("3"),
			},
			expectedResult: basetypes.NewMapValueMust(types.StringType, map[string]attr.Value{
				"a": basetypes.NewStringValue("1"),
				"c": basetypes.NewStringValue("3"),
			}),
		},
		// some keys unknown, but all result keys known
		{
			keys: []basetypes.StringValue{
				basetypes.NewStringValue("a"),
				basetypes.NewStringUnknown(),
				basetypes.NewStringValue("c"),
			},
			resultKeys: []basetypes.StringValue{
				basetypes.NewStringValue("a"),
				basetypes.NewStringValue("c"),
			},
			values: []basetypes.StringValue{
				basetypes.NewStringValue("1"),
				basetypes.NewStringValue("2"),
				basetypes.NewStringValue("3"),
			},
			expectedResult: basetypes.NewMapValueMust(types.StringType, map[string]attr.Value{
				"a": basetypes.NewStringValue("1"),
				"c": basetypes.NewStringValue("3"),
			}),
		},
		// some key value pairs unknown, but all result keys known
		{
			keys: []basetypes.StringValue{
				basetypes.NewStringValue("a"),
				basetypes.NewStringUnknown(),
				basetypes.NewStringValue("c"),
			},
			resultKeys: []basetypes.StringValue{
				basetypes.NewStringValue("a"),
				basetypes.NewStringValue("c"),
			},
			values: []basetypes.StringValue{
				basetypes.NewStringValue("1"),
				basetypes.NewStringUnknown(),
				basetypes.NewStringValue("3"),
			},
			expectedResult: basetypes.NewMapValueMust(types.StringType, map[string]attr.Value{
				"a": basetypes.NewStringValue("1"),
				"c": basetypes.NewStringValue("3"),
			}),
		},
		// some key values unknown, but all result keys known
		{
			keys: []basetypes.StringValue{
				basetypes.NewStringValue("a"),
				basetypes.NewStringValue("b"),
				basetypes.NewStringValue("c"),
			},
			resultKeys: []basetypes.StringValue{
				basetypes.NewStringValue("a"),
				basetypes.NewStringValue("c"),
			},
			values: []basetypes.StringValue{
				basetypes.NewStringValue("1"),
				basetypes.NewStringUnknown(),
				basetypes.NewStringValue("3"),
			},
			expectedResult: basetypes.NewMapValueMust(types.StringType, map[string]attr.Value{
				"a": basetypes.NewStringValue("1"),
				"c": basetypes.NewStringValue("3"),
			}),
		},
		// some key values unknown, including some for result keys
		{
			keys: []basetypes.StringValue{
				basetypes.NewStringValue("a"),
				basetypes.NewStringValue("b"),
				basetypes.NewStringValue("c"),
			},
			resultKeys: []basetypes.StringValue{
				basetypes.NewStringValue("a"),
				basetypes.NewStringValue("c"),
			},
			values: []basetypes.StringValue{
				basetypes.NewStringValue("1"),
				basetypes.NewStringValue("2"),
				basetypes.NewStringUnknown(),
			},
			expectedResult: basetypes.NewMapValueMust(types.StringType, map[string]attr.Value{
				"a": basetypes.NewStringValue("1"),
				"c": basetypes.NewStringUnknown(),
			}),
		},
		// not all result keys known
		{
			keys: []basetypes.StringValue{
				basetypes.NewStringValue("a"),
				basetypes.NewStringValue("b"),
				basetypes.NewStringValue("c"),
			},
			resultKeys: []basetypes.StringValue{
				basetypes.NewStringUnknown(),
				basetypes.NewStringValue("c"),
			},
			values: []basetypes.StringValue{
				basetypes.NewStringValue("1"),
				basetypes.NewStringValue("2"),
				basetypes.NewStringValue("3"),
			},
			expectedResult: basetypes.NewMapUnknown(types.StringType),
		},
	}

	for _, test := range tests {
		testname := fmt.Sprintf("%+v,%+v,%+v,%+v", test.keys, test.resultKeys, test.values, test.expectedResult)

		t.Run(testname, func(t *testing.T) {
			actualResult := resolveMap(test.keys, test.resultKeys, test.values)

			if !reflect.DeepEqual(test.expectedResult, actualResult) {
				t.Errorf("Got %+v, wanted %+v", actualResult, test.expectedResult)
			}
		})
	}
}
