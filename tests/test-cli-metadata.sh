#!/bin/bash

# CLI tests for testing valid/invalid entity metadata.

set -euxo pipefail

OASIS_REGISTRY=${PWD}/oasis-registry/oasis-registry
FIXTURES_DIR=${PWD}/tests/fixtures/

if [[ -n ${LEDGER_SIGNER_PATH:-""} ]]; then
	SIGNER="ledger"
	ENTITY_UPDATE_FLAGS=(
		--signer.dir ./
		--signer.backend plugin
		--signer.plugin.name ledger
		--signer.plugin.path "$LEDGER_SIGNER_PATH"
		# Disable entity update CLI command's internal validation of the
		# provided entity metadata to test the Ledger-based signer's validation.
		--skip-validation
	)
	# Error message prefix when entity metadata update is rejected by the
	# Ledger-based signer.
	LEDGER_ERROR="signature/signer/plugin: failed to sign: ledger: failed to sign message: ledger/oasis: failed to sign:"
else
	SIGNER="file"
	ENTITY_UPDATE_FLAGS=(
		--assume_yes
		--signer.dir ${FIXTURES_DIR}/entity-1
	)
fi

# Terminal colors.
RED='\033[0;31m'
OFF='\033[0m'

# Print the given argument to stderr in red color.
print_error() {
	printf "${RED}$1${OFF}\n" >&2
}

# Call "oasis-registry entity update" command with the given metadata fixture
# and verify registry's state afterwards.
update_entity_and_verify() {
	local fixture=metadata/$1
	local kind=$2
	if ${OASIS_REGISTRY} entity update "${ENTITY_UPDATE_FLAGS[@]}" \
		"${FIXTURES_DIR}/$fixture" > >(tee cmd.out) 2>&1; then
		# Entity update completed successfully.
		if [[ "$kind" != "valid" ]]; then
			print_error "Invalid entity update with fixture $fixture should fail."
			exit 1
		fi
	else
		# Entity update failed.
		if [[ "$kind" = "valid" ]]; then
			print_error "Valid entity update with fixture $fixture should not fail."
			exit 1
		fi
		# Ensure invalid entity metadata update was rejected by the Ledger-based
		# signer itself.
		if [[ $SIGNER == "ledger" ]]; then
			if ! grep -q "$LEDGER_ERROR" cmd.out; then
				print_error "Invalid entity update with fixture $fixture should be rejected by the Ledger-based signer itself."
				exit 1
			fi
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

# Update entity metadata with too long field values.
update_entity_and_verify update-name-too-long.json invalid
update_entity_and_verify update-url-too-long.json invalid
update_entity_and_verify update-email-too-long.json invalid
update_entity_and_verify update-keybase-too-long.json invalid
update_entity_and_verify update-twitter-too-long.json invalid

# Cleanup if everything went well.
rm -rf ${REGISTRY_DIR}
