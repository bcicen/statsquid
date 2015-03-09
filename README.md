# Dockerstats

NOTE: This project is a work in progress. When it is suitable for production use, this message will be removed and a release will be made

dockerstats is a python module for aggregating containers stats across multiple docker hosts 

# Install

# Usage

```python
from dockerstats import DockerStats

ds = DockerStats(config_file='config.yaml') #see config-sample.yaml for sample config file
ds.get_stat()
```

```
{"blkio_stats": {"io_service_time_recursive": [], "sectors_recursive": []...
```

# Documentation
