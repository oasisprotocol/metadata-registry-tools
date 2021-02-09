# Oasis Metadata Registry Tools

[![CI test status][github-ci-tests-badge]][github-ci-tests-link]
[![CI lint status][github-ci-lint-badge]][github-ci-lint-link]

<!-- markdownlint-disable line-length -->
[github-ci-tests-badge]: https://github.com/oasisprotocol/oasis-core-rosetta-gateway/workflows/ci-tests/badge.svg
[github-ci-tests-link]: https://github.com/oasisprotocol/oasis-core-rosetta-gateway/actions?query=workflow:ci-tests+branch:master
[github-ci-lint-badge]: https://github.com/oasisprotocol/oasis-core-rosetta-gateway/workflows/ci-lint/badge.svg
[github-ci-lint-link]: https://github.com/oasisprotocol/oasis-core-rosetta-gateway/actions?query=workflow:ci-lint+branch:master
<!-- markdownlint-enable line-length -->

This repository contains tools for working with the [Oasis Metadata Registry].

[Oasis Metadata Registry]: https://github.com/oasisprotocol/metadata-registry

## Building

To build the `oasis-registry` tool, run:

```sh
make build
```

## Usage

_NOTE: Currently, you will need to build the `oasis-registry` tool yourself._

**NOTE: Support for signing entity metadata statements with the Ledger-based
signer is available in [Oasis app 1.9.0+ releases] which will soon be available
via Ledger Live's Manager.**

To sign an entity metadata statement, e.g.

```json
{
  "v": 1,
  "serial": 1,
  "name": "My entity name",
  "url": "https://my.entity/url",
  "email": "my@entity.org",
  "keybase": "my_keybase_handle",
  "twitter": "my_twitter_handle"
}
```

save it as a JSON file, e.g. `entity-metadata.json`, and run:

```sh
./oasis-registry/oasis-registry entity update \
  <SIGNER-FLAGS> \
  entity-metadata.json
```

where `<SIGNER-FLAGS>` are replaced by the appropriate signer CLI flags for your
signer (e.g. Ledger-based signer, File-based signer).

For more details, run:

```sh
./oasis-registry/oasis-registry entity update --help
```

_NOTE: The same signer flags as used by the Oasis Node CLI are supported.
See [Oasis CLI Tools' documentation on Signer Flags][oasis-cli-flags] for more
details._

The `oasis-registry entity update` command will output a preview of the entity
metadata statement you are about to sign:

```text
You are about to sign the following entity metadata descriptor:
  Version: 1
  Serial:  1
  Name:    My entity name
  URL:     https://my.entity/url
  Email:   my@entity.org
  Keybase: my_keybase_handle
  Twitter: my_twitter_handle
```

and ask you for confirmation.

It will store the signed entity metadata statement to the
`registry/entity/<HEX-ENCODED-ENTITY-PUBLIC-KEY>.json` file, where
`<HEX-ENCODED-ENTITY-PUBLIC-KEY>` corresponds to your hex-encoded entity's
public key, e.g.
`918cfe60b903e9d2c3003eaa78997f4fd95d66597f20cea8693e447b6637604c.json`.

<!-- markdownlint-disable line-length -->
[oasis-cli-flags]:
  https://docs.oasis.dev/general/manage-tokens/oasis-cli-tools/setup#signer-flags
[Oasis app 1.9.0+ releases]: https://github.com/Zondax/ledger-oasis/releases
<!-- markdownlint-enable line-length -->

### Contributing Entity Metadata Statement to Production Oasis Metadata Registry

See the [Contributing New Statements guide][contrib-guide] at the
[Oasis Metadata Registry]'s web site.

[contrib-guide]:
  https://github.com/oasisprotocol/metadata-registry#contributing-new-statements

## Development

### Examples

For some examples of using this Go library, check the [`examples/`] directory.

To build all examples, run:

```sh
make build-examples
```

To run the `lookup` example that lists the entity metadata statements in the
production Oasis Metadata Registry, run:

```sh
./examples/lookup/lookup
```

It should give an output similar to:

```text
[ms7M1v8HfItCnNNJ0tfE/PsYQsmeD+XpfGF1v0zR2Xo=]
  Name:    Everstake
  URL:     https://everstake.one
  Email:   inbox@everstake.one
  Keybase: everstake
  Twitter: everstake_pool

[gb8SHLeDc69Elk7OTfqhtVgE2sqxrBCDQI84xKR+Bjg=]
  Name:    Bi23 Labs
  URL:     https://bi23.com
  Email:   support@bi23.com
  Keybase: sunxmldapp
  Twitter: bi23com

... output trimmed ...
```

[`examples/`]: examples/

### Test Vectors

To generate the entity metadata test vectors, run:

```sh
make gen_vectors
```

### Tests

To run all tests, run:

```sh
make test
```

This will run all Make's test targets which include Go unit tests and CLI tests.

_NOTE: CLI tests with Ledger signer will be skipped unless the
`LEDGER_SIGNER_PATH` is set and exported._

#### Tests with Ledger-based signer

To run CLI tests with Ledger-based signer, you need to follow these steps:

1. Download the latest [Oasis Core Ledger] release from
   <https://github.com/oasisprotocol/oasis-core-ledger/releases>.

2. Extract the `oasis_core_ledger_<VERSION>_<OS>_amd64.tar.gz` tarball.

3. Set `LEDGER_SIGNER_PATH` environment variable to the path of the extracted
   `ledger-signer` binary and export it, e.g.:

   ```sh
   export LEDGER_SIGNER_PATH="/path/to/oasis_core_ledger_1.2.0_linux_amd64/ledger-signer"
   ```

4. Connect your Ledger device and make sure the Oasis app is open.

5. Run tests with:

   ```sh
   make test-cli-ledger
   ```

[Oasis Core Ledger]: https://docs.oasis.dev/oasis-core-ledger/
