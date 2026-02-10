# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| v1.0.x  | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in kubectl-analyze-images, please report it responsibly.

**Do not open a public GitHub issue for security vulnerabilities.**

Instead, please open a GitHub issue with the **"security"** label and include:

1. A description of the vulnerability
2. Steps to reproduce the issue
3. Any relevant logs or error output
4. Your assessment of the severity

We will acknowledge your report within **7 days** and provide a timeline for a fix.

## Scope

kubectl-analyze-images is a **read-only** kubectl plugin with a limited attack surface:

- **Does not modify cluster state** -- it only reads node and pod information
- **Does not store credentials** -- it relies on the standard kubeconfig for authentication
- **Does not make external network calls** -- it only communicates with the Kubernetes API server configured in your kubeconfig
- **Does not write to disk** -- output is sent to stdout only

The primary security considerations are:

- **RBAC permissions**: The plugin requires `list` permissions for nodes and pods. Ensure these are granted through appropriately scoped roles.
- **Output handling**: If piping JSON output to other tools, standard input sanitization practices apply.

## Disclosure Policy

We follow a coordinated disclosure process:

1. Reporter submits vulnerability
2. We acknowledge within 7 days
3. We investigate and develop a fix
4. We release a patched version
5. We publicly disclose the vulnerability after the fix is available
