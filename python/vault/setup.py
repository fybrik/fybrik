from setuptools import setup
import os

setup(name='fybrik_python_vault',
      version=os.environ.get('PYBRIK_PYTHON_VAULT_VERSION', '0.2.0'),
      description='Python Vault Package for Fybrik Components',
      long_description='## Python Vault Package for Fybrik Components',
      long_description_content_type='text/markdown',
      license='Apache License, Version 2.0',
      author='FybrikUser',
      author_email='FybrikUser@il.ibm.com',
      url='https://github.com/fybrik/fybrik/tree/master/python/vault',
      packages=['fybrik_python_vault'],
      install_requires=[
          'fybrik_python_logging',
          'fybrik_python_tls',
          'requests'
      ],
)
