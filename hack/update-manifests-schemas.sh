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
    while read -r _manifest_dir; do
      local _version="${_manifest_dir##*/}"
      local _group="$( basename "$( dirname "${_manifest_dir}" )" )"
      local _pkg_dir="${_GARDENER_APIS_DIRS}/${_group%%.*}/${_version}"

      echo "Updating manifests ${_group}/${_version} ..."
      # Found float, the usage of which is highly discouraged, as support for them varies across languages.
      # Please consider serializing your float as string instead. If you are really sure you want to use them, re-run with crd:allowDangerousTypes=true
      _go_tool controller-gen crd:allowDangerousTypes=true \
        paths="${_pkg_dir}" \
        output:crd:dir="${_manifest_dir}"

      echo "Updating schemas ${_group}/${_version} ..."
      local _schema_dir="${_SCHEMAS_DIR}/${_group}/${_version}"
      mkdir -p "${_schema_dir}"
      for _crd_file in "${_manifest_dir}"/*.yaml; do
        [[ -f "${_crd_file}" ]] || continue
        ( cd "${_schema_dir}" && FILENAME_FORMAT="{kind}" _go_tool openapi2jsonschema-go "${_crd_file}" )
      done
    done < <( find "${_MANIFESTS_DIR}" -mindepth 2 -maxdepth 2 -type d )
}

_main "$@"
