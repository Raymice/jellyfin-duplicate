# Jellyfin API Field Requirements

## Overview

This document explains the required fields for the Jellyfin API calls and ensures that all necessary data is retrieved for proper duplicate detection.

## Critical Fix Applied

### ❌ Issue Found
The Jellyfin API call was missing the `Path` field in the query parameters:

```go
// ❌ INCORRECT - Missing Path field
SetQueryParam("Fields", "ProviderIds,ProductionYear")
```

### ✅ Issue Fixed
Added `Path` to the fields parameter:

```go
// ✅ CORRECT - Includes all required fields
SetQueryParam("Fields", "ProviderIds,ProductionYear,Path")
```

## Complete API Field Requirements

### Required Fields for Duplicate Detection

| Field | Purpose | Required | Notes |
|-------|---------|----------|-------|
| **Path** | File path for similarity comparison | ✅ **YES** | Essential for duplicate detection |
| **ProductionYear** | Year for Name-Year key generation | ✅ **YES** | Used in grouping algorithm |
| **Name** | Movie name for Name-Year key generation | ✅ **YES** | Always included by default |
| **ProviderIds** | TMDB/IMDB IDs (legacy support) | ⚠️ Optional | Kept for reference |

### Why Each Field is Needed

#### 1. **Path** (CRITICAL)
```json
{
  "Path": "/movies/inception.mkv"
}
```

**Purpose:**
- Used for Levenshtein distance calculation
- Determines if files are duplicates or mismatches
- Essential for the core duplicate detection algorithm

**Impact if Missing:**
- ❌ Duplicate detection completely fails
- ❌ No path similarity comparison possible
- ❌ Application cannot function properly

#### 2. **ProductionYear** (CRITICAL)
```json
{
  "ProductionYear": 2010
}
```

**Purpose:**
- Used in Name-Year key generation (`"Inception-2010"`)
- Distinguishes remakes and different versions
- Essential for proper movie grouping

**Impact if Missing:**
- ❌ Movies grouped incorrectly
- ❌ Remakes treated as duplicates
- ❌ Key generation fails

#### 3. **Name** (CRITICAL)
```json
{
  "Name": "Inception"
}
```

**Purpose:**
- Used in Name-Year key generation
- Primary identifier for movies
- Always included by default in Jellyfin responses

**Impact if Missing:**
- ❌ Cannot identify movies
- ❌ Key generation impossible
- ❌ Basic functionality broken

#### 4. **ProviderIds** (Optional)
```json
{
  "ProviderIds": {
    "Tmdb": "27205",
    "Imdb": "tt1375666"
  }
}
```

**Purpose:**
- Legacy support for TMDB/IMDB IDs
- Reference information only
- Not used in current algorithm

**Impact if Missing:**
- ✅ No impact on current functionality
- ✅ Only affects legacy code paths

## Complete API Call

### Correct Implementation
```go
_, err := c.client.R().
    SetHeader("X-MediaBrowser-Token", c.apiKey).
    SetQueryParam("Recursive", "true").
    SetQueryParam("IncludeItemTypes", "Movie").
    SetQueryParam("Fields", "ProviderIds,ProductionYear,Path").  // ✅ ALL FIELDS
    SetQueryParam("ParentId", libraryID).
    SetQueryParam("StartIndex", fmt.Sprintf("%d", startIndex)).
    SetQueryParam("Limit", fmt.Sprintf("%d", limit)).
    SetResult(&result).
    Get(fmt.Sprintf("%s/Items", c.baseURL))
```

### Field Parameter Format
```
Fields=ProviderIds,ProductionYear,Path
```

- **Comma-separated** list of required fields
- **No spaces** between field names
- **Case-sensitive** field names

## API Response Structure

### Complete Movie Object
```json
{
  "Items": [
    {
      "Id": "movie123",
      "Name": "Inception",
      "Path": "/movies/inception.mkv",          // ✅ REQUIRED
      "ProductionYear": 2010,                    // ✅ REQUIRED
      "ProviderIds": {
        "Tmdb": "27205",
        "Imdb": "tt1375666"
      },
      "Type": "Movie",
      "RunTimeTicks": 96000000000
    }
  ],
  "TotalRecordCount": 150
}
```

## Error Scenarios and Solutions

### 1. Missing Path Field
**Symptom:** Empty or null path values

**Cause:**
```go
// ❌ WRONG
SetQueryParam("Fields", "ProviderIds,ProductionYear")
```

**Solution:**
```go
// ✅ CORRECT
SetQueryParam("Fields", "ProviderIds,ProductionYear,Path")
```

### 2. Missing ProductionYear Field
**Symptom:** Year = 0 for all movies

**Cause:**
```go
// ❌ WRONG
SetQueryParam("Fields", "ProviderIds,Path")
```

**Solution:**
```go
// ✅ CORRECT
SetQueryParam("Fields", "ProviderIds,ProductionYear,Path")
```

### 3. No Fields Specified
**Symptom:** Missing all optional fields

**Cause:**
```go
// ❌ WRONG - No Fields parameter
// Missing SetQueryParam("Fields", ...) entirely
```

**Solution:**
```go
// ✅ CORRECT
SetQueryParam("Fields", "ProviderIds,ProductionYear,Path")
```

## Testing Field Requirements

### Test Cases

