# Duplicate Detection Key Strategy Change

## Overview

The duplicate detection algorithm has been updated to use a new key strategy for grouping movies. Instead of relying on TMDB/IMDB IDs, the application now uses `{movie.Name}-{movie.ProductionYear}` as the primary key for identifying potential duplicates.

## Rationale for the Change

### Problems with TMDB/IMDB ID Approach

1. **Missing Metadata**: Many movies in Jellyfin libraries may not have TMDB or IMDB IDs, especially:
   - Home movies
   - Custom content
   - Poorly scraped media
   - User-uploaded content

2. **Incorrect Metadata**: Automated scrapers can sometimes assign wrong IDs

3. **Limited Scope**: Only works for movies with proper metadata from major databases

4. **Complex Logic**: Required fallback logic and special handling for missing IDs

### Benefits of Name-Year Approach

1. **Universal Coverage**: Every movie has a name and production year
2. **Simpler Logic**: No need for complex fallback mechanisms
3. **More Intuitive**: Matches how humans identify movies
4. **Better for Remakes**: Properly handles movies with the same name but different years
5. **Works with All Content**: Including home movies and custom content

## Implementation Details

### Key Generation

```go
// Old approach (removed)
var key string
if movie.ProviderIds.Tmdb != "" {
    key = "tmdb:" + movie.ProviderIds.Tmdb
} else if movie.ProviderIds.Imdb != "" {
    key = "imdb:" + movie.ProviderIds.Imdb
} else {
    // Skip movies without provider IDs
    continue
}

// New approach
key := fmt.Sprintf("%s-%d", movie.Name, movie.ProductionYear)
```

### Data Model Changes

Added `ProductionYear` field to the `Movie` struct:

```go
type Movie struct {
    ID             string `json:"Id"`
    Name           string `json:"Name"`
    Path           string `json:"Path"`
    ProductionYear int    `json:"ProductionYear"`  // NEW FIELD
    ProviderIds    struct {
        Tmdb string `json:"Tmdb"`
        Imdb string `json:"Imdb"`
    } `json:"ProviderIds"`
}
```

### Jellyfin API Changes

Updated the API request to include `ProductionYear` in the fields:

```go
SetQueryParam("Fields", "ProviderIds,ProductionYear")  // Added ProductionYear
```

### Duplicate Detection Logic

The core algorithm remains the same, but now groups movies by the new key:

```go
// Create a map to group movies by their Name and ProductionYear
movieMap := make(map[string][]models.Movie)

for _, movie := range movies {
    // Use Name-ProductionYear as the key
    // This handles cases where movies have the same name but different years
    key := fmt.Sprintf("%s-%d", movie.Name, movie.ProductionYear)
    
    movieMap[key] = append(movieMap[key], movie)
}
```

## Examples of Key Generation

### Example 1: Same Movie, Different Files
```
Movie 1: "The Matrix" (1999) -> Key: "The Matrix-1999"
Movie 2: "The Matrix" (1999) -> Key: "The Matrix-1999"
Result: Grouped together for duplicate detection
```

### Example 2: Remakes (Different Years)
```
Movie 1: "King Kong" (1933) -> Key: "King Kong-1933"
Movie 2: "King Kong" (2005) -> Key: "King Kong-2005"
Result: Different keys, not considered duplicates
```

### Example 3: Same Movie, Different Names
```
Movie 1: "Star Wars: Episode IV - A New Hope" (1977) -> Key: "Star Wars: Episode IV - A New Hope-1977"
Movie 2: "Star Wars" (1977) -> Key: "Star Wars-1977"
Result: Different keys, not grouped together
```

## Edge Cases Handled

### 1. Missing Production Year

If a movie has `ProductionYear = 0`:
- Key becomes: `"MovieName-0"`
- Still works for grouping
- Can be improved with better year detection

### 2. Special Characters in Names

The key generation handles special characters naturally:
```
"Schindler's List" (1993) -> "Schindler's List-1993"
"Pulp Fiction" (1994) -> "Pulp Fiction-1994"
```

### 3. Unicode Characters

Unicode characters in movie names are preserved:
```
"仙剑奇侠传" (2005) -> "仙剑奇侠传-2005"
"ポケモン" (1997) -> "ポケモン-1997"
```

## Comparison with Previous Approach

