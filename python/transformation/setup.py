from setuptools import setup
import os

setup(name='fybrik_python_transformation',
      version=os.environ.get('FYBRIK_PYTHON_TRANSFORMATION_VERSION', '0.1.0'),
      description='Python Transformation Package for Fybrik',
      long_description='## Python Transformation Package for Fybrik',
      long_description_content_type='text/markdown',
      license='Apache License, Version 2.0',
      author='FybrikUser',
      author_email='FybrikUser@ibm.com',
      url='https://github.com/fybrik/fybrik/tree/master/python/transformation',
      packages=['fybrik_python_transformation'],
      install_requires=[
          "pandas==1.4.2",
          "pyarrow==8.0.0",
      ],
)
