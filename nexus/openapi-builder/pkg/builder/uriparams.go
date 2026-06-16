// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// uriParamRegex matches "{token}" placeholders in a URI template.
var uriParamRegex = regexp.MustCompile(`{([^{}]+)}`)

// buildURIParams builds the full OpenAPI parameter list for one URI:
//
//   - Path parameters extracted from the URI template (one per "{token}").
//   - Explicit header parameters declared on the URI.
//   - Parent-hierarchy parameters (one per ancestor CRD) emitted as
//     query parameters by default, unless options.HeaderAliases maps the
//     ancestor to an HTTP header, in which case it becomes a header param.
//
// `nodes` is a snapshot of every NodeInfo in the target domain, keyed by
// CRD type. It is used to look up parent descriptions and to identify
// singletons (which contribute no parameter).
//
// `nodesByName` is the same snapshot indexed by `NodeInfo.Name` for fast
// description lookup when resolving a path token's canonical groupKind.
//
// `restURI` carries the URI string, its PathParams alias map, and the
// explicit Headers list.
//
// `hierarchy` is the CRD's ParentHierarchy[] — a list of CRD types whose
// identities the URI implicitly inherits.
func buildURIParams(
	restURI RestURIs,
	headers []string,
	hierarchy []string,
	nodes map[string]NodeInfo,
	nodesByName map[string]NodeInfo,
	opts Options,
) []*openapi3.ParameterRef {
	rawParams := uriParamRegex.FindAllStringSubmatch(restURI.URI, -1)
	resolvedParams := resolveURITokens(rawParams, restURI.PathParams)

	parameters := make([]*openapi3.ParameterRef, 0, len(rawParams)+len(headers)+len(hierarchy))

	// Path parameters.
	for i, p := range rawParams {
		alias := p[1]
		canonical := resolvedParams[i][1]
		desc := "Name of the " + alias + " node"
		if ni, ok := nodesByName[canonical]; ok && ni.Description != "" {
			desc = ni.Description
		}
		parameters = append(parameters, &openapi3.ParameterRef{
			Value: openapi3.NewPathParameter(alias).
				WithRequired(true).
				WithSchema(openapi3.NewStringSchema()).
				WithDescription(desc),
		})
	}

	// Explicit header parameters declared on the URI.
	for _, h := range headers {
		desc := "Header for " + h
		if ni, ok := nodesByName[h]; ok && ni.Description != "" {
			desc = ni.Description
		}
		parameters = append(parameters, &openapi3.ParameterRef{
			Value: openapi3.NewHeaderParameter(h).
				WithRequired(true).
				WithSchema(openapi3.NewStringSchema()).
				WithDescription(desc),
		})
	}

	// Parent-hierarchy parameters.
	for _, parent := range hierarchy {
		ni, ok := nodes[parent]
		if !ok || ni.IsSingleton {
			continue
		}
		desc := ni.Description
		if desc == "" {
			desc = "Name of the " + ni.Name + " node"
		}

		// Ignored parents are dropped entirely UNLESS a HeaderAlias is
		// configured for them — in which case they are emitted as a
		// required header parameter (e.g. orgs.Org -> x-org-id).
		_, ignored := opts.IgnoredParents[ni.Name]
		if ignored {
			headerName, hasHeader := opts.HeaderAliases[ni.Name]
			if !hasHeader {
				continue
			}
			if paramExists(ni.Name, resolvedParams) ||
				headerExists(ni.Name, headers) ||
				rawHeaderExists(headerName, headers) {
				continue
			}
			parameters = append(parameters, &openapi3.ParameterRef{
				Value: openapi3.NewHeaderParameter(headerName).
					WithRequired(true).
					WithSchema(openapi3.NewStringSchema()).
					WithDescription(desc),
			})
			continue
		}

		// Non-ignored parents: emit as header if a HeaderAlias is
		// configured for this parent, else as a query parameter.
		// (api-gw's runtime behaviour: HeaderAliases applies even when
		// IgnoredParents is empty, so the alias acts as a "promote to
		// header" annotation.)
		if headerName, hasHeader := opts.HeaderAliases[ni.Name]; hasHeader {
			if paramExists(ni.Name, resolvedParams) ||
				headerExists(ni.Name, headers) ||
				rawHeaderExists(headerName, headers) {
				continue
			}
			parameters = append(parameters, &openapi3.ParameterRef{
				Value: openapi3.NewHeaderParameter(headerName).
					WithRequired(true).
					WithSchema(openapi3.NewStringSchema()).
					WithDescription(desc),
			})
			continue
		}

		if paramExists(ni.Name, resolvedParams) || headerExists(ni.Name, headers) {
			continue
		}
		parameters = append(parameters, &openapi3.ParameterRef{
			Value: openapi3.NewQueryParameter(ni.Name).
				WithRequired(true).
				WithSchema(openapi3.NewStringSchema()).
				WithDescription(desc),
		})
	}

	return parameters
}

// resolveURITokens rewrites each URI token to its canonical groupKind via
// the per-URI PathParams alias map. Output shape matches the input
// (so [][1] is the resolved token). Tokens not in PathParams pass
// through unchanged, preserving backward compatibility with pre-alias
// datamodels that use canonical form directly.
func resolveURITokens(rawParams [][]string, pathParams map[string]string) [][]string {
	if len(pathParams) == 0 {
		return rawParams
	}
	out := make([][]string, len(rawParams))
	for i, p := range rawParams {
		if len(p) < 2 {
			out[i] = p
			continue
		}
		if canonical, ok := pathParams[p[1]]; ok {
			resolved := make([]string, len(p))
			copy(resolved, p)
			resolved[1] = canonical
			out[i] = resolved
			continue
		}
		out[i] = p
	}
	return out
}

func paramExists(name string, params [][]string) bool {
	for _, p := range params {
		if len(p) >= 2 && p[1] == name {
			return true
		}
	}
	return false
}

func headerExists(name string, headers []string) bool {
	for _, h := range headers {
		if h == name {
			return true
		}
	}
	return false
}

func rawHeaderExists(headerName string, headers []string) bool {
	for _, h := range headers {
		if strings.EqualFold(h, headerName) {
			return true
		}
	}
	return false
}

// constructUpdateParam returns the "?update_if_exists=" query parameter
// appended to PUT operations on default URIs (CRUD endpoints). Mirrors
// the existing behaviour in both consumers.
func constructUpdateParam() *openapi3.ParameterRef {
	return &openapi3.ParameterRef{
		Value: openapi3.NewQueryParameter("update_if_exists").
			WithRequired(false).
			WithSchema(openapi3.NewBoolSchema()).
			WithDescription("If set to false, disables update of preexisting object. Default value is true"),
	}
}
