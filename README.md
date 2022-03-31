# dologen

A small CLI tool to generate a Docker `config.json` with registry credentials. Ideal for Kubernetes secrets.

## Usage

**Note:** `username`, `password` or `password-file` and `server` are required!

```
  -b, --base64                 output result base64 encoded
  -p, --password string        password for docker registry
  -f, --password-file string   password file for docker registry
  -s, --server string          docker registry server
  -u, --username string        username for docker registry
  -v, --version                Print the current version and exit
```
