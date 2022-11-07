from setuptools import setup
import os

setup(name='fybrik_python_tls',
      version=os.environ.get('FYBRIK_PYTHON_TLS_VERSION', '0.1.0'),
      description='Python TLS Package for Fybrik Components',
      long_description='## Python TLS Package for Fybrik Components',
      long_description_content_type='text/markdown',
      license='Apache License, Version 2.0',
      author='FybrikUser',
      author_email='FybrikUser@il.ibm.com',
      url='https://github.com/fybrik/fybrik/tree/master/python/tls',
      packages=['fybrik_python_tls'],
      install_requires=[
          'fybrik_python_logging',
          'requests'
      ],
)