### Old Approach (TMDB/IMDB IDs)

**Pros:**
- Very accurate when IDs are correct
- Works well for well-scraped commercial movies

**Cons:**
- Fails for movies without IDs
- Complex fallback logic needed
- Doesn't handle remakes well
- Requires external metadata

### New Approach (Name-Year)

**Pros:**
- Works for all movies
- Simple and intuitive
- Handles remakes correctly
- No external dependencies
- Universal coverage

**Cons:**
- Less precise for movies with similar names
- Depends on accurate name/year metadata
- May group different movies with same name/year

## Impact on Duplicate Detection

### Positive Impacts

1. **Increased Coverage**: Now detects duplicates in movies that previously had no IDs
2. **Simpler Logic**: No need to skip movies or handle missing ID cases
3. **Better User Experience**: More intuitive grouping based on visible metadata
4. **Remake Handling**: Properly distinguishes between different versions

### Potential Challenges

1. **Name Variations**: Movies with slightly different names won't be grouped
2. **Year Accuracy**: Incorrect years may cause false groupings
3. **Special Editions**: Different editions of the same movie may not be grouped

## Recommendations for Users

### 1. Ensure Accurate Metadata

For best results:
- Use reliable scrapers for your Jellyfin library
- Manually verify metadata for important movies
- Correct any obvious year errors

### 2. Understanding the Results

- **Duplicates (≥95% similarity)**: Same movie in different locations/folders
- **Mismatches (<95% similarity)**: Different movies with similar names/years

### 3. Manual Review

Always review the results, especially for:
- Movies with similar names
- Remakes and reboots
- Special editions and director's cuts

## Future Enhancements

### 1. Fuzzy Name Matching

Add optional fuzzy matching for movie names:
```go
// Pseudocode for future enhancement
if levenshteinDistance(movie1.Name, movie2.Name) < threshold {
    // Consider as potential match even if names aren't identical
}
```

### 2. Year Normalization

Handle year variations (e.g., "2020" vs "2020s"):
```go
// Extract year from various formats
func normalizeYear(yearField interface{}) int {
    // Handle "2020", "2020s", "Early 2020s", etc.
}
```

### 3. Alternative Titles

Consider alternative titles in the key:
```go
key := fmt.Sprintf("%s-%d", movie.PrimaryName, movie.ProductionYear)
// Also check: movie.AlternativeTitles
```

### 4. Configurable Key Strategy

Allow users to choose between different key strategies:
```go
// Configuration option
KeyStrategy: "name-year" | "tmdb-id" | "imdb-id" | "hybrid"
```

## Testing the New Approach

### Test Cases

1. **Identical Movies**: Same name, same year, different paths
2. **Remakes**: Same name, different years
3. **Sequels**: Different names, different years
4. **Special Editions**: Similar names, same year
5. **Missing Years**: Movies with year = 0

### Expected Results

| Scenario | Expected Behavior |
|----------|-------------------|
| Same movie, different locations | Grouped together, high similarity |
| Remakes (same name, different years) | Different groups |
| Sequels (different names) | Different groups |
| Same name, same year, different movies | Grouped together, low similarity |

## Migration Guide

### For Existing Users

1. **Update Configuration**: No changes needed
2. **Review Results**: New duplicates may appear
3. **Adjust Thresholds**: May need to tweak similarity thresholds
4. **Verify Metadata**: Check for accurate names and years

### For New Users

1. **Ensure Good Metadata**: Use reliable scrapers
2. **Start with Defaults**: Use default similarity threshold (95%)
3. **Review Initial Results**: Understand how movies are grouped
4. **Adjust as Needed**: Fine-tune based on your library

## Conclusion

The change from TMDB/IMDB IDs to Name-Year keys represents a significant improvement in the duplicate detection algorithm. It provides universal coverage, simpler logic, and more intuitive results while maintaining the same core functionality. Users can expect to find more potential duplicates, especially in libraries with mixed or incomplete metadata.

The new approach is particularly beneficial for:
- Libraries with home movies or custom content
- Collections with mixed metadata quality
- Users who prefer simpler, more transparent grouping logic
- Scenarios involving remakes and reboots

While the new method may have slightly less precision for commercial movies with perfect metadata, the overall improvement in coverage and usability makes it the better choice for most Jellyfin users.