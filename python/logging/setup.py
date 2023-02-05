from setuptools import setup
import os

setup(name='fybrik_python_logging',
      version=os.environ.get('FYBRIK_PYTHON_LOGGING_VERSION', '0.1.0'),
      description='Python Logging Package for Fybrik Components',
      long_description='## Python logging Package for Fybrik Components',
      long_description_content_type='text/markdown',
      license='Apache License, Version 2.0',
      author='FybrikUser',
      author_email='FybrikUser@ibm.com',
      url='https://github.com/fybrik/fybrik/tree/master/python/logging',
      packages=['fybrik_python_logging'],
      install_requires=[
          'JSON-log-formatter==0.5.0',
      ],
)
