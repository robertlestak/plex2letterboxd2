# plex2letterboxd2

Fully automated sync of Plex watched movies to Letterboxd with browser automation for hands-free imports.

## Features

- Export Plex watched movies with IMDb IDs, ratings, and watch dates
- Browser automation (RPA) to automatically upload and import to Letterboxd
- Kubernetes CronJob support for scheduled syncs
- Docker support for easy deployment

## Acknowledgments

This project conceptually builds upon [mtimkovich/plex2letterboxd](https://github.com/mtimkovich/plex2letterboxd) by adding browser automation to fully automate the Plex to Letterboxd sync workflow, eliminating the manual CSV upload step.

## A Note on Stability

Since Letterboxd does not provide an official API for imports, this tool relies on browser automation to interact with the Letterboxd web interface. As a result, changes to the Letterboxd website may impact the tool's functionality. While efforts are made to maintain compatibility, users should be aware that occasional adjustments may be necessary. PRs are welcome to help keep the tool up to date.

## Quick Start

### Local Usage

```bash
export PLEX_URL="http://your-plex-server:32400"
export PLEX_TOKEN="your-plex-token"
export LETTERBOXD_USERNAME="your-letterboxd-username"
export LETTERBOXD_PASSWORD="your-letterboxd-password"

go run cmd/p2l2/p2l2.go
```

Or with flags:

```bash
go run cmd/p2l2/p2l2.go -plex-url="http://your-plex-server:32400" -plex-token="your-token" -output="movies.csv"
```

### Docker

```bash
docker build -t plex2letterboxd2 .
docker run -e PLEX_URL="http://your-server:32400" -e PLEX_TOKEN="your-token" plex2letterboxd2
```

### Kubernetes

```bash
# Create secret with Plex credentials
kubectl create secret generic plex2letterboxd2-secret \
  --from-literal=PLEX_URL="http://plex:32400" \
  --from-literal=PLEX_TOKEN="your-token" \
  --from-literal=LETTERBOXD_USERNAME="your-username" \
  --from-literal=LETTERBOXD_PASSWORD="your-password"

# Install Helm chart
helm install plex2letterboxd2 ./helm/plex2letterboxd2 \
  --set secretName="plex2letterboxd2-secret" \
  --set schedule="0 2 * * *"
```

## Configuration

### Getting Your Plex Token

1. Sign in to Plex Web App
2. Open any media item
3. Click "Get Info" → "View XML"
4. Look for `X-Plex-Token` in the URL

### Helm Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Container image repository | `plex2letterboxd2` |
| `image.tag` | Container image tag | `latest` |
| `schedule` | CronJob schedule | `0 2 * * *` |
| `secretName` | Kubernetes secret name | `plex2letterboxd2-secret` |
| `imagePullSecrets` | Image pull secrets | `[]` |

## Manual Import

If you prefer to manually import the CSV:

1. Run the tool with `-import=false` to generate `letterboxd.csv`
2. Go to https://letterboxd.com/import/
3. Upload the CSV file
4. Review and confirm the import

## License

MIT
