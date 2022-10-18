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

def make_http_request(request_type, url, data=None, headers=None,
                      tls_min_version=None, verify=True, cert=None):
    """ Making http request call.
    ref: https://requests.readthedocs.io/en/latest/user/advanced/#transport-adapters

    Args:
        request_type (string): the http request type. For example: "POST" or "GET"
        url (string): The url.
        data (dictionary, optional): list of tuples, bytes, or file-like
        object to send in the body of the request. Defaults to None.
        headers (dictionary, optional): Dictionary of HTTP Headers to send with the request. Defaults to None.
        tls_min_version (string, optional): the minimum tls version to use. Defaults to None.
        verify (optional): Either a boolean, in which case it controls whether we verify
        the Vault server's TLS certificate, or a string, in which case it must be a path
        to a CA bundle to use. Defaults to ``True``.
        cert (tuple, optional): the module ('cert', 'key') pair.
    Returns:
        requests.Response
    """

    # Set tls properties when the vault server uses tls.
    context = create_ssl_context(tls_min_version)
    s = requests.Session()
    s.mount("https://", SSLContextAdapter(context))
 
    req = requests.Request(request_type, url=url, data=data, headers=headers)
    prep = s.prepare_request(req)
    # Merge environment settings into session
    settings = s.merge_environment_settings(prep.url, {}, None, verify=verify, cert=cert)
    response = s.send(prep, **settings)
    return response

def vault_jwt_auth(jwt, vault_address, vault_path, role, datasetID, tls_min_version=None,
                   verify=True, cert=None):
    """Authenticate against Vault using a JWT token (i.e., k8s sa token)

    Args:
        jwt (string): the service account jwt to use to authenticate against Vault.
        vault_address (string): the Vault address.
        vault_path (string): the path to authenticate against Vault using Kuberenetes auth method.
        For example: `/v1/auth/kubernetes/login`.
        role (string): the name of Vault role.
        datasetID (string): the dataset ID.
        tls_min_version (string, optional): the minimum tls version to use. Defaults to None.
        verify (optional): Either a boolean, in which case it controls whether we verify
        the Vault server's TLS certificate, or a string, in which case it must be a path
        to a CA bundle to use. Defaults to ``True``.
        cert (tuple, optional): the module ('cert', 'key') pair.
    Returns:
        the json-encoded content of a response which contains a Vault token, if any
    """

    full_auth_path = vault_address + vault_path
    st = ' '
    if cert:
        st = st.join(cert)
    logger.trace('authenticating against vault using a JWT token',
        extra={'full_auth_path': str(full_auth_path),
               DataSetID: datasetID, 'verify': verify, 'cert' : st})
    json = {"jwt": jwt, "role": role}
    response = make_http_request("POST", full_auth_path, data=json, tls_min_version=tls_min_version,
                                 verify=verify, cert=cert)
    if response.status_code == 200:
        return response.json()
    logger.error("vault authentication failed",
        extra={Error: str(response.status_code) + ': ' + str(response.json()),
               DataSetID: datasetID, ForUser: True})
    return None

def get_raw_secret_from_vault(jwt, secret_path, vault_address, vault_path, role, datasetID,
                              tls_min_version=None, verify=True, cert=None):
    """Get a raw secret from vault by providing a valid jwt token

    Args:
        jwt (string): the service account jwt to use to authenticate against Vault.
        secret_path (string): Vault path of the secret.
        vault_address (string): the Vault address.
        vault_path (string): the path to authenticate against Vault using Kuberenetes auth method.
        For example: `/v1/auth/kubernetes/login`.
        role (string): the name of Vault role.
        datasetID (string): the dataset ID.
        tls_min_version (string, optional): the minimum tls version to use. Defaults to None.
        verify (optional): Either a boolean, in which case it controls whether we verify
        the Vault server's TLS certificate, or a string, in which case it must be a path
        to a CA bundle to use. Defaults to ``True``.
        cert (tuple, optional): the module ('cert', 'key') pair.

    Returns:
        the json-encoded content of a response which contains the secret, if any
    """
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
    response = make_http_request("GET", secret_full_path, headers=headers,
                                 tls_min_version=tls_min_version, verify=verify, cert=cert)
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
