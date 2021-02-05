#!/bin/bash

set -euxo pipefail

OASIS_REGISTRY=${PWD}/oasis-registry/oasis-registry
FIXTURES_DIR=${PWD}/tests/fixtures/entity-ledger

SIGNER_FLAGS=(
  --signer.dir ./
  --signer.backend plugin
  --signer.plugin.name ledger
  --signer.plugin.path "$LEDGER_SIGNER_PATH"
)

update_entity_and_verify() {
    local fixture=$1
    local kind=$2
    if ${OASIS_REGISTRY} entity update "${SIGNER_FLAGS[@]}" "${FIXTURES_DIR}/$fixture"; then
        if [[ "$kind" != "valid" ]]; then
            echo "Invalid entity update with fixture $fixture should fail."
            exit 1
        fi
    else
        if [[ "$kind" = "valid" ]]; then
            echo "Valid entity update with fixture $fixture should not fail."
            exit 1
        fi
    fi
    ${OASIS_REGISTRY} verify
}

# Create a temporary directory for the registry.
REGISTRY_DIR=$(mktemp -d --tmpdir oasis-registry-tests-XXXXXX)

cd ${REGISTRY_DIR}

# Initialize the registry.
${OASIS_REGISTRY} init

# Verify registry integrity.
${OASIS_REGISTRY} verify

# Create new entity metadata.
update_entity_and_verify metadata.json valid

# Update entity metadata with fields having maximum lengths.
update_entity_and_verify update-max-lengths.json valid

# Update entity metadata with invalid version.
update_entity_and_verify update-invalid-version.json invalid

# Update entity metadata with invalid serial number.
#update_entity_and_verify update-invalid-serial.json invalid

# Update entity metadata with too long field values.
update_entity_and_verify update-name-too-long.json invalid
update_entity_and_verify update-url-too-long.json invalid
update_entity_and_verify update-email-too-long.json invalid
update_entity_and_verify update-keybase-too-long.json invalid
update_entity_and_verify update-twitter-too-long.json invalid
