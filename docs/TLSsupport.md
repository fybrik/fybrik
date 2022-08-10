# Support Transport Layer Security (TLS) for Communication with Connectors

## Requirements
- TLS communication is an optional feature, if it is not turned ON, the regular TCP/HTTP communication should be available
- TLS can be extended to mutual TLS
- Client (Fybrik manager) and servers (connectors) configuration can be set independently, if there is no match
  the connection will fail.
  - Client when it starts always loads all provided certificates and keys, after that it tries to connect to the server
  based on server's URL.
  - Server loads the provided certificates only it is configured to use TLS, and loads CA certificates if mutual TLS is ON.
- The system should allow setting required TLS parameters, Certificates, Certificate Authorities (CA).
- If CA are not set, the default CA will be used. Due to close communication between Fybrik manager and connectors, 
  explicit setting CA will **replace** the default goLang or Java CA settings. Optionally, they can be merged.
- Optionally the system will allow setting the minimal supported TLS protocol, possible values are:
  TLS 1.0 ([RFC 2246](https://datatracker.ietf.org/doc/html/rfc2246)), 
  TLS 1.1 ([RFC 4346](https://datatracker.ietf.org/doc/html/rfc4346)), 
  TLS 1.2 ([RFC 5246](https://datatracker.ietf.org/doc/html/rfc5246)) and 
  TLS 1.3 (RFC [8446](https://datatracker.ietf.org/doc/html/rfc8446))
- The current implementation will not allow changing the default cipher suites. (golang allows defining a list of enabled 
TLS 1.0–1.2 cipher suites. TLS 1.3 cipher suites are not configurable.)
- In order to allow automatic certificates renew or revoke, [cert-manager](https://cert-manager.io/) can be used to 
  manage the certificates.
- If a server is configured to use TLS it will not support unencrypted (e.g. HTTP) communications.
- The current implementation will not automatically reload renewed certificates and requires the restart of relevant Pods.
  In the future, we need to add a mechanism to do it automatically without restart of Pods.

## Implementation Design
- Certificates, keys and CA certificates will be stored in Kubernetes secrets and mounted to the relevant Pods as 
certificate files.
- Certificate and keys will be stored in the **"tls-cert"** secret mounted directory. The certificate file is
 **"tls.crt"** (in goLang); the key file is "tls.key" (in goLang)
  - The secret which contains the CA certificates is mounted to a local directory **tls-cacert** in the container. Only 
  files with extension ".crt" are considered.
- Other variables will be provided as Pod's environment variables: 
  - **"USE_TLS"** - use or not TLS, the server settings only
  - **"USE_MTLS"** - should the server require mutual TLS, the server settings only
  - **"MIN_TLS_VERSION"** - minimal supported TLS version, possible values are `TLS-1.0`, `TLS-1.1`, `TLS-1.2` and 
`TLS-1.3`. If value is not set, the default system value is used.
