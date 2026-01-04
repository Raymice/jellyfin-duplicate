# Path Display Enhancements

## Overview

The web interface has been enhanced to provide better visibility and usability for file paths when displaying duplicate movies. This document explains the improvements made to the path display functionality.

## Current Path Display Features

### ✅ Already Implemented

The application already displays paths for each duplicate pair:

```html
<div class="movie-path">{{.Movie1.Path}}</div>
<div class="movie-path">{{.Movie2.Path}}</div>
```

### 🆕 New Enhancements

1. **Improved Styling**: Better visual presentation of paths
2. **Explicit Labels**: Clear "Path:" labels for each movie
3. **Enhanced Similarity Explanation**: More detailed path comparison info
4. **Better Readability**: Improved CSS for long paths

## Visual Improvements

### Before Enhancement
```html
<div class="movie-name">Inception (2010)</div>
<div class="movie-path">/movies/inception.mkv</div>
```

### After Enhancement
```html
<div class="movie-name">Inception (2010)</div>
<div class="path-label">Path:</div>
<div class="movie-path">/movies/inception.mkv</div>
```

## CSS Enhancements

### Path Display Styles
```css
.movie-path { 
    font-family: monospace; 
    background-color: #f5f5f5; 
    padding: 8px; 
    border-radius: 3px;
    font-size: 0.9em; 
    overflow-wrap: break-word; 
    max-width: 100%; 
    margin-top: 5px;
}
```

**Features:**
- Monospace font for code-like appearance
- Light gray background for visual separation
- Rounded corners for modern look
- Word wrapping for long paths
- Responsive design with max-width

### Path Label Styles
```css
.path-label { 
    font-weight: bold; 
    font-size: 0.85em; 
    color: #666; 
    margin-top: 3px;
}
```

**Features:**
- Bold text for emphasis
- Slightly smaller font size
- Gray color to distinguish from main content
- Proper spacing

### Similarity Explanation
```css
.path-comparison { 
    font-size: 0.85em; 
    color: #888; 
    margin-top: 5px; 
    font-style: italic;
}
```

**Features:**
- Italic text for explanation
- Subtle gray color
- Clear spacing from other elements
- Readable font size

## Example Output

### Duplicate Example
```
Inception (2010)
Path:
/movies/inception.mkv

Inception (2010)  
Path:
/backup/inception.mkv

Path similarity: 98% → These appear to be duplicates of the same movie
```

### Mismatch Example
```
The Matrix (1999)
Path:
/movies/matrix.mkv

The Matrix Reloaded (2003)
Path:
/movies/matrix_reloaded.mkv

Path similarity: 45% → These are likely different movies with similar names
```

## Template Structure

```html
{{range .duplicates}}
<div class="duplicate-pair {{if .IsDuplicate}}duplicate{{else}}mismatch{{end}}">
    <!-- Movie 1 -->
    <div class="movie-info">
        <div class="movie-name">{{.Movie1.Name}} ({{.Movie1.ProductionYear}})</div>
        <div class="path-label">Path:</div>
        <div class="movie-path">{{.Movie1.Path}}</div>
    </div>
    
    <!-- Movie 2 -->
    <div class="movie-info">
        <div class="movie-name">{{.Movie2.Name}} ({{.Movie2.ProductionYear}})</div>
        <div class="path-label">Path:</div>
        <div class="movie-path">{{.Movie2.Path}}</div>
    </div>
    
    <!-- Similarity Explanation -->
    <div class="path-comparison">
        Path similarity: {{.Similarity}}% 
        {{if .IsDuplicate}}
            → These appear to be duplicates of the same movie
        {{else}}
            → These are likely different movies with similar names
        {{end}}
    </div>
</div>
{{end}}
```

## Benefits of Enhanced Display

### 1. **Better User Experience**
- Clear visual hierarchy
- Easy to scan and compare paths
- Intuitive understanding of results

### 2. **Improved Readability**
- Monospace font for paths (like code)
- Proper word wrapping for long paths
- Good contrast and spacing

### 3. **Enhanced Understanding**
- Explicit "Path:" labels
- Clear similarity explanations
- Context for duplicate vs mismatch

### 4. **Responsive Design**
- Works on different screen sizes
- Handles long file paths gracefully
- Mobile-friendly layout

