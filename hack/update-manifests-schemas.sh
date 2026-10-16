#!/usr/bin/env bash
# SPDX-FileCopyrightText: Contributors to the Gardener project
#
# SPDX-License-Identifier: Apache-2.0

set -e

_SCRIPT_NAME="${0##*/}"
_SCRIPT_DIR=$( dirname "$( readlink -f -- "${0}" )" )
_PROJECT_DIR="${_SCRIPT_DIR}/.."
_MANIFESTS_DIR="${_PROJECT_DIR}/test/manifests/crd"
_SCHEMAS_DIR="${_PROJECT_DIR}/test/schemas"
_GARDENER_APIS_DIRS="$( go list -m -f '{{ .Dir }}' github.com/gardener/gardener/pkg/apis )"

function _go_tool() {
    go tool -modfile "${_PROJECT_DIR}/internal/tools/go.mod" "$@";
}

# Main entrypoint
function _main() {
    echo "Gardener APIs dir: ${_GARDENER_APIS_DIRS}"

    # Manifests, used for testing.
    for _group_version in core.gardener.cloud/v1 core.gardener.cloud/v1beta1 operator.gardener.cloud/v1alpha1 extensions.gardener.cloud/v1alpha1 resources.gardener.cloud/v1alpha1; do
        local _version="${_group_version##*/}"
        local _group="${_group_version%/*}"
        local _manifest_dir="${_MANIFESTS_DIR}/${_group}/${_version}"
        local _pkg_dir="${_GARDENER_APIS_DIRS}/${_group%%.*}/${_version}"

        echo "Updating manifests ${_group}/${_version} ..."
        # allowDangerousTypes=true is required because operator/v1alpha1 uses float fields
        _go_tool controller-gen crd:allowDangerousTypes=true \
          paths="${_pkg_dir}" \
          output:crd:dir="${_manifest_dir}"
    done

    # Schemas, used for schema validation of examples/.
    for _group_version in core.gardener.cloud/v1 core.gardener.cloud/v1beta1 operator.gardener.cloud/v1alpha1 extensions.gardener.cloud/v1alpha1 resources.gardener.cloud/v1alpha1; do
        local _version="${_group_version##*/}"
        local _group="${_group_version%/*}"
        local _manifest_dir="${_MANIFESTS_DIR}/${_group}/${_version}"
        local _schema_dir="${_SCHEMAS_DIR}/${_group}/${_version}"

        echo "Updating schemas ${_group}/${_version} ..."
        mkdir -p "${_schema_dir}"
        for _crd_file in "${_manifest_dir}"/*.yaml; do
            [[ -f "${_crd_file}" ]] || continue
            ( cd "${_schema_dir}" && FILENAME_FORMAT="{kind}" _go_tool openapi2jsonschema-go "${_crd_file}" )
        done
    done

    # Remove these manifests, because they are needed only for schema generation
    # and not for the testing.
    rm -rf "${_MANIFESTS_DIR}"/core.gardener.cloud/v1
    rm -rf "${_MANIFESTS_DIR}"/core.gardener.cloud/v1beta1
    rm -rf "${_MANIFESTS_DIR}"/operator.gardener.cloud/v1alpha1
}

_main "$@"
