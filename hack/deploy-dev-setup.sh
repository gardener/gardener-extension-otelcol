#!/usr/bin/env bash
# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -e

_SCRIPT_NAME="${0##*/}"
_SCRIPT_DIR=$( dirname "$( readlink -f -- "${0}" )" )
_PROJECT_DIR="${_SCRIPT_DIR}/.."
_TOOLS_MOD_FILE="${_PROJECT_DIR}/internal/tools/go.mod"
_KUSTOMIZE=$( go tool -n -modfile "${_TOOLS_MOD_FILE}" kustomize )
_YQ=$( go tool -n -modfile "${_TOOLS_MOD_FILE}" yq )

# Validates that expected env vars are set
function _validate_env_vars() {
  local _want_vars=(
    EXTENSION_IMAGE
    KUBECONFIG
  )

  for _var in "${_want_vars[@]}"; do
    if [ -z "${!_var}" ]; then
      echo "Required var ${_var} is not set"
      exit 1
    fi
  done
}

# Main entrypoint
function _main() {
  _validate_env_vars

  local _digest=""
  local _image=""
  local _tag=""

  _digest=$( docker image inspect "${EXTENSION_IMAGE}" | ${_YQ} '.[0].RepoDigests[0]' )

  if [[ -z "${_digest}" ]]; then
    echo "No digest found for image ${EXTENSION_IMAGE}"
    exit 1
  fi

  _image=$( echo "${_digest}" | cut -d '@' -f 1 )
  _tag=$( echo "${_digest}" | cut -d '@' -f 2 )

  export IMAGE_REPO="${_image}" IMAGE_TAG="${_tag}"

  "${_KUSTOMIZE}" build "${_PROJECT_DIR}/examples/dev-setup" | \
    "${_YQ}" \
      'select(.kind == "ControllerDeployment" and .metadata.name == "otelcol").helm.values.image.repository |= strenv(IMAGE_REPO) |
       select(.kind == "ControllerDeployment" and .metadata.name == "otelcol").helm.values.image.tag |= strenv(IMAGE_TAG)' | \
         kubectl apply --server-side --force-conflicts=true -f -
}

_main "$@"
