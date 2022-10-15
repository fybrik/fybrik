#
# Copyright 2022 IBM Corp.
# SPDX-License-Identifier: Apache-2.0
#
import requests
from fybrik_python_logging import logger, Error, DataSetID, ForUser
from fybrik_python_tls import SSLContextAdapter, create_ssl_context


def get_jwt_from_file(file_name):
    """
    Getting a jwt from a file.
    Typically, an SA token, which would be at: /var/run/secrets/kubernetes.io/serviceaccount/token
    """
    with open(file_name) as f:
        return f.read()

def make_http_request(request_type, url, tls_min_version, data=None, headers=None, verify=None, cert=None):
    """
    Making http request call.
    ref: https://requests.readthedocs.io/en/latest/user/advanced/#transport-adapters
    """
    # Set properties related to tls when the vault server uses tls.
    context = create_ssl_context(tls_min_version)
    s = requests.Session()
    s.mount("https://", SSLContextAdapter(context))
 
    req = requests.Request(request_type, url=url, data=data, headers=headers)
    prep = s.prepare_request(req)
    response = s.send(prep, verify=verify, cert=cert)
    return response

def vault_jwt_auth(jwt, vault_address, vault_path, role, datasetID, tls_min_version, verify=None, cert=None):
    """Authenticate against Vault using a JWT token (i.e., k8s sa token)"""
    full_auth_path = vault_address + vault_path
    st = ' '
    if cert:
        st = st.join(cert)
    logger.trace('authenticating against vault using a JWT token',
        extra={'full_auth_path': str(full_auth_path),
               DataSetID: datasetID, 'verify': verify, 'cert' : st})
    json = {"jwt": jwt, "role": role}
    response = make_http_request("POST", full_auth_path, tls_min_version, data=json, verify=verify, cert=cert)
    if response.status_code == 200:
        return response.json()
    logger.error("vault authentication failed",
        extra={Error: str(response.status_code) + ': ' + str(response.json()),
               DataSetID: datasetID, ForUser: True})
    return None

def get_raw_secret_from_vault(jwt, secret_path, vault_address, vault_path, role, datasetID, tls_min_version,
                              verify=None, cert=None):
    """Get a raw secret from vault by providing a valid jwt token"""
    st = ' '
    if cert:
        st = st.join(cert)
    logger.trace('getting vault credentials',
        extra={'vault_address': str(vault_address),
               'secret_path': str(secret_path),
               'vault_path': str(vault_path),
               'role': str(role),
               DataSetID: datasetID,
               'verify': verify,
               'cert' : st,
               ForUser: True})
    vault_auth_response = vault_jwt_auth(jwt, vault_address, vault_path, role, datasetID, tls_min_version,
                                         verify, cert)
    if vault_auth_response is None:
        logger.error("Empty vault authorization response",
                     extra={DataSetID: datasetID, ForUser: True})
        return None
    if not "auth" in vault_auth_response or not "client_token" in vault_auth_response["auth"]:
        logger.error("Malformed vault authorization response",
                     extra={DataSetID: datasetID, ForUser: True})
        return None
    client_token = vault_auth_response["auth"]["client_token"]
    secret_full_path = vault_address + secret_path
    headers={"X-Vault-Token" : client_token}
    response = make_http_request("GET", secret_full_path, tls_min_version, headers=headers, verify=verify, cert=cert)
    logger.debug('Response received from vault when accessing credentials: ' + str(response.status_code),
        extra={'credentials_path': str(secret_full_path),
               DataSetID: datasetID, ForUser: True})
    if response.status_code == 200:
        response_json = response.json()
        if 'data' in response_json:
            return response_json['data']
        else:
            logger.error("Malformed secret response. Expected the 'data' field in JSON",
                         extra={DataSetID: datasetID, ForUser: True})
    else:
        logger.error("Error reading credentials from vault",
            extra={Error: str(response.status_code) + ': ' + str(response.json()),
                   DataSetID: datasetID, ForUser: True})
    return None
