package parser

import (
	"regexp"
)

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

// pathParamExists checks if a name exists as a path parameter in the parsed URI params.
func pathParamExists(name string, params [][]string) bool {
	for _, p := range params {
		if len(p) >= 2 && p[1] == name {
			return true
		}
	}
	return false
}
