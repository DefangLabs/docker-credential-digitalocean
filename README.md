# docker-credential-digitalocean
Docker Credential Helper for Digital Ocean.

This helper uses the environment variable `DIGITALOCEAN_TOKEN` and invokes the `/v2/registry/docker-credentials` API to get Docker credentials for the DO Container Registry.
