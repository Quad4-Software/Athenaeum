# Security Policy

## Reporting a vulnerability

Please report security issues privately through GitHub's private vulnerability
reporting on this repository (Security -> Report a vulnerability), not in a
public issue. Include a description, reproduction steps, and the affected
version or commit. We aim to acknowledge reports within a few days.

## Release supply chain

Stable releases are built and published only by the
`.github/workflows/release.yml` workflow on tag pushes. Nightly builds come
from `.github/workflows/nightly.yml` and master image builds from
`.github/workflows/ci.yml`. Both binaries and the GHCR image carry keyless
Sigstore signatures and GitHub build provenance attestations.

### Verify the container image

The image is pushed to `ghcr.io/quad4-software/athenaeum` with a cosign
keyless signature, a build provenance attestation, and an SBOM attestation.

```sh
# Signature (stable release)
cosign verify ghcr.io/quad4-software/athenaeum:latest \
  --certificate-identity-regexp '^https://github.com/Quad4-Software/Athenaeum/.github/workflows/release\.yml@.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com'

# Provenance and SBOM attestations
gh attestation verify oci://ghcr.io/quad4-software/athenaeum:latest \
  --repo Quad4-Software/Athenaeum
```

For `:nightly` and `:sha-<short>` tags built by nightly.yml, replace the
workflow path in the identity regexp with `nightly\.yml`; for the `:latest`
master image built by ci.yml use `ci\.yml`.

### Verify release binaries

Each release binary ships a `<name>.sigstore.json` cosign bundle, plus a
signed `SHA256SUMS.txt` covering all binaries. Attestations and binaries are
published from `release.yml` (stable) or `nightly.yml` (nightly).

```sh
sha256sum -c SHA256SUMS.txt --ignore-missing

cosign verify-blob --bundle SHA256SUMS.txt.sigstore.json \
  --certificate-identity-regexp '^https://github.com/Quad4-Software/Athenaeum/.github/workflows/release\.yml@.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  SHA256SUMS.txt

gh attestation verify athenaeum-vX.Y.Z-linux-amd64 \
  --repo Quad4-Software/Athenaeum
```

The SBOM (`athenaeum-sbom.spdx.json`) lists every Go module and npm package
in SPDX JSON and is covered by a GitHub SBOM attestation.

Build provenance uses GitHub artifact attestations (SLSA build level 2:
keyless signing tied to the release workflow OIDC identity, non-forgeable
provenance stored in the repository attestation store).

## Container hardening

The published image runs as a fixed non-root user (uid/gid 10001), has no
shell login, and only writes under `/data` and `/library`. The shipped
compose files already apply a read-only root filesystem, a `/tmp` tmpfs,
`cap_drop: [ALL]`, and `no-new-privileges`. The app can additionally sandbox
itself with Linux Landlock and seccomp (`ATHENAEUM_SANDBOX=try`, the compose
default); these work without extra capabilities.

Minimal hardened run:

```sh
docker run -d --name athenaeum \
  --read-only --cap-drop ALL --security-opt no-new-privileges:true \
  --tmpfs /tmp:size=16m,mode=1777 \
  -p 8080:8080 \
  -v athenaeum-data:/data -v ./library:/library \
  ghcr.io/quad4-software/athenaeum:latest
```

Rootless Podman works the same way:

```sh
podman run -d --name athenaeum --read-only --cap-drop ALL \
  --security-opt no-new-privileges:true --tmpfs /tmp:size=16m,mode=1777 \
  -p 8080:8080 -v athenaeum-data:/data -v ./library:/library \
  ghcr.io/quad4-software/athenaeum:latest
```

Container HEALTHCHECK calls `athenaeum --healthcheck`, which probes
`/api/health` on loopback and exits non-zero on failure.
