#!/bin/bash

set -euxo pipefail

OASIS_REGISTRY=${PWD}/oasis-registry/oasis-registry
FIXTURES_DIR=${PWD}/tests/fixtures

# Create a temporary directory for the registry.
REGISTRY_DIR=$(mktemp -d --tmpdir oasis-registry-tests-XXXXXX)
mkdir -p ${REGISTRY_DIR}/fork-1
cd ${REGISTRY_DIR}/fork-1

# Make sure all test entity private keys have correct permissions.
find ${FIXTURES_DIR} -name entity.pem -exec chmod 600 {} \;

# Initialize the registry.
${OASIS_REGISTRY} init

# Verify registry integrity.
${OASIS_REGISTRY} verify

# Create new entity metadata.
${OASIS_REGISTRY} entity update \
	--signer.dir ${FIXTURES_DIR}/entity-1 \
	${FIXTURES_DIR}/entity-1/metadata.json

# Verify registry integrity.
${OASIS_REGISTRY} verify

# Create some more entities.
${OASIS_REGISTRY} entity update \
	--signer.dir ${FIXTURES_DIR}/entity-2 \
	${FIXTURES_DIR}/entity-2/metadata.json

# Verify registry integrity.
${OASIS_REGISTRY} verify

###############################################################
# Create a new fork of the registry and update entity metadata.
###############################################################
cd ${REGISTRY_DIR}
cp -a fork-1 fork-2
cd fork-2

# Update the first entity metadata.
${OASIS_REGISTRY} entity update \
    --signer.dir ${FIXTURES_DIR}/entity-1 \
    ${FIXTURES_DIR}/entity-1/update.json

# Verify registry integrity.
${OASIS_REGISTRY} verify
${OASIS_REGISTRY} verify --update ../fork-1

#############################################
# Create a bad fork that removes a statement.
#############################################
cd ${REGISTRY_DIR}
cp -a fork-1 fork-3
cd fork-3

# Remove a statement.
rm registry/entity/749c9846553512eb62d9828c0b54be04d18bd3961ff5137a9d5520c8017291c4.json

# Verify registry integrity (update verification should fail).
${OASIS_REGISTRY} verify
! ${OASIS_REGISTRY} verify --update ../fork-1

#########################################################
# Create a bad fork that does not bump the serial number.
#########################################################
cd ${REGISTRY_DIR}
mkdir fork-4
cd fork-4

# Create new entity metadata.
${OASIS_REGISTRY} entity update \
    --signer.dir ${FIXTURES_DIR}/entity-1 \
    ${FIXTURES_DIR}/entity-1/update-bad.json

# Create some more entities.
${OASIS_REGISTRY} entity update \
    --signer.dir ${FIXTURES_DIR}/entity-2 \
    ${FIXTURES_DIR}/entity-2/metadata.json

# Verify registry integrity (update verification should fail).
${OASIS_REGISTRY} verify
! ${OASIS_REGISTRY} verify --update ../fork-1

# Cleanup if everything went well.
rm -rf ${REGISTRY_DIR}
