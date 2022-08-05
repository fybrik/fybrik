# Support Transport Layer Security (TLS) for Communication with Connectors

## Requirements
- TLS communication is an optional feature, if it is not turned ON, the regular TCP/HTTP communication should be available
- TLS can be extended to mutual TLS
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
- In order to allow automatic certificates renew or revoke, [cert-manager](https://cert-manager.io/) will be used to 
  manage the certificates.
- Client (Fybrik manager) and servers (connectors) configuration can be set independently, if there is no match 
the connection will fail. 
- If a server is configured to use TLS it will not support unencrypted (e.g. HTTP) communications. 

## Implementation Design
- Certificates, keys and CA certificates will be stored in cert-manager and mounted to the relevant Pods as 
certificate directories.
  - Certificate and keys will be stored in the **"tls-cert"** cert-manager mounted directory. The certificate file is
   **"tls.crt"** (in goLang); the key file is "tls.key" (in goLang)
  - CA certificates will be stored in the **"tls-cacert"** cert-manager mounted directory in files with extension **".crt"**
- Other variables will be provided as Pod's environment variables: 
  - **"USE_TLS"** - use or not TLS, the server settings only
  - **"USE_MTLS"** - should the server require mutual TLS, the server settings only
  - **"MIN_TLS_VERSION"** - minimal supported TLS version (exact values depend on the programming language)
