from setuptools import setup
import os

setup(name='fybrik_python_logging',
      version=os.environ.get('RELEASE', '0.1'),
      description='Python Logging Package for Fybrik Components',
      license='Apache License, Version 2.0',
      author='Doron Chen',
      author_email='cdoron@il.ibm.com',
      url='https://github.com/fybrik/fybrik/tree/master/pkg/logging/python_logging',
      packages=['fybrik_python_logging'],
      install_requires=[
          'JSON-log-formatter==0.5.0',
      ],
)
