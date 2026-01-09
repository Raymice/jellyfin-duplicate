# Jellyfin Duplicate Finder

**Intelligent duplicate movie detection with multi-user play status analysis**

A Go application that helps you safely identify and remove duplicate movies from your Jellyfin server while preserving user watch history. The application analyzes movie metadata, file paths, and play status across all users to provide intelligent recommendations for safe duplicate removal.

## Features

- Fetches all movies from your Jellyfin server
- Identifies potential duplicates based on movie name and production year
- Uses Levenshtein distance to analyze file paths for similarity
- Classifies duplicates vs mismatches (95% similarity threshold)
- **Multi-user play status analysis** - checks if users have seen different versions
- **Play status discrepancy detection** - identifies when users have seen one version but not the other
- **Safe deletion guidance** - only recommends deletion when play status is identical
- **Play status synchronization** - allows marking movies as seen for specific users
- **Movie deletion** - permanently remove duplicate movies from Jellyfin
- Provides both HTML web interface and JSON API endpoints

## Installation

### Native Installation

1. Clone this repository
2. Install dependencies:

   ```bash
   go mod download
   ```

3. Build the application:

   ```bash
   go build -o jellyfin-duplicate .
   ```

### Docker Installation

#### Using Docker Compose (Recommended)

1. Start the container:

   ```bash
   docker-compose up -d
   ```

#### Using Docker Directly

1. Build the Docker image:

   - amd64

      ```bash
      docker build --platform linux/amd64 -t jellyfin-duplicate .
      ```

   - arm64

      ```bash
      docker build --platform linux/arm64 -t jellyfin-duplicate .
      ```

2. Run the container:
   ```bash
   docker run -d \
     -p 8080:8080 \
     -e JELLYFIN_URL="your-jellyfin-url" \
     -e JELLYFIN_API_KEY="your-api-key" \
     -e JELLYFIN_USER_ID="your-user-id" \
     --name jellyfin-duplicate \
     jellyfin-duplicate
   ```

3. (Optional) Mount a custom configuration file:

   ```bash
   docker run -d \
     -p 8080:8080 \
     -v /path/to/your/config.json:/app/config.prod.json \
     -e JELLYFIN_URL="your-jellyfin-url" \
     -e JELLYFIN_API_KEY="your-api-key" \
     -e JELLYFIN_USER_ID="your-user-id" \
     --name jellyfin-duplicate \
     jellyfin-duplicate
   ```

## Configuration

### Native Configuration

Edit the `cmd/api/main.go` file to configure:
- Jellyfin server URL
- API key
- User ID (for library access)

### Docker Configuration

The application can be configured using environment variables:

- `JELLYFIN_URL`: URL of your Jellyfin server (required)
- `JELLYFIN_API_KEY`: Jellyfin API key (required)
- `JELLYFIN_USER_ID`: Jellyfin user ID (required)
- `ENVIRONMENT`: `development` or `production` (default: `production`)

You can also mount a custom `config.prod.json` file to `/app/config.prod.json` in the container.

## Usage

### Native Usage

Run the application:
```bash
./jellyfin-duplicate
```

Access the web interface at: `http://localhost:8080`

**Available endpoints:**
- Web interface: `http://localhost:8080` - Interactive duplicate analysis
- Analysis page: `http://localhost:8080/analysis` - Detailed results with play status
- JSON API: `http://localhost:8080/api/duplicates` - Machine-readable results
- Mark as seen: `http://localhost:8080/api/mark-as-seen?movieId=XXX&userId=YYY`
- Delete movie: `http://localhost:8080/api/delete-movie?movieId=XXX`

### Docker Usage

After starting the container, access the web interface at: `http://localhost:8080`

**Available endpoints:**
- Web interface: `http://localhost:8080` - Interactive duplicate analysis
- Analysis page: `http://localhost:8080/analysis` - Detailed results with play status
- JSON API: `http://localhost:8080/api/duplicates` - Machine-readable results
- Mark as seen: `http://localhost:8080/api/mark-as-seen?movieId=XXX&userId=YYY`
- Delete movie: `http://localhost:8080/api/delete-movie?movieId=XXX`

To stop the container:
```bash
docker stop jellyfin-duplicate
```

To restart the container:
```bash
docker start jellyfin-duplicate
```

## How It Works

