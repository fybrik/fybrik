# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

"""
The server of the secret-provider app.
"""
import os
import logging
import hcl
from flask import Flask
from flask import Response
from flask import request
from flask import json
import util
import argparse
import sys

app = Flask(__name__)

def configure_app(config_location, logging_level):
    """
    Read load configuration into app.config. Use config_location variable to find configuration.
    If config_location is the empty string, loads hard-coded configuration. logging_level determines logging level.
    """
    
    if logging_level == "info":
        logging.basicConfig(level=logging.INFO)
    elif logging_level == "debug":
        logging.basicConfig(level=logging.DEBUG)
    else:
        sys.exit("unsupported logging level")
    logging.info("Logging level: %s", str(logging_level))
    if config_location != "":
        logging.info("Reading configurations from file: %s", config_location)
        with open(config_location, 'r') as conf_file:
            conf = hcl.load(conf_file)
            # example: "http://vault.vault.svc.cluster.local:8200"
            app.config["vault_address"] = conf["vault_address"]
            logging.info("vault_address: %s", str(app.config["vault_address"]))

            # example: "/v1/auth/kubernetes/login"
            app.config["vault_path"] = conf["vault_path"]
            logging.info("vault_path: %s", str(app.config["vault_path"]))

             # example: "https://iam.cloud.ibm.com/identity/token"
            app.config["iam_endpoint"] = conf["iam_endpoint"]
            logging.info("iam_endpoint: %s", str(app.config["iam_endpoint"]))

            # example: "/var/run/secrets/kubernetes.io/serviceaccount/token"
            app.config["jwt_location"] = conf["jwt_location"]
            logging.info("jwt_location: %s", str(app.config["jwt_location"]))
    else:
        print("Using hard coded configurations")
        log_message = """Using hard coded configurations:
        vault_address = {}
        vault_path = {}
        iam_endpoint = {}
        jwt_location = {}""".format("http://vault.vault.svc.cluster.local:8200",\
                                 "/v1/auth/kubernetes/login", \
                                 "https://iam.cloud.ibm.com/identity/token",\
                                 "/var/run/secrets/kubernetes.io/serviceaccount/token")
        logging.info(log_message)
        app.config["vault_address"] = "http://vault.vault.svc.cluster.local:8200"
        app.config["vault_path"] = "/v1/auth/kubernetes/login"
        app.config["iam_endpoint"] = "https://iam.cloud.ibm.com/identity/token"
        app.config["jwt_location"] = "/var/run/secrets/kubernetes.io/serviceaccount/token"

@app.route('/get-secret')
def get_secret():
    """
    The api endpoint for retrieving secret from the kms.
    http parameters:
    - role: a role to assume when authenticating against vault
    - secret_name: the location of the secret in vault
    - jwt (optional): a jwt token, to use as the identity of the caller
    """
    logging.info("in get-secret")
    logging.debug(str(request.args))
    if "jwt" in request.args:
        jwt = request.args["jwt"]
    else:
        jwt = util.get_jwt_from_file(app.config["jwt_location"])
    if "secret_name" in request.args and "role" in request.args:
        secret_name = request.args["secret_name"]
        role = request.args["role"]
    else:
        logging.info("http parameters are missing")
        resp = Response(status=400)
        return resp
    logging.debug("jwt = %s", str(jwt))
    logging.debug("secret_name = %s", str(secret_name))
    logging.debug("role = %s", str(role))
    raw_secret = util.get_raw_secret_from_vault(jwt, secret_name, app.config["vault_address"], app.config["vault_path"], role)
    if raw_secret is None:
        logging.debug("Didn't get a secret from the vault")
        resp = Response(status=400)
        return resp
    logging.debug("raw_secret: %s", str(raw_secret))
    response = app.response_class(
        response=json.dumps(raw_secret),
        status=200,
        mimetype='application/json'
    )
    logging.debug("created response: %s", str(response))

    return response

@app.route('/get-iam-token')
def get_iam_token():
    """
    The api endpoint for retrieving IAM token in exchange for API keys (which are in vault).
    http parameters:
    - role: a role to assume when authenticating against vault
    - secret_name: the location of the secret in vault
    - jwt (optional): a jwt token, to use as the identity of the caller
    """
    logging.debug(str(request.args))
    if "jwt" in request.args:
        jwt = request.args["jwt"]
    else:
        jwt = util.get_jwt_from_file(app.config["jwt_location"])
    if "secret_name" in request.args and "role" in request.args:
        secret_name = request.args["secret_name"]
        role = request.args["role"]
    else:
        resp = Response(status=400)
        return resp
    logging.debug("jwt = %s", str(jwt))
    logging.debug("secret_name = %s", str(secret_name))
    logging.debug("role = %s", str(role))
    iam_access_token = util.exchange_jwt_to_iam_token(jwt, app.config["iam_endpoint"], \
    secret_name, app.config["vault_address"], app.config["vault_path"], role)
    if iam_access_token is None:
        resp = Response(status=401)
        return resp
    logging.debug("iam_access_token = %s", str(iam_access_token))
    return iam_access_token

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument("-c", "--config", type=str, default="", help="Configuration location. Leave empty for the hard-coded development configuration.")
    parser.add_argument("-l", "--logging", type=str, default="info", help="logging level", choices=["info", "debug"])
    args = parser.parse_args()
    configure_app(args.config, args.logging)
    app.run(host='0.0.0.0', port=5555)
