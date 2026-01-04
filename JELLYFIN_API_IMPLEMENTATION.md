# Jellyfin API Implementation Guide

## Overview

This document explains how the Jellyfin Duplicate Finder interacts with the Jellyfin API to retrieve movie data, with a focus on the correct HTTP calls, filters, and parameters.

## API Endpoints Used

### 1. Get User Libraries

**Endpoint**: `GET /Users/{UserId}/Views`

**Purpose**: Retrieve all libraries accessible to the specified user

**Parameters**:
- `UserId`: The Jellyfin user ID (from config)

**Headers**:
- `X-MediaBrowser-Token`: API key for authentication

**Response**:
```json
{
  "Items": [
    {
      "Id": "library1",
      "Name": "Movies",
      "CollectionType": "movies"
    },
    {
      "Id": "library2",
      "Name": "TV Shows",
      "CollectionType": "tvshows"
    }
  ]
}
```

### 2. Get Movies from Library

**Endpoint**: `GET /Items`

**Purpose**: Retrieve all movies from a specific library with pagination support

**Parameters**:
- `ParentId`: The library ID to filter by
- `Recursive`: `true` to include items in subfolders
- `IncludeItemTypes`: `Movie` to filter only movies
- `Fields`: `ProviderIds,ProductionYear` to include specific fields
- `StartIndex`: For pagination (0, 100, 200, ...)
- `Limit`: Number of items per page (default: 100)

**Headers**:
- `X-MediaBrowser-Token`: API key for authentication

**Response**:
```json
{
  "Items": [
    {
      "Id": "movie1",
      "Name": "Inception",
      "Path": "/movies/inception.mkv",
      "ProductionYear": 2010,
      "ProviderIds": {
        "Tmdb": "27205",
        "Imdb": "tt1375666"
      }
    }
  ],
  "TotalRecordCount": 150
}
```

## Implementation Details

### Complete API Client Code

```go
package jellyfin

import (
    "fmt"
    "jellyfin-duplicate/internal/models"
    "github.com/go-resty/resty/v2"
)

type Client struct {
    baseURL string
    apiKey  string
    userID  string
    client  *resty.Client
}

func NewClient(baseURL, apiKey string) *Client {
    return &Client{
        baseURL: baseURL,
        apiKey:  apiKey,
        client:  resty.New(),
    }
}

func (c *Client) SetUserID(userID string) error {
    c.userID = userID
    return nil
}

func (c *Client) GetAllMovies() ([]models.Movie, error) {
    var movies []models.Movie
    
    // Step 1: Get all libraries accessible to the user
    libraries, err := c.getLibraries()
    if err != nil {
        return nil, fmt.Errorf("failed to get libraries: %v", err)
    }
    
    // Step 2: For each library, get all movies
    for _, library := range libraries {
        libraryMovies, err := c.getMoviesFromLibrary(library.ID)
        if err != nil {
            return nil, fmt.Errorf("failed to get movies from library %s: %v", library.Name, err)
        }
        movies = append(movies, libraryMovies...)
    }
    
    return movies, nil
}

func (c *Client) getLibraries() ([]Library, error) {
    if c.userID == "" {
        return nil, fmt.Errorf("user ID not set")
    }
    
    var result struct {
        Items []Library `json:"Items"`
    }
    
    _, err := c.client.R().
        SetHeader("X-MediaBrowser-Token", c.apiKey).
        SetResult(&result).
        Get(fmt.Sprintf("%s/Users/%s/Views", c.baseURL, c.userID))
    
    if err != nil {
        return nil, err
    }
    
    return result.Items, nil
}

func (c *Client) getMoviesFromLibrary(libraryID string) ([]models.Movie, error) {
    var allMovies []models.Movie
    
    // Pagination support for large libraries
    startIndex := 0
    limit := 100 // Can be adjusted based on performance needs
    
    for {
        var result struct {
            Items            []models.Movie `json:"Items"`
            TotalRecordCount  int           `json:"TotalRecordCount"`
        }
        
        _, err := c.client.R().
            SetHeader("X-MediaBrowser-Token", c.apiKey).
            SetQueryParam("Recursive", "true").
            SetQueryParam("IncludeItemTypes", "Movie").
            SetQueryParam("Fields", "ProviderIds,ProductionYear").
            SetQueryParam("ParentId", libraryID).
            SetQueryParam("StartIndex", fmt.Sprintf("%d", startIndex)).
            SetQueryParam("Limit", fmt.Sprintf("%d", limit)).
            SetResult(&result).
            Get(fmt.Sprintf("%s/Items", c.baseURL))
        
        if err != nil {
            return nil, err
        }
        
        // Add movies from this page
        allMovies = append(allMovies, result.Items...)
        
        // Check if we've fetched all movies
        if len(allMovies) >= result.TotalRecordCount {
            break
        }
        
        // Move to next page
        startIndex += limit
    }
    
    return allMovies, nil
}

type Library struct {
    ID   string `json:"Id"`
    Name string `json:"Name"`
}
```

