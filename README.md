# AI Outreach Generator

## Configuration

1. Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

2. Update the `.env` file with your API keys:

- Get your Airtable Personal Access Token from: https://airtable.com/create/tokens
- Get your Anthropic API Key from: https://console.anthropic.com/
- Get your Airtable Base ID from the URL when viewing your base: `airtable.com/BASE_ID/...`
- Set your Airtable Table Name (the name of the table containing your contacts)

3. Make sure your Airtable Personal Access Token has the following scopes:

- `data.records:read` - to read records
- `schema.bases:read` - to read base schema
- Access to the specific base you want to use

## Running the Application

```bash
# Install dependencies
make dev-deps

# Run in development mode
make run
```

# Build for current platform

```sh
make build
```

# Build for all platforms

```sh
make build-all
```

# Clean up binaries

```sh
make clean
```
