Generate report, of Djobi®© pipeline runs.

## Requirement

* go
* docker

## Usage

### Env. variables

* ``TINTIN_PIPELINES_URLS`` URL to pipelines definitions (git)
* ``TINTIN_PIPELINES_PATH`` relative path to pipelines definitions
* ``HTML_TEMPLATE`` the HTML template to serve
* ``METRICS_LOG_API_URL`` elasticsearch URL to get job details
* ``FRONT_URLS_PATH`` YAML file with magic links
* ``LOG_LEVEL``

### CLI

```
./tintin server
```

## Project workflow

* https://pre-commit.com/
* https://github.com/commitizen/cz-cli

