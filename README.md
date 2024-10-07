
# RSS Feed Processor

## Endpoints

The program exposes two HTTP endpoints:

### 1. `/rss-path` (POST)

- **Description**: Accepts RSS feed URLs and file paths as input for processing.
- **Request Body**:
   ```json
   {
       "urls": ["https://feeds.bbci.co.uk/news/rss.xml"],
       "filePaths": ["/path/to/rssFeed.xml"]
   }
   ```
- **Response**: Success message indicating that the URLs and file paths are being processed in the background.
  ```json
     {
         "status": "request acknowledged"
     }
     ```

### 2. `/rss-feeds` (GET)

- **Description**: Exposes the processed RSS feed summary as an HTML page. The feed items (titles and publication dates) are listed on the page.
- **URL**: [http://localhost:8000/rss-feeds](http://localhost:8000/rss-feeds)
## Requirements

- **Go version**: 1.22 or above

## Setup

1. **Edit Params** (if any are required):
   Set the `port` and `outputFilePath` in `main.go` file as required. 
2. **Install dependencies** (if any are required):
   ```bash
   go mod tidy
   ```
3. **Run the program**:
   ```bash
   go run main.go
   ```

4. **Access the HTML output**:

   Once the program is running, visit `http://localhost:{port}/rss-feeds` in your browser to view the RSS feed summary.

## Usage

### Fetching RSS Feeds
- The program accepts a list of RSS feed URLs and filePaths and processes each feed concurrently using Goroutines.
- The program parses the `title` and `pubDate` from each `<item>` tag in the RSS feeds.
