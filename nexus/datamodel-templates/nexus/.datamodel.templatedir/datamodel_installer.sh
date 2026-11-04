#!/bin/sh

set -eux

NAME=${NAME:-}
IMAGE="file://${NAME}"
TITLE=${TITLE:-}
SKIP_CRD_INSTALLATION=${SKIP_CRD_INSTALLATION:-false}
GRAPHQL_ENABLED=${GRAPHQL_ENABLED:-false}

# User can pass the custom HTTP URL where the graphql plugin can be downloaded via internet
GRAPHQL_PATH=${GRAPHQL_PATH:-NA}

echo '
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  creationTimestamp: null
  name: datamodels.nexus.com
spec:
  conversion:
    strategy: None
  group: nexus.com
  names:
    kind: Datamodel
    listKind: DatamodelList
    plural: datamodels
    shortNames:
    - datamodel
    singular: datamodel
  scope: Cluster
  versions:
  - name: v1
    schema:
      openAPIV3Schema:
        type: object
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            properties:
              name:
                type: string
              url:
                type: string
              title:
                type: string
                default: "Datamodel APIs"
              enableGraphql:
                type: boolean
                default: false
              graphqlPath:
                type: string
                default: ""
            type: object
    served: true
    storage: true
' | kubectl apply -f -

# This is to support older way of installing datamodel from local folder
if [ "$SKIP_CRD_INSTALLATION" = "false" ]; then
    # Multi-version Kinds' CRDs (conversion.strategy: Webhook) are generated
    # with placeholder tokens for the nexus-api-gw Service's name/namespace
    # (see orch-utils/nexus/compiler/pkg/generator.ConversionWebhookServiceName{,Namespace}Placeholder):
    # that Service's real name depends on the Helm release name nexus-runtime
    # was installed under, which isn't known at datamodel compile time, but
    # every nexus-api-gw Service carries the same stable label regardless of
    # release name. Resolve it here and substitute it in before applying -
    # plain `kubectl apply` validates
    # spec.conversion.webhook.clientConfig.service.name as a DNS label, so
    # the placeholder text itself would be rejected by the API server.
    #
    # Only bother at all if some CRD actually needs it: a datamodel with no
    # multi-version Kinds shouldn't be made to depend on nexus-runtime.
    # This install runs early relative to the rest of a release (often
    # effectively a pre-install step), and nexus-runtime is typically a
    # *sibling subchart of this very release* in production - so on a
    # first install, blocking/polling for its Service wouldn't just be
    # slow, it would never succeed (nexus-runtime's own resources haven't
    # been applied yet either). So: look it up (covers upgrades, and setups
    # where nexus-runtime is installed separately/beforehand); if not
    # found, fall back to guessing the name nexus-runtime's own chart will
    # produce for itself - "<release>-nexus-api-gw" - using this release's
    # own name (HELM_RELEASE_NAME, templated in from `.Release.Name`),
    # which is correct whenever nexus-runtime shares a release with this
    # datamodel.
    NEEDS_API_GW_SVC=$(sed -n '/__NEXUS_API_GW_SERVICE_NAME__/p' /crds/*.yaml 2>/dev/null)
    if [ -n "$NEEDS_API_GW_SVC" ]; then
        # No "default" fallback here deliberately: this Job always runs in
        # the same namespace it's installed into, which varies per
        # deployment, and every pod gets this file via its mounted
        # ServiceAccount token unless automountServiceAccountToken is
        # explicitly disabled (not the case here) - silently guessing a
        # namespace on read failure would misconfigure the webhook target
        # the same way the hardcoded "nexus-api-gw"/empty-namespace bugs
        # this replaces did. Fail loudly on a genuinely empty/missing file
        # (checked explicitly below) rather than silently assuming one -
        # but the `|| true` here is still required: POSIX `read` exits
        # non-zero when its input ends without a trailing newline (as this
        # file always does), even though it assigned $INSTALL_NS correctly,
        # and `set -e` would otherwise abort the script right here on that
        # non-error.
        INSTALL_NS=""
        read -r INSTALL_NS < /var/run/secrets/kubernetes.io/serviceaccount/namespace || true
        if [ -z "$INSTALL_NS" ]; then
            echo "ERROR: could not determine this Job's namespace from /var/run/secrets/kubernetes.io/serviceaccount/namespace" >&2
            exit 1
        fi
        API_GW_SVC=$(kubectl get svc -n "$INSTALL_NS" -l app.kubernetes.io/component=nexus-api-gw -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
        if [ -z "$API_GW_SVC" ]; then
            if [ -z "$HELM_RELEASE_NAME" ]; then
                echo "ERROR: no Service labeled app.kubernetes.io/component=nexus-api-gw found in namespace $INSTALL_NS, and HELM_RELEASE_NAME is unset so the fallback name can't be guessed; cannot resolve conversion webhook target for multi-version CRDs" >&2
                exit 1
            fi
            API_GW_SVC="${HELM_RELEASE_NAME}-nexus-api-gw"
        fi
        # cert-manager's cainjector populates
        # spec.conversion.webhook.clientConfig.caBundle itself from whatever
        # Certificate the CRD's cert-manager.io/inject-ca-from annotation
        # points at. nexus-runtime's helm chart always names that
        # Certificate "<release>-nexus-api-gw-webhook" - the same
        # "<release>-nexus-api-gw" prefix as the Service we just resolved,
        # so reuse it rather than re-deriving the release name.
        CA_INJECT_FROM="${INSTALL_NS}/${API_GW_SVC}-webhook"
        sed -i "s/__NEXUS_API_GW_SERVICE_NAME__/${API_GW_SVC}/g; s/__NEXUS_API_GW_SERVICE_NAMESPACE__/${INSTALL_NS}/g; s#__NEXUS_API_GW_CA_INJECT_FROM__#${CA_INJECT_FROM}#g" /crds/*.yaml
    fi
    kubectl apply -f /crds --recursive
    [ "$GRAPHQL_PATH" != "NA" ] && GRAPHQL_ENABLED=true
    if [ -f /build/server ]; then
        GRAPHQL_ENABLED=true
    fi
fi

echo "$NAME"
# We will create datamodel object
if [ -n "$NAME" ] && [ -n "$IMAGE" ]; then
  if [ -n "$TITLE" ]; then
    echo '
      apiVersion: nexus.com/v1
      kind: Datamodel
      metadata:
        name: '"$NAME"'
      spec:
        name: '"$NAME"'
        url: '"$IMAGE"'
        title: '"$TITLE"'
        enableGraphql: '"$GRAPHQL_ENABLED"'' | kubectl apply -f -
  else
    echo '
    apiVersion: nexus.com/v1
    kind: Datamodel
    metadata:
      name: '"$NAME"'
    spec:
      name: '"$NAME"'
      url: '"$IMAGE"'
      enableGraphql: '"$GRAPHQL_ENABLED"'' | kubectl apply -f -
  fi
fi
