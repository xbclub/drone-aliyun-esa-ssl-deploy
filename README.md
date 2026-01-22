# Aliyun ESA SSL Upload

Upload SSL certificates to Aliyun ESA using the ESA API.

## Environment Variables

- `PLUGIN_PUBLIC_KEY`: Path to the certificate public key file (PEM).
- `PLUGIN_PRIVATE_KEY`: Path to the certificate private key file (PEM).
- `PLUGIN_ACCESS_KEY_ID`: Aliyun access key ID. Optional if using the default credential chain.
- `PLUGIN_ACCESS_KEY_SECRET`: Aliyun access key secret. Optional if using the default credential chain.
- `PLUGIN_ENDPOINT`: ESA endpoint. Default: `esa.cn-hangzhou.aliyuncs.com`.
- `PLUGIN_CERT_TYPE`: Certificate type. Default: `upload`.

## Usage

```bash
export PLUGIN_PUBLIC_KEY=/path/to/cert.pem
export PLUGIN_PRIVATE_KEY=/path/to/key.pem
export PLUGIN_ACCESS_KEY_ID=your_ak
export PLUGIN_ACCESS_KEY_SECRET=your_sk

go run .
```

## Build

```bash
go build -o esa-ssl-upload .
```

## Example Code

The original ESA SDK example is kept in `example.go` and guarded by a build tag.

```bash
go build -tags=example ./...
```
