# :wrench: LINSTOR Gateway Configuration File

The LINSTOR Gateway daemon can be configured using a [toml](https://toml.io) file.

## Location

By default, the application will look for the configuration file in `/etc/linstor-gateway/linstor-gateway.toml`. This
can be overridden with the `--config` command line flag.

## Options

### LINSTOR

| Key           | Default Value | Description                                                                                                                                                                                                                                              |
|---------------|---------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `linstor.controllers` | `[]`          | A list of LINSTOR controllers to try.<br>Each of these IP addresses or hostnames is probed for a LINSTOR controller; the first one that sends a valid response is used. <br> This should include all nodes in the cluster that could potentially host the LINSTOR controller. <br>If this list is empty, `localhost:3370` will be used as the LINSTOR controller. |

## Example

```toml
[linstor]
controllers = ["10.10.1.1", "10.10.1.2", "10.10.1.3"]
```