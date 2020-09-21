# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

"""
Util module for the Secret-Provider.
"""

import logging
import os
import requests
import sys

def get_jwt_from_env():
    """Getting a JWT token from environment variable, used for development"""
    logging.info("Getting a JWT from environment variable, used for development")
    jwt = os.getenv('JWT')
    logging.debug("jwt: %s", str(jwt))
    return jwt

def get_jwt_from_file(file_name):
    """
    Getting a jwt from a file.
    Typically, an SA token, which would be at: /var/run/secrets/kubernetes.io/serviceaccount/token
    """
    with open(file_name) as f:
        return f.read()

def vault_jwt_auth(jwt, vault_address, vault_path, role):
    """Authenticate against Vault using a JWT token (i.e., k8s sa token)"""
    full_auth_path = vault_address + vault_path
    logging.debug("full_auth_path = %s", str(full_auth_path))
    json = {"jwt": jwt, "role": role}
    response = requests.post(full_auth_path, json=json)
    if response.status_code == 200:
        return response.json()
    return None

def get_raw_secret_from_vault(jwt, secret_path, vault_address, vault_path, role):
    """Get a raw secret from vault by providing a valid jwt token"""
    vault_auth_response = vault_jwt_auth(jwt, vault_address, vault_path, role)
    if vault_auth_response is None:
        return None
    client_token = vault_auth_response["auth"]["client_token"]
    logging.debug("client_token: %s", str(client_token))
    secret_full_path = vault_address + secret_path
    logging.debug("secret_full_path = %s", str(secret_full_path))
    response = requests.get(secret_full_path, headers={"X-Vault-Token" : client_token})
    logging.debug("response: %s", str(response.json()))
    if response.status_code == 200:
        return response.json()['data']
    return None

def get_iam_access_token(api_key, iam_endpoint):
    """Exchange API-KEY with IAM token"""
    headers = {
        'Content-Type': 'application/x-www-form-urlencoded',
        'Accept': 'application/json'
        }
    data = {
        'grant_type': 'urn:ibm:params:oauth:grant-type:apikey',
        'apikey': api_key
    }
    response = requests.post(iam_endpoint, data=data, headers=headers)
    logging.debug("iam-response = %s", str(response.json()))
    if response.status_code == 200:
        return response.json()['access_token']
    return None

def exchange_jwt_to_iam_token(jwt, iam_endpoint="https://iam.cloud.ibm.com/identity/token", \
    secret_path="/v1/kv/hello", vault_address="http://127.0.0.1:8200", \
    vault_path="/v1/auth/kubernetes/login", role="demo"):
    """
    Get an IAM token.
    The flow goes as follows:
    - Authenticate against Vault (using the provided jwt token) and get an API-KEY from it
    - Use the API key to get an IAM token
    """
    raw_secret = get_raw_secret_from_vault(jwt, secret_path, vault_address, vault_path, role)
    if raw_secret is None:
        return None
    api_key = raw_secret["api_key"]
    logging.debug("api-key: %s", str(api_key))
    if api_key is None:
        return None
    access_token = get_iam_access_token(api_key, iam_endpoint)
    return access_token