1. The application fetches all movies from your Jellyfin libraries
2. It groups movies by their name and production year
3. For each group with multiple movies, it compares file paths using Levenshtein distance
4. If path similarity is ≥95%, it's classified as a **potential duplicate**
5. If path similarity is <95%, it's classified as a **potential mismatch**
6. **Play status analysis**: For each duplicate pair, the application checks if users have seen both versions
7. **Safe deletion guidance**: Only shows delete buttons when both versions have identical play status
8. **Discrepancy detection**: Identifies when users have seen one version but not the other

## Why Play Status Matters

**⚠️ IMPORTANT: Always check play status before deleting duplicates!**

The application analyzes play status across all users to prevent data loss:

- **✅ Safe to delete**: When both versions have identical play status (same users have seen both)
- **❌ Not safe to delete**: When users have seen one version but not the other
- **🔄 Play status discrepancies**: The application helps you synchronize play status before deletion

### Example scenarios:

1. **Identical play status**: Both versions have been seen by the same users → Safe to delete one
2. **Different play status**: User A saw version 1, User B saw version 2 → NOT safe to delete
3. **Partial overlap**: Some users saw both, others saw only one → Requires synchronization

The application provides tools to mark movies as seen for specific users, ensuring you don't lose watch history when removing duplicates.

## Dependencies

- [Gin](https://github.com/gin-gonic/gin) - Web framework
- [Resty](https://github.com/go-resty/resty) - HTTP client
- [Levenshtein](https://github.com/texttheater/golang-levenshtein) - String similarity
- [Lo](https://github.com/samber/lo) - Utility functions for Go
- [Logrus](https://github.com/sirupsen/logrus) - Structured logging

## Docker

The application includes a Dockerfile for containerized deployment with **multi-platform support**. The Docker image is built using a multi-stage approach:

- **Builder stage**: Uses `golang:1.25-alpine` to compile the Go application for the target platform
- **Production stage**: Uses `alpine:latest` for a minimal runtime environment

### Multi-Platform Build Support

The Dockerfile now supports building for different platforms using `BUILDPLATFORM`, `TARGETOS`, and `TARGETARCH` variables:

```bash
# Build for a specific platform
docker build --platform linux/arm64 -t jellyfin-duplicate:arm64 .

# Build for multiple platforms using buildx
docker buildx build \
  --platform linux/amd64,linux/arm64,linux/arm/v7 \
  --build-arg TARGETOS=linux \
  -t jellyfin-duplicate:multiarch \
  --push .
```

### Available Build Script

A `docker-build.sh` script is provided for easy multi-platform building:

```bash
# Build for current platform
./docker-build.sh

# Build for specific platform
./docker-build.sh --platform linux/arm64

# Build for all platforms
./docker-build.sh --all

# Build and run
./docker-build.sh --run
```

The image includes:
- The compiled Go binary (built for the target platform)
- Production configuration (`config.prod.json`)
- HTML templates
- All required runtime dependencies

Image size is optimized by using Alpine Linux and multi-stage builds.

## Recommended Workflow

### Step-by-step guide for safe duplicate removal:

1. **Run analysis**: Access the `/analysis` page to see all potential duplicates
2. **Review duplicates**: Check the similarity percentage and file paths
3. **Check play status**: Look for the "✅ Safe to delete" notice or play status discrepancies
4. **Handle discrepancies**: If users have seen different versions:
   - Use the "Update Selected Users" button to synchronize play status
   - Select which users should have the other version marked as seen
5. **Delete safely**: Only delete movies that show the "✅ Safe to delete" notice
6. **Verify results**: Refresh the page to ensure the duplicate is removed

### Best Practices:

- **Always check play status** before deleting any movie
- **Synchronize play status** when users have seen different versions
- **Delete one version at a time** and verify the results
- **Backup your library** before making bulk deletions
- **Check file paths** to ensure you're deleting the correct version

## Troubleshooting

### Common issues and solutions:

**Issue**: "Not safe to delete" warning appears
- **Solution**: Check play status discrepancies and use the synchronization feature

**Issue**: Users have seen different versions
- **Solution**: Use the "Update Selected Users" button to mark the other version as seen

**Issue**: Delete button is disabled
- **Solution**: This means play status is not identical - synchronize first

**Issue**: API calls fail
- **Solution**: Check your Jellyfin API key and user permissions

## License

MIT