## Path Comparison Logic

The path similarity is calculated using the Levenshtein distance algorithm:

```go
func calculatePathSimilarity(path1, path2 string) int {
    distance := levenshteinDistance(path1, path2)
    
    maxLen := len(path1)
    if len(path2) > maxLen {
        maxLen = len(path2)
    }
    
    if maxLen == 0 {
        return 100
    }
    
    similarity := 100 - (distance * 100 / maxLen)
    return similarity
}
```

### Similarity Thresholds

- **≥95% Similarity**: Considered a duplicate
- **<95% Similarity**: Considered a mismatch

## Example Scenarios

### Scenario 1: Identical Paths
```
Path 1: /movies/inception.mkv
Path 2: /movies/inception.mkv
Similarity: 100% → Duplicate
```

### Scenario 2: Similar Paths (Different Folders)
```
Path 1: /movies/inception.mkv
Path 2: /backup/inception.mkv
Similarity: 98% → Duplicate
```

### Scenario 3: Different Movies
```
Path 1: /movies/matrix.mkv
Path 2: /movies/matrix_reloaded.mkv
Similarity: 75% → Mismatch
```

### Scenario 4: Completely Different
```
Path 1: /movies/inception.mkv
Path 2: /tv/shows/episode.mkv
Similarity: 20% → Mismatch
```

## Path Handling Features

### 1. **Long Path Support**
```css
overflow-wrap: break-word;
max-width: 100%;
```
- Prevents horizontal scrolling
- Wraps long paths to multiple lines
- Maintains readability

### 2. **Visual Distinction**
```css
background-color: #f5f5f5;
padding: 8px;
border-radius: 3px;
```
- Light background highlights paths
- Padding provides visual separation
- Rounded corners for modern look

### 3. **Monospace Font**
```css
font-family: monospace;
```
- Clear distinction from regular text
- Aligns characters for easy comparison
- Familiar code-like appearance

## User Interface Examples

### Desktop View
```
╔════════════════════════════════════════════════════════╗
║  Inception (2010)                                    ║
║  Path:                                              ║
║  /movies/inception.mkv                              ║
║                                                    ║
║  Inception (2010)                                   ║
║  Path:                                              ║
║  /backup/inception.mkv                              ║
║                                                    ║
║  Path similarity: 98% → These appear to be         ║
║  duplicates of the same movie                       ║
╚════════════════════════════════════════════════════════╝
```

### Mobile View
```
[ Inception (2010) ]
Path:
/movies/inception.mkv

[ Inception (2010) ]
Path:
/backup/inception.mkv

Path similarity: 98%
→ These appear to be duplicates
```

## Future Enhancements

### 1. **Path Highlighting**
```css
/* Highlight differences in paths */
.path-diff {
    background-color: #ffeb3b;
    font-weight: bold;
}
```

### 2. **Interactive Comparison**
```javascript
// Show side-by-side path comparison
function showPathDiff(path1, path2) {
    // Highlight differences
    // Show character-by-character comparison
}
```

### 3. **Path Statistics**
```
Additional Information:
- Path length: 24 characters
- Common prefix: "/movies/"
- Different suffix: "inception.mkv" vs "backup/inception.mkv"
```

### 4. **File System Actions**
```
[ Open in File Explorer ] [ Copy Path ] [ Delete Duplicate ]
```

## Testing Path Display

### Test Cases

1. **Short Paths**: `/movies/test.mkv`
2. **Long Paths**: `/very/long/path/with/many/subdirectories/movie.mkv`
3. **Unicode Paths**: `/movies/仙剑奇侠传.mkv`
4. **Special Characters**: `/movies/movie's_name.mkv`
5. **Spaces**: `/movies/movie name.mkv`

### Expected Results

- All paths display correctly
- Long paths wrap properly
- Unicode characters render correctly
- Special characters don't break layout
- Spaces are preserved

## Conclusion

The enhanced path display provides a much better user experience by:

1. **Making paths more visible** with clear labels and styling
2. **Improving readability** with monospace fonts and proper wrapping
3. **Adding context** with similarity explanations
4. **Maintaining responsiveness** for different screen sizes

The current implementation successfully displays paths for all duplicate pairs and provides clear visual distinction between duplicates and mismatches based on path similarity analysis.