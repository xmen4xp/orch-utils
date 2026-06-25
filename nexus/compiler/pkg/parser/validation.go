package parser

import (
	"regexp"
	"strings"
)

// PathParamAliasFor returns the URI-token alias for a canonical
// "package.Kind" RestName by lowercasing the Kind portion. This mirrors the
// rule applied by the datamodel migrator and openapi-generator path-param
// aliasing.
//
//	"orgs.Org"       -> "org"
//	"spaces.Space"   -> "space"
//	"aislice.AISlice"-> "aislice"
func PathParamAliasFor(canonical string) string {
	idx := strings.LastIndex(canonical, ".")
	if idx < 0 {
		return strings.ToLower(canonical)
	}
	return strings.ToLower(canonical[idx+1:])
}

// pathParamAliases is a process-wide registry of alias -> set of canonicals
// declared via RestURIs.PathParams across all RestAPISpec values. The same
// alias may legitimately map to different canonical types in disjoint URI
// subtrees (e.g. {cluster} -> clusters.Clusters under /v1/inventory and
// {cluster} -> configcluster.ConfigCluster under /v1/config). Because URIs
// themselves are unique, there is no routing ambiguity; validators that need
// to disambiguate use ResolvePathParamAliasInHierarchy with the caller's
// hierarchy context.
var pathParamAliases = map[string]map[string]struct{}{}

// RegisterPathParamAlias records a URI-token alias declared by a RestURIs
// PathParams entry. Idempotent. Multiple canonicals are allowed per alias —
// disambiguation happens at resolution time via ResolvePathParamAliasInHierarchy.
func RegisterPathParamAlias(alias, canonical string) {
	if alias == "" || canonical == "" {
		return
	}
	if pathParamAliases[alias] == nil {
		pathParamAliases[alias] = map[string]struct{}{}
	}
	pathParamAliases[alias][canonical] = struct{}{}
}

// ResolvePathParamAlias returns one canonical "package.Kind" RestName for a
// previously registered alias, or "" if none was registered. If the alias is
// declared in multiple subtrees, the returned canonical is unspecified;
// callers that need hierarchy-correct resolution must use
// ResolvePathParamAliasInHierarchy.
func ResolvePathParamAlias(alias string) string {
	for c := range pathParamAliases[alias] {
		return c
	}
	return ""
}

// ResolvePathParamAliasInHierarchy returns the canonical RestName for an
// alias, picking the candidate that belongs to the caller-provided set of
// valid hierarchy nodes. Returns "" if no registered canonical matches.
// Used by ExtensionRestAPI validation, which knows its associated node's
// hierarchy and can pick the right canonical even when the alias is
// overloaded across disjoint subtrees.
func ResolvePathParamAliasInHierarchy(alias string, validNodes map[string]bool) string {
	for c := range pathParamAliases[alias] {
		if validNodes[c] {
			return c
		}
	}
	return ""
}

// ValidateRequiredParents checks that all non-singleton, non-ignored parent nodes
// appear as path parameters in the given URI. It returns two lists:
//   - missing: parent RestNames that are required but not found in the URI
//   - ignored: parent RestNames that are missing but configured as ignored
func ValidateRequiredParents(uri string, crdName string, parentsMap map[string]NodeHelper, ignoredParams []string) (missing []string, ignored []string) {
	r := regexp.MustCompile(`\{([^{}]+)\}`)
	uriParams := r.FindAllStringSubmatch(uri, -1)

	ignoredSet := make(map[string]struct{})
	for _, val := range ignoredParams {
		ignoredSet[val] = struct{}{}
	}

	crdHelper, ok := parentsMap[crdName]
	if !ok {
		return nil, nil
	}

	for _, parentCrd := range crdHelper.Parents {
		parentHelper, ok := parentsMap[parentCrd]
		if !ok {
			continue
		}

		if parentHelper.IsSingleton {
			continue
		}

		parentName := parentHelper.RestName

		if !pathParamExists(parentName, uriParams) {
			if _, isIgnored := ignoredSet[parentName]; isIgnored {
				ignored = append(ignored, parentName)
			} else {
				missing = append(missing, parentName)
			}
		}
	}

	return missing, ignored
}

// pathParamExists checks if a name exists as a path parameter in the parsed
// URI params. A canonical "package.Kind" RestName is also considered present
// when the URI uses its lowercase-Kind alias (e.g. "{org}" matches
// RestName "orgs.Org"). This supports path-param aliasing for both
// RestURIs (which carry an explicit PathParams map) and ExtensionRestAPI
// URIs (which do not).
func pathParamExists(name string, params [][]string) bool {
	formulaAlias := PathParamAliasFor(name)
	for _, p := range params {
		if len(p) < 2 {
			continue
		}
		token := p[1]
		// Direct canonical match or formula-derived alias.
		if token == name || token == formulaAlias {
			return true
		}
		// User-declared alias registered via RestURIs.PathParams (e.g.
		// {datacenter} -> datacenters.DataCenters). The alias may be
		// overloaded across disjoint subtrees, so check membership in the
		// set of canonicals registered for this token.
		if _, ok := pathParamAliases[token][name]; ok {
			return true
		}
	}
	return false
}