## Key Parameters Explained

### 1. Authentication

- **`X-MediaBrowser-Token`**: Required for all API calls
- **User ID**: Specified in config.json, used to get accessible libraries

### 2. Filtering

- **`ParentId`**: Filters items to only those in the specified library
- **`IncludeItemTypes=Movie`**: Ensures only movies are returned
- **`Recursive=true`**: Includes movies in subfolders

### 3. Fields

- **`Fields=ProviderIds,ProductionYear`**: Requests specific fields to reduce payload
- **ProviderIds**: Contains TMDB/IMDB IDs (kept for reference)
- **ProductionYear**: Essential for the new key strategy

### 4. Pagination

- **`StartIndex`**: Starting position for pagination
- **`Limit`**: Number of items per page
- **`TotalRecordCount`**: Total number of items available

## Error Handling

### Common Error Scenarios

1. **Authentication Failure**
   - Invalid API key
   - Expired token
   - User permissions issue

2. **Network Issues**
   - Server unavailable
   - Timeout
   - DNS resolution failure

3. **Data Issues**
   - Invalid library ID
   - Corrupted response
   - Missing required fields

### Error Handling Strategy

```go
// Example error handling in the client
if err != nil {
    // Check for specific error types
    if restyErr, ok := err.(*resty.ResponseError); ok {
        statusCode := restyErr.Response.StatusCode()
        if statusCode == 401 {
            return nil, fmt.Errorf("authentication failed: %v", err)
        } else if statusCode == 404 {
            return nil, fmt.Errorf("library not found: %v", err)
        }
    }
    return nil, fmt.Errorf("API request failed: %v", err)
}
```

## Performance Considerations

### 1. Pagination Strategy

- **Default Limit**: 100 items per request (Jellyfin default)
- **Adjustable**: Can be increased for better performance
- **Trade-off**: Larger limits vs. memory usage

### 2. Field Selection

- **Minimal Fields**: Only request needed fields
- **Reduced Payload**: Faster transfers, less parsing
- **Future-proof**: Easy to add more fields if needed

### 3. Caching

**Potential Optimization**:
```go
// Cache results to avoid repeated API calls
var movieCache = make(map[string][]models.Movie)

func (c *Client) getMoviesFromLibrary(libraryID string) ([]models.Movie, error) {
    if movies, cached := movieCache[libraryID]; cached {
        return movies, nil
    }
    
    // Fetch from API
    movies, err := c.fetchMoviesFromLibrary(libraryID)
    if err != nil {
        return nil, err
    }
    
    // Cache results
    movieCache[libraryID] = movies
    return movies, nil
}
```

### 4. Parallel Requests

**Potential Optimization**:
```go
// Fetch from multiple libraries in parallel
var wg sync.WaitGroup
var mu sync.Mutex
var allMovies []models.Movie

for _, library := range libraries {
    wg.Add(1)
    go func(lib Library) {
        defer wg.Done()
        movies, err := c.getMoviesFromLibrary(lib.ID)
        if err != nil {
            // Handle error
            return
        }
        
        mu.Lock()
        allMovies = append(allMovies, movies...)
        mu.Unlock()
    }(library)
}

wg.Wait()
```

## API Rate Limiting

### Best Practices

1. **Respect Server Limits**: Don't overload the Jellyfin server
2. **Add Delays**: For large libraries, add small delays between requests
3. **Retry Logic**: Implement exponential backoff for failed requests
4. **Monitor Performance**: Watch for slow responses or timeouts

### Example Rate Limiting

```go
// Add delay between paginated requests
func (c *Client) getMoviesFromLibrary(libraryID string) ([]models.Movie, error) {
    // ... existing code ...
    
    for {
        // ... make API request ...
        
        // Add small delay for large libraries
        if startIndex > 0 {
            time.Sleep(100 * time.Millisecond)
        }
        
        // ... rest of the code ...
    }
}
```

## Debugging API Issues

### 1. Enable Debug Logging

```go
// Create client with debug logging
client := resty.New()
if os.Getenv("DEBUG") == "1" {
    client.SetDebug(true)
    client.SetLogger(log.New(os.Stdout, "[JELLYFIN] ", log.LstdFlags))
}
```

### 2. Log API Responses

