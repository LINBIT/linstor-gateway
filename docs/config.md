# :wrench: LINSTOR Gateway Configuration File

The LINSTOR Gateway daemon can be configured using a [toml](https://toml.io) file.

## Location

By default, the application will look for the configuration file in `/etc/linstor-gateway/linstor-gateway.toml`. This
can be overridden with the `--config` command line flag.

## Options

### LINSTOR

| Key                   | Default Value | Description                                                                                                                                                                                                                                                                                                                                                       |
| --------------------- | ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `linstor.controllers` | `[]`          | A list of LINSTOR controllers to try.<br>Each of these IP addresses or hostnames is probed for a LINSTOR controller; the first one that sends a valid response is used. <br> This should include all nodes in the cluster that could potentially host the LINSTOR controller. <br>If this list is empty, `localhost:3370` will be used as the LINSTOR controller. |

### Server

| Key                           | Default Value | Description                                                                                                                                                                                                                                                                                                                                                       |
| ----------------------------- | ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `server.cors_allowed_origins` | `[]`          | Additional allowed origins for CORS.<br>If `linstor.controllers` is set, origins are automatically generated for each controller's 3370 port with both http and https (e.g., `["http://10.10.1.1:3370", "https://10.10.1.1:3370"]`).<br>These user-defined origins are **merged** with the auto-generated ones.<br>If both are empty, **no origins are allowed**. |

## Example

```toml
[linstor]
controllers = ["10.10.1.1", "10.10.1.2", "10.10.1.3"]

[server]
# Optional: add extra CORS origins (merged with auto-generated controller origins)
# cors_allowed_origins = ["https://example.com"]
```
