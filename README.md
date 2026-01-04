# Jellyfin Duplicate Finder

A Go application that finds duplicate movies in your Jellyfin server by analyzing movie metadata and file paths.

## Features

- Fetches all movies from your Jellyfin server
- Identifies potential duplicates based on TMDB/IMDB IDs
- Uses Levenshtein distance to analyze file paths
- Classifies duplicates vs mismatches (95% similarity threshold)
- Provides both HTML and JSON API endpoints

## Installation

1. Clone this repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Build the application:
   ```bash
   go build -o jellyfin-duplicate ./cmd/api
   ```

## Configuration

Edit the `cmd/api/main.go` file to configure:
- Jellyfin server URL
- API key
- User ID (for library access)

## Usage

Run the application:
```bash
./jellyfin-duplicate
```

Access the web interface at: `http://localhost:8080`

API endpoint: `http://localhost:8080/api/duplicates`

## How It Works

1. The application fetches all movies from your Jellyfin libraries
2. It groups movies by their TMDB or IMDB IDs
3. For each group with multiple movies, it compares file paths using Levenshtein distance
4. If path similarity is ≥95%, it's classified as a duplicate
5. If path similarity is <95%, it's classified as a mismatch

## Dependencies

- [Gin](https://github.com/gin-gonic/gin) - Web framework
- [Resty](https://github.com/go-resty/resty) - HTTP client
- [Levenshtein](https://github.com/texttheater/golang-levenshtein) - String similarity

## License

MIT