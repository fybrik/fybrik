#
# Copyright 2022 IBM Corp.
# SPDX-License-Identifier: Apache-2.0
#
import requests
from fybrik_python_logging import logger, Error, DataSetID, ForUser

def test_func():
    print("test func")
    logger.trace('authenticating against vault using a JWT token',
        extra={'full_auth_path': str("full_auth_path"),
               DataSetID: "datasetID"})


def get_jwt_from_file(file_name):
    """
    Getting a jwt from a file.
    Typically, an SA token, which would be at: /var/run/secrets/kubernetes.io/serviceaccount/token
    """
    with open(file_name) as f:
        return f.read()

def vault_jwt_auth(jwt, vault_address, vault_path, role, datasetID):
    """Authenticate against Vault using a JWT token (i.e., k8s sa token)"""
    full_auth_path = vault_address + vault_path
    logger.trace('authenticating against vault using a JWT token',
        extra={'full_auth_path': str(full_auth_path),
               DataSetID: datasetID})
    json = {"jwt": jwt, "role": role}
    response = requests.post(full_auth_path, json=json)
    if response.status_code == 200:
        return response.json()
    logger.error("vault authentication failed",
        extra={Error: str(response.status_code) + ': ' + str(response.json()),
               DataSetID: datasetID, ForUser: True})
    return None

def get_raw_secret_from_vault(jwt, secret_path, vault_address, vault_path, role, datasetID):
    """Get a raw secret from vault by providing a valid jwt token"""
    vault_auth_response = vault_jwt_auth(jwt, vault_address, vault_path, role, datasetID)
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
    response = requests.get(secret_full_path, headers={"X-Vault-Token" : client_token})
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

def get_credentials_from_vault(vault_credentials, datasetID):
    jwt_file_path = vault_credentials.get('jwt_file_path', '/var/run/secrets/kubernetes.io/serviceaccount/token')
    jwt = get_jwt_from_file(jwt_file_path)
    vault_address = vault_credentials.get('address', 'https://localhost:8200')
    secret_path = vault_credentials.get('secretPath', '/v1/secret/data/cred')
    vault_auth = vault_credentials.get('authPath', '/v1/auth/kubernetes/login')
    role = vault_credentials.get('role', 'demo')
    logger.trace('getting vault credentials',
        extra={'jwt_file_path': str(jwt_file_path),
               'vault_address': str(vault_address),
               'secret_path': str(secret_path),
               'vault_auth': str(vault_auth),
               'role': str(role),
               DataSetID: datasetID,
               ForUser: True})
    credentials = get_raw_secret_from_vault(jwt, secret_path, vault_address, vault_auth, role, datasetID)
    if not credentials:
        raise ValueError("Vault credentials are missing")
    if 'access_key' in credentials and 'secret_key' in credentials:
        if credentials['access_key'] and credentials['secret_key']:
            return credentials['access_key'], credentials['secret_key']
        else:
            if not credentials['access_key']:
                logger.error("'access_key' must be non-empty",
                             extra={DataSetID: datasetID, ForUser: True})
            if not credentials['secret_key']:
                logger.error("'secret_key' must be non-empty",
                             extra={DataSetID: datasetID, ForUser: True})
    logger.error("Expected both 'access_key' and 'secret_key' fields in vault secret",
                 extra={DataSetID: datasetID, ForUser: True})
    raise ValueError("Vault credentials are missing")