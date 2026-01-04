# Jellyfin Duplicate Finder - Setup Guide

## Prerequisites

Before you can run this application, you need:

1. **Go 1.21 or later** installed on your system
2. **Jellyfin server** with API access enabled
3. **Jellyfin API key** (from your user settings)
4. **User ID** (from your Jellyfin user profile)

## Installation Steps

### 1. Install Go

If you don't have Go installed:

- **macOS**: `brew install go`
- **Linux**: Follow instructions at https://golang.org/doc/install
- **Windows**: Download installer from https://golang.org/dl/

### 2. Clone this repository

```bash
git clone https://github.com/yourusername/jellyfin-duplicate.git
cd jellyfin-duplicate
```

### 3. Install dependencies

```bash
go mod download
```

### 4. Configure the application

Edit `config.json` with your Jellyfin server details:

```json
{
  "jellyfin_url": "http://your-jellyfin-server:8096",
  "api_key": "your_api_key_here",
  "user_id": "your_user_id_here",
  "server_port": "8080"
}
```

#### How to get your Jellyfin API key:
1. Log in to your Jellyfin web interface
2. Go to your user profile (top right) → Settings
3. Go to the "API Keys" tab
4. Create a new API key or use an existing one

#### How to get your User ID:
1. Log in to your Jellyfin web interface
2. Go to your user profile (top right) → Settings
3. The User ID is shown in the URL: `.../web/index.html#!/userprofile.html?userId=YOUR_USER_ID`

### 5. Build the application

```bash
go build -o jellyfin-duplicate ./cmd/api
```

### 6. Run the application

```bash
./jellyfin-duplicate
```

### 7. Access the web interface

Open your browser and go to: `http://localhost:8080`

## API Endpoints

- **Web Interface**: `http://localhost:8080/`
- **JSON API**: `http://localhost:8080/api/duplicates`

## Troubleshooting

### Common Issues

**1. "Failed to load config" error**
- Make sure `config.json` exists in the same directory as the executable
- Verify the JSON syntax is correct

**2. "Failed to get libraries" error**
- Check your Jellyfin server URL is correct
- Verify your API key is valid
- Ensure your User ID is correct
- Make sure your Jellyfin server is running

**3. No duplicates found**
- This might be correct if you don't have duplicates
- Check that your movies have TMDB/IMDB IDs in Jellyfin
- Verify you have multiple movies with the same IDs

### Debugging

Run with debug logging:
```bash
DEBUG=1 ./jellyfin-duplicate
```

Check the Jellyfin server logs for API access issues.

## Configuration Options

You can customize the similarity threshold by modifying the `findDuplicates()` function in `internal/handlers/handler.go`.

The current threshold is 95% similarity for duplicates. You can adjust this by changing:
```go
isDuplicate := similarity >= 95  // Change 95 to your desired threshold
```

## Running as a Service

To run this as a background service, you can use systemd (Linux) or launchd (macOS).

### systemd example (Linux)

Create `/etc/systemd/system/jellyfin-duplicate.service`:

```ini
[Unit]
Description=Jellyfin Duplicate Finder
After=network.target

[Service]
User=yourusername
WorkingDirectory=/path/to/jellyfin-duplicate
ExecStart=/path/to/jellyfin-duplicate/jellyfin-duplicate
Restart=always

[Install]
WantedBy=multi-user.target
```

Then enable and start:
```bash
sudo systemctl enable jellyfin-duplicate
sudo systemctl start jellyfin-duplicate
```

## Updating

To update to the latest version:

```bash
cd jellyfin-duplicate
git pull origin main
go build -o jellyfin-duplicate ./cmd/api
sudo systemctl restart jellyfin-duplicate  # if running as service
```

## Support

If you encounter issues, please check:
- Your Jellyfin server logs
- The application logs
- That your API key and user ID are correct
- That your Jellyfin server is accessible from the machine running this application