```go
// Log response details for debugging
resp, err := c.client.R().
    SetHeader("X-MediaBrowser-Token", c.apiKey).
    // ... other settings ...
    SetResult(&result).
    Get(apiURL)

if os.Getenv("DEBUG") == "1" {
    log.Printf("API Response: Status=%d, Body=%s", 
        resp.StatusCode(), 
        string(resp.Body()))
}
```

### 3. Common Debugging Steps

1. **Verify API URL**: Ensure the base URL is correct
2. **Check Authentication**: Confirm API key and user ID are valid
3. **Test with curl**: Manually test API calls
4. **Inspect Responses**: Look for error messages or unexpected data
5. **Check Network**: Ensure connectivity to Jellyfin server

## Testing the API Implementation

### Manual Testing with curl

```bash
# Test getting libraries
curl -X GET "http://localhost:8096/Users/{UserId}/Views" \
  -H "X-MediaBrowser-Token: {API_KEY}"

# Test getting movies from a library
curl -X GET "http://localhost:8096/Items" \
  -H "X-MediaBrowser-Token: {API_KEY}" \
  -d "ParentId={LIBRARY_ID}" \
  -d "Recursive=true" \
  -d "IncludeItemTypes=Movie" \
  -d "Fields=ProviderIds,ProductionYear" \
  -d "StartIndex=0" \
  -d "Limit=10"
```

### Automated Testing

```go
// Test the Jellyfin client
func TestJellyfinClient(t *testing.T) {
    // Mock server setup would go here
    
    client := jellyfin.NewClient("http://test-server:8096", "test-api-key")
    client.SetUserID("test-user-id")
    
    // Test getLibraries
    libraries, err := client.getLibraries()
    if err != nil {
        t.Fatalf("Failed to get libraries: %v", err)
    }
    
    if len(libraries) == 0 {
        t.Error("Expected at least one library")
    }
    
    // Test getMoviesFromLibrary
    movies, err := client.getMoviesFromLibrary(libraries[0].ID)
    if err != nil {
        t.Fatalf("Failed to get movies: %v", err)
    }
    
    // Verify movies have required fields
    for _, movie := range movies {
        if movie.Name == "" {
            t.Error("Movie missing name")
        }
        if movie.Path == "" {
            t.Error("Movie missing path")
        }
        // ProductionYear can be 0 for some movies
    }
}
```

## Troubleshooting Guide

### Common Issues and Solutions

#### 1. "401 Unauthorized" Error

**Causes**:
- Invalid API key
- Incorrect user ID
- Expired token

**Solutions**:
- Verify API key in config.json
- Check user ID in config.json
- Regenerate API key in Jellyfin settings

#### 2. "404 Not Found" Error

**Causes**:
- Incorrect Jellyfin server URL
- Invalid library ID
- Wrong endpoint path

**Solutions**:
- Verify base URL in config.json
- Check that Jellyfin server is running
- Test with curl to isolate the issue

#### 3. No Movies Returned

**Causes**:
- Empty library
- Incorrect filters
- Permission issues

**Solutions**:
- Verify library contains movies in Jellyfin web interface
- Check filters (IncludeItemTypes=Movie)
- Ensure user has access to the library

#### 4. Slow Performance

**Causes**:
- Large library without pagination
- Network latency
- Server performance issues

**Solutions**:
- Implement proper pagination
- Adjust limit parameter
- Check Jellyfin server performance

#### 5. Missing ProductionYear

**Causes**:
- Movies not properly scraped
- Metadata issues
- Permission problems

**Solutions**:
- Rescan library in Jellyfin
- Check movie metadata in web interface
- Verify user permissions

## API Reference

### Official Jellyfin API Documentation

- **API Docs**: https://api.jellyfin.org/
- **GitHub**: https://github.com/jellyfin/jellyfin
- **Swagger UI**: Available on your Jellyfin server at `/swagger`

### Key Endpoints for This Application

1. **GET /Users/{UserId}/Views** - Get user libraries
2. **GET /Items** - Get items with filtering
3. **GET /Items/{Id}** - Get specific item details (not used currently)

### Useful API Tools

1. **Postman**: For API testing and exploration
2. **curl**: Command-line HTTP client
3. **Jellyfin Swagger UI**: Built-in API documentation
4. **resty**: Go HTTP client used in this project

## Conclusion

The Jellyfin API implementation in this project follows best practices for:

- **Authentication**: Proper use of API keys and user IDs
- **Filtering**: Efficient filtering by library and item type
- **Pagination**: Handling large libraries with multiple requests
- **Error Handling**: Graceful handling of API errors
- **Performance**: Minimal field selection and efficient requests

The current implementation provides a solid foundation that can be extended with additional features like caching, parallel requests, and more sophisticated error handling as needed.