```go
// Test 1: Verify Path field is included
func TestPathFieldIncluded(t *testing.T) {
    client := jellyfin.NewClient("http://test", "key")
    
    // Mock API call and verify Fields parameter
    // Should contain "Path"
}

// Test 2: Verify ProductionYear field is included
func TestProductionYearFieldIncluded(t *testing.T) {
    client := jellyfin.NewClient("http://test", "key")
    
    // Mock API call and verify Fields parameter
    // Should contain "ProductionYear"
}

// Test 3: Verify all required fields are present
func TestAllRequiredFieldsPresent(t *testing.T) {
    movies, err := client.GetAllMovies()
    if err != nil {
        t.Fatal(err)
    }
    
    for _, movie := range movies {
        if movie.Path == "" {
            t.Error("Movie missing Path field")
        }
        if movie.ProductionYear == 0 && movie.Name != "Home Video" {
            t.Error("Movie missing ProductionYear field")
        }
        if movie.Name == "" {
            t.Error("Movie missing Name field")
        }
    }
}
```

### Manual Testing

```bash
# Test API call with curl
curl -X GET "http://localhost:8096/Items" \
  -H "X-MediaBrowser-Token: YOUR_API_KEY" \
  -d "ParentId=YOUR_LIBRARY_ID" \
  -d "Fields=ProviderIds,ProductionYear,Path" \
  -d "IncludeItemTypes=Movie" \
  -d "Recursive=true" \
  -d "Limit=10"

# Verify response contains Path field
```

## Field Requirements Checklist

- [x] **Path** field included in API call
- [x] **ProductionYear** field included in API call
- [x] **Name** field (included by default)
- [x] Proper comma-separated format
- [x] No spaces in field list
- [x] Case-sensitive field names

## Impact of Missing Fields

### Missing Path Field
```
❌ Duplicate detection fails
❌ No path similarity calculation
❌ Application cannot determine duplicates
❌ Core functionality broken
```

### Missing ProductionYear Field
```
❌ Incorrect movie grouping
❌ Remakes treated as duplicates
❌ Key generation uses year=0
❌ Less accurate results
```

### Missing Both Fields
```
❌ Complete application failure
❌ No duplicate detection possible
❌ All functionality broken
❌ Requires code fix
```

## Best Practices

### 1. Always Specify Required Fields
```go
// ✅ GOOD
SetQueryParam("Fields", "ProviderIds,ProductionYear,Path")
```

### 2. Include All Fields in One Parameter
```go
// ✅ GOOD - Single Fields parameter
SetQueryParam("Fields", "Field1,Field2,Field3")

// ❌ BAD - Multiple Fields parameters
SetQueryParam("Fields", "Field1")
SetQueryParam("Fields", "Field2") // Overwrites previous
```

### 3. Use Consistent Field Order
```go
// ✅ GOOD - Consistent order
SetQueryParam("Fields", "ProviderIds,ProductionYear,Path")
```

### 4. Document Field Requirements
```go
// ✅ GOOD - Clear documentation
// Required fields: ProviderIds, ProductionYear, Path
// Optional fields: (none currently)
```

## API Field Reference

### Jellyfin Item Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| **Path** | string | File system path | ✅ Yes |
| **ProductionYear** | int | Release year | ✅ Yes |
| **Name** | string | Item name | ✅ Yes |
| **ProviderIds** | object | External IDs | ⚠️ No |
| **Type** | string | Item type | ⚠️ No |
| **RunTimeTicks** | long | Duration | ⚠️ No |
| **Overview** | string | Description | ⚠️ No |
| **Genres** | array | Genre tags | ⚠️ No |

### Minimum Required Fields
```
ProviderIds,ProductionYear,Path
```

### Recommended Fields
```
ProviderIds,ProductionYear,Path,Type,RunTimeTicks
```

## Troubleshooting

### Issue: Path Field Missing from Response

**Checklist:**
1. [ ] Verify `Fields` parameter includes `Path`
2. [ ] Check API call in browser developer tools
3. [ ] Test with curl to isolate issue
4. [ ] Verify Jellyfin server version supports Path field
5. [ ] Check user permissions for library access

**Solution:**
```go
// Ensure Path is in the Fields parameter
SetQueryParam("Fields", "ProviderIds,ProductionYear,Path")
```

### Issue: ProductionYear Always 0

**Checklist:**
1. [ ] Verify `Fields` parameter includes `ProductionYear`
2. [ ] Check if movies have year metadata in Jellyfin
3. [ ] Rescan library in Jellyfin web interface
4. [ ] Verify scraper is working properly

**Solution:**
```go
// Ensure ProductionYear is in the Fields parameter
SetQueryParam("Fields", "ProviderIds,ProductionYear,Path")
```

## Conclusion

The critical fix of adding `Path` to the Jellyfin API `Fields` parameter ensures that:

1. **✅ Duplicate detection works properly**
2. **✅ Path similarity calculations are accurate**
3. **✅ All required data is retrieved from Jellyfin**
4. **✅ The application functions as intended**

### Final Implementation
```go
// ✅ CORRECT - All required fields included
_, err := c.client.R().
    SetHeader("X-MediaBrowser-Token", c.apiKey).
    SetQueryParam("Recursive", "true").
    SetQueryParam("IncludeItemTypes", "Movie").
    SetQueryParam("Fields", "ProviderIds,ProductionYear,Path").  // ✅ CRITICAL
    SetQueryParam("ParentId", libraryID).
    SetQueryParam("StartIndex", fmt.Sprintf("%d", startIndex)).
    SetQueryParam("Limit", fmt.Sprintf("%d", limit)).
    SetResult(&result).
    Get(fmt.Sprintf("%s/Items", c.baseURL))
```

This fix resolves the core issue and ensures the application can properly retrieve and compare file paths for duplicate detection.