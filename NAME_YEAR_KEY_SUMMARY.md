# Summary: Name-Year Key Strategy Implementation

## 🎯 Overview

The Jellyfin Duplicate Finder has been successfully updated to use `{movie.Name}-{movie.ProductionYear}` as the primary key for duplicate detection, replacing the previous TMDB/IMDB ID-based approach.

## ✅ Changes Implemented

### 1. **Data Model Update** (`internal/models/movie.go`)
```go
type Movie struct {
    ID             string `json:"Id"`
    Name           string `json:"Name"`
    Path           string `json:"Path"`
    ProductionYear int    `json:"ProductionYear"`  // ✅ ADDED
    ProviderIds    struct {
        Tmdb string `json:"Tmdb"`
        Imdb string `json:"Imdb"`
    } `json:"ProviderIds"`
}
```

### 2. **Jellyfin API Update** (`internal/jellyfin/client.go`)
```go
// Updated to fetch ProductionYear field
SetQueryParam("Fields", "ProviderIds,ProductionYear")
```

### 3. **Duplicate Detection Logic** (`internal/handlers/handler.go`)
```go
// New key generation strategy
key := fmt.Sprintf("%s-%d", movie.Name, movie.ProductionYear)

// Old approach (removed):
// if movie.ProviderIds.Tmdb != "" {
//     key = "tmdb:" + movie.ProviderIds.Tmdb
// } else if movie.ProviderIds.Imdb != "" {
//     key = "imdb:" + movie.ProviderIds.Imdb
// } else {
//     continue  // ❌ Skipped movies without IDs
// }
```

### 4. **UI Enhancement** (`web/templates/duplicates.html`)
```html
<!-- Added production year display -->
<div class="movie-name">{{.Movie1.Name}} ({{.Movie1.ProductionYear}})</div>
<div class="movie-name">{{.Movie2.Name}} ({{.Movie2.ProductionYear}})</div>
```

## 🔧 Technical Implementation

### Key Generation Algorithm
```go
func (h *Handler) findDuplicates() ([]models.DuplicateResult, error) {
    movies, err := h.jellyfinClient.GetAllMovies()
    if err != nil {
        return nil, err
    }
    
    // Create a map to group movies by their Name and ProductionYear
    movieMap := make(map[string][]models.Movie)
    
    for _, movie := range movies {
        // Use Name-ProductionYear as the key
        // This handles cases where movies have the same name but different years
        key := fmt.Sprintf("%s-%d", movie.Name, movie.ProductionYear)
        
        movieMap[key] = append(movieMap[key], movie)
    }
    
    // Rest of duplicate detection logic remains the same
    // ... (path similarity comparison, etc.)
}
```

## 🎬 Example Scenarios

### ✅ Correctly Identified Duplicates
```
Movie A: "Inception" (2010) from /movies/inception.mkv
Movie B: "Inception" (2010) from /backup/inception.mkv
Key: "Inception-2010" (both)
Result: Grouped together → Path similarity check → Marked as duplicate
```

### ✅ Properly Separated Remakes
```
Movie A: "King Kong" (1933)
Movie B: "King Kong" (2005)
Key A: "King Kong-1933"
Key B: "King Kong-2005"
Result: Different keys → Not considered duplicates
```

### ✅ Handles Edge Cases
```
// Missing year (falls back to 0)
Movie: "Home Video" (0)
Key: "Home Video-0"

// Unicode characters
Movie: "仙剑奇侠传" (2005)
Key: "仙剑奇侠传-2005"

// Special characters
Movie: "Schindler's List" (1993)
Key: "Schindler's List-1993"
```

## 📊 Comparison: Old vs New Approach

| Aspect | Old Approach (TMDB/IMDB) | New Approach (Name-Year) |
|--------|--------------------------|--------------------------|
| **Coverage** | Limited to movies with IDs | Works for ALL movies |
| **Complexity** | Complex fallback logic | Simple, straightforward |
| **Remake Handling** | Poor (same ID) | Excellent (different years) |
| **Custom Content** | Fails (no IDs) | Works perfectly |
| **Dependency** | External metadata | Built-in movie data |
| **User Experience** | Confusing skips | Intuitive grouping |

## 🚀 Benefits

### 1. **Universal Coverage**
- Works for home movies, custom content, and poorly scraped media
- No more "skipped movies" due to missing IDs

### 2. **Simpler Logic**
- No complex fallback chains
- Cleaner, more maintainable code
- Easier to understand and debug

### 3. **Better Remake Handling**
- Properly distinguishes between different versions
- "King Kong" (1933) ≠ "King Kong" (2005)
- "Batman" (1989) ≠ "Batman" (2022)

### 4. **More Intuitive**
- Matches how humans identify movies
- Visible metadata (name/year) vs hidden IDs
- Easier to explain to users

### 5. **Robust**
- Handles Unicode characters
- Works with special characters
- Gracefully handles missing years

## ⚠️ Considerations

### Potential Limitations

1. **Name Variations**: "Star Wars" vs "Star Wars: Episode IV"
   - **Solution**: Consider adding fuzzy name matching

2. **Year Accuracy**: Incorrect years may cause issues
   - **Solution**: Use reliable scrapers, manual verification

3. **Special Editions**: "Movie (Director's Cut)" vs "Movie (Theatrical)"
   - **Solution**: Could add edition detection

### When Old Approach Might Be Better

- Libraries with perfect TMDB/IMDB metadata
- Commercial movies only (no custom content)
- Need for absolute precision over coverage

## 🔮 Future Enhancements

### 1. **Hybrid Approach**
```go
// Configurable strategy
if config.UseHybridKey {
    if movie.ProviderIds.Tmdb != "" {
        key = "tmdb:" + movie.ProviderIds.Tmdb
    } else {
        key = fmt.Sprintf("%s-%d", movie.Name, movie.ProductionYear)
    }
}
```

### 2. **Fuzzy Name Matching**
```go
// For movies with similar names
if levenshteinDistance(movie1.Name, movie2.Name) < 5 {
    // Consider as potential match
}
```

### 3. **Year Normalization**
```go
// Handle "2020s", "Early 2020", etc.
func normalizeYear(yearField interface{}) int {
    // Extract clean year from various formats
}
```

### 4. **Alternative Titles**
```go
// Consider alternative titles
key := fmt.Sprintf("%s-%d", movie.PrimaryName, movie.ProductionYear)
if contains(movie.AlternativeTitles, "Original Title") {
    altKey := fmt.Sprintf("%s-%d", "Original Title", movie.ProductionYear)
    // Check both keys
}
```

## 📚 Documentation Updates

### Updated Files
- `KEY_STRATEGY_CHANGE.md` - Detailed explanation of the change
- `LEVENSHTEIN_IMPLEMENTATION.md` - Pure Go Levenshtein implementation
- `README.md` - Updated with new approach
- `SETUP_GUIDE.md` - No changes needed (backward compatible)

### New Documentation
- `NAME_YEAR_KEY_SUMMARY.md` - This summary
- Comprehensive test cases and examples
- Migration guide for existing users

## 🎯 Conclusion

The transition to the Name-Year key strategy represents a significant improvement in the Jellyfin Duplicate Finder's functionality. It provides:

- **Better coverage** for all types of movies
- **Simpler, more maintainable** code
- **More intuitive** results that match user expectations
- **Robust handling** of edge cases and special content

This change makes the application more useful for a wider range of Jellyfin users, especially those with diverse libraries containing home movies, custom content, or mixed metadata quality.

**The new approach is recommended for most users** and provides a solid foundation for future enhancements like fuzzy matching and hybrid strategies.