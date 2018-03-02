from setuptools import setup, find_packages
import os
import io

# long_description = io.open(
    # os.path.join(os.path.dirname(__file__), 'README.rst'), encoding='utf-8').read()

setup(
    name="dialoguss",
    version="0.1.0",
    description="CLI tool for testing HTTP based USSD applications",
    # long_description=long_description,

    # The project URL.
    url='https://github.com/nndi-oss/dialoguss',

    # Author details
    author='Zikani Nyirenda Mwase',
    author_email='zikani@nndi-tech.com',

    # Choose your license
    license='MIT',

    classifiers=[
         'Development Status :: 5 - Production/Stable',
         'Intended Audience :: Developers',
         'Natural Language :: English',
         'License :: OSI Approved :: MIT License',
         'Programming Language :: Python',
         'Programming Language :: Python :: 3.5',
    ],
    entry_points={
        'console_scripts': [ 'dialoguss=dialoguss.core:main' ]
    },
    test_suite="test_dialoguss.py",
    packages=find_packages(),
    include_package_data = True, # include files listed in MANIFEST.in
    install_requires=[
        'requests>=2.0.0', 'pyyaml', 'Flask'
    ],
)
