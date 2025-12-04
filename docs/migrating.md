# Migrating LINSTOR Gateway

## From v1 to v2

In version 2.x, the default REST API port was changed from `8080` to `8337`.

If you are upgrading from LINSTOR Gateway v1.x to v2.x, please ensure that you update any configurations,
firewall rules, or scripts that reference the old port `8080` to the new default port `8337`.

The CLI and Go client have been updated to use the new default port, so no action is necessary if you
don't use the API directly.