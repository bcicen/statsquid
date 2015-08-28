from setuptools import setup

exec(open('statsquid/version.py').read())

setup(name='statsquid',
      version=version,
      packages=['statsquid'],
      description='Docker container stats aggregator',
      author='Bradley Cicenas',
      author_email='bradley.cicenas@gmail.com',
      url='https://github.com/bcicen/statsquid',
      install_requires=['docker-py >= 1.0.0','urllib3 >= 1.8'],
      license='http://opensource.org/licenses/MIT',
      classifiers=(
          'Intended Audience :: Developers',
          'License :: OSI Approved :: MIT License ',
          'Natural Language :: English',
          'Programming Language :: Python',
          'Programming Language :: Python :: 2.6',
          'Programming Language :: Python :: 2.7',
      ),
      keywords='docker stats docker-py',
      entry_points = {
        'console_scripts' : ['statsquid = statsquid.cli:main']
      }
)
