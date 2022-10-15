#
# Copyright 2022 IBM Corp.
# SPDX-License-Identifier: Apache-2.0
#

import ssl
import requests
from fybrik_python_logging import logger

def create_ssl_context(tls_min_version=None):
    """Create a SSLContext object with custome settings."""
    context = ssl.create_default_context()
    if tls_min_version != None:
       logger.debug("set minimum_version")
       context.minimum_version = tls_min_version
    return context
    
class SSLContextAdapter(requests.adapters.HTTPAdapter):
    """A custome built-in HTTP Adapter.
       ref: https://requests.readthedocs.io/en/latest/user/advanced/#transport-adapters
    """
    def __init__(self, ssl_context=None, **kwargs):
         self.ssl_context = ssl_context
         super().__init__(**kwargs)
        
    def init_poolmanager(self, *args, **kwargs):
         kwargs['ssl_context'] = self.ssl_context
         return super(SSLContextAdapter, self).init_poolmanager(*args, **kwargs)
