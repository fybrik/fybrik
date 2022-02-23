from setuptools import setup
import os

setup(name='fybrik_python_logging',
      #version=os.environ.get('RELEASE', '0.1'),
      version='0.0.1',
      description='Python Logging Package for Fybrik Components',
      license='Apache License, Version 2.0',
      packages=['.'],
      install_requires=[
          'JSON-log-formatter==0.5.0',
      ],
)
