# Podium Features

Comprehensive list of all features in Podium.

## Core Features

### 🚀 **Go & Gin Framework**

- Built with Go 1.21+ for performance and reliability
- Gin web framework v1.11.0 for fast HTTP routing
- Minimal dependencies, lightweight binary

### 📝 **Markdown Support**

- Write blog posts in Markdown
- Create static pages in Markdown
- Powered by Blackfriday v2.1.0 parser
- Full CommonMark support

### 🔄 **Auto-Discovery**

- Static pages automatically appear in navigation
- New posts automatically added to lists
- No manual configuration needed

## Content Management

### 📅 **Date & Timestamps**

- Add publication dates to posts (`Date: 2025-11-03`)
- Automatic newest-first sorting
- Dates displayed on posts and lists
- ISO 8601 date format (YYYY-MM-DD)

### 🏷️ **Tags System**

- Add tags to posts (`Tags: golang, tutorial`)
- Tag-based filtering (`/tags/:tag`)
- Tags displayed on posts
- Clickable tags for easy navigation
- Pagination support for tag views

### 📄 **Draft Support**

- Mark posts as drafts (`Draft: true`)
- Mark pages as drafts
- Drafts hidden from public view
- Perfect for work-in-progress content

### ⭐ **Featured Posts**

- Pin important posts to the top (`Featured: true`)
- Special ⭐ Featured badge displayed
- Enhanced visual styling with gradient background
- Golden accent color and shadow
- Featured posts sorted by date among themselves
- Perfect for announcements, popular content, or highlights
- Appears before regular posts in listings

### 📆 **Post Scheduling**

- Schedule posts for future publication
- Format: `PublishDate: 2025-12-01 09:00` (YYYY-MM-DD HH:MM)
- Posts hidden until publish date/time
- Automatic visibility once scheduled time arrives
- No server restart required
- Hidden from:
  - Blog posts list
  - Individual post pages
  - Tag filtering
  - RSS feed
  - Sitemap
- Perfect for timed announcements and content planning

### 🔖 **Post Excerpts**

- Automatic excerpt generation
- Configurable length (`excerpt_length: 200`)
- Displayed on posts list pages
- HTML stripped and entities decoded
- Smart word-boundary truncation

### ⏱️ **Reading Time**

- Automatic reading time calculation
- Based on ~225 words per minute
- Displayed on posts and lists
- Format: "< 1 min read", "1 min read", "X min read"

## Design & User Experience

### 🎨 **Theme Toggle**

- Dark and light themes
- Manual toggle in navigation
- Persists via localStorage
- CSS variables for easy customization
- Smooth transitions

### 📱 **Mobile Responsive**

- Responsive breakpoints at 768px and 480px
- Touch-friendly tap targets (44px minimum)
- Mobile-optimized navigation
- Flexible layouts
- Properly sized typography
- Scrollable code blocks

### 🖨️ **Print-Friendly**

- Clean print styles
- Hides navigation and footer
- Shows external link URLs
- Optimized typography for print
- Proper page breaks
- Black and white output

### 💻 **Syntax Highlighting**

- Highlight.js v11.9.0 integration
- 140+ languages supported
- GitHub Dark theme
- Automatic code detection
- Line highlighting support

### 🎯 **Clean UI**

- Modern, minimal design
- Consistent spacing and typography
- Hover effects and transitions
- Professional color scheme
- Card-based layouts

## Social & Discovery

### 🔗 **Share Buttons**

- Twitter (X) sharing
- LinkedIn sharing
- Facebook sharing
- Reddit sharing
- Copy link to clipboard
- Branded platform colors
- Touch-friendly buttons
- Hidden when printing

### 🌐 **Social Media Icons**

- Configurable social media links in footer
- Support for Twitter, Bluesky, LinkedIn, GitHub, Reddit, Facebook
- SVG icons with hover effects
- Only displays icons for configured platforms
- Toggle visibility with `show_social_links` config
- Appears on all pages (homepage, posts, pages, error)
- Responsive design with proper spacing
- Accessible with ARIA labels and titles

### � **Umami Analytics**

- Privacy-focused, open-source analytics
- Optional integration via configuration
- Automatically injected when configured
- Simply add `umami_script_url` and `umami_website_id` to config
- Works across all pages (homepage, posts, pages, error)
- No tracking if not configured
- Supports self-hosted or cloud Umami instances

### �📡 **RSS/Atom Feed**

- Available at `/feed.xml`
- Configurable item count (`feed_items: 20`)
- RSS 2.0 with Atom namespace
- Auto-discovery link in HTML
- Full post descriptions
- Tags as categories
- Publication dates

### 🗺️ **Sitemap.xml**

- Available at `/sitemap.xml`
- All posts and pages included
- Homepage and posts list
- Last modification dates
- Priority values
- Standard sitemap.org schema
- SEO optimized

## Navigation & Organization

### 📄 **Pagination**

- Configurable posts per page (`posts_per_page: 10`)
- Previous/Next buttons
- Page indicator
- Works with tag filtering
- Query parameter based (`?page=N`)

### 🧭 **Routes**

- `/` - Home page
- `/posts` - Posts list (paginated)
- `/posts/:slug` - Individual post
- `/page/:slug` - Static page
- `/tags/:tag` - Tag filter (paginated)
- `/feed.xml` - RSS feed
- `/sitemap.xml` - Sitemap
- `/assets/*` - Static assets

## Configuration

### ⚙️ **YAML Configuration**

Complete YAML-based configuration in `config.yaml`:

```yaml
# Site metadata
site_title: "Podium"
site_description: "A simple and elegant blogging platform"
site_author: "Morten Johansen"
site_author_url: "https://johansen.foo"
site_url: "http://localhost:8080"
home_intro: "Podium is a lightweight web application built with Go and the Gin framework. It supports markdown-based blog posts and static pages that automatically appear in the navigation when added."
show_quick_links: true

# Server
port: 8080

# Pagination
posts_per_page: 10

# RSS Feed
feed_items: 20

# Excerpts
excerpt_length: 200

# Social Media Links
show_social_links: true
social_twitter: ""
social_bluesky: ""
social_linkedin: ""
social_github: ""
social_reddit: ""
social_facebook: ""

# Analytics
umami_script_url: "" # e.g., "https://analytics.yourdomain.com/script.js"
umami_website_id: "" # Your Umami website ID

# Paths
posts_folder: "posts"
static_folder: "static"
templates_folder: "templates"
assets_folder: "assets"
```

**Configuration Options:**

- `site_title` - Website name (navigation & titles)
- `site_description` - Site description (meta tags & homepage)
- `site_author` - Author name (footer)
- `site_author_url` - Optional URL to link your name in the footer (e.g., personal website)
- `site_url` - Full site URL (RSS & sitemap)
- `home_intro` - Homepage introduction text (About section)
- `show_quick_links` - Toggle Quick Links section on homepage (true/false)
- `port` - Server port (default: 8080)
- `posts_per_page` - Posts per page (default: 10)
- `feed_items` - RSS feed items (default: 20)
- `excerpt_length` - Excerpt character limit (default: 200)
- `show_social_links` - Toggle social media icons in footer (true/false, default: false)
- `social_twitter` - Twitter/X profile URL
- `social_bluesky` - Bluesky profile URL
- `social_linkedin` - LinkedIn profile URL
- `social_github` - GitHub profile URL
- `social_reddit` - Reddit profile URL
- `social_facebook` - Facebook profile URL
- `umami_script_url` - Umami analytics script URL (optional)
- `umami_website_id` - Your Umami website ID for tracking (optional)
- Folder paths for content, templates, and assets

### 🔧 **Fallback Defaults**

- Works without config.yaml
- Sensible defaults for all settings
- Easy to customize

### 🔥 **Config Hot Reload**

- Automatic config reload on file changes
- Works in both development and production modes
- No server restart required
- ~500ms reload delay (debounced)
- Perfect for updating social links, site settings, etc.
- Works even when running as a system service

## System Integration

### 🔧 **Cross-Platform Service**

- Windows Service support
- macOS LaunchAgent/LaunchDaemon
- Linux systemd support
- Install/uninstall commands
- Start/stop/restart commands
- Auto-start on boot
- Powered by kardianos/service

### 🏗️ **Build System**

- Makefile for easy building
- Cross-platform build scripts (build.sh, build.bat)
- Build for all platforms with one command
- Organized bin/ directory
- Support for:
  - Linux (amd64, arm64)
  - macOS (amd64, arm64)
  - Windows (amd64)

## Technical Features

### ⚡ **Performance**

- Fast Go compilation
- Minimal runtime overhead
- Efficient markdown parsing
- Static asset serving
- No database required
- File-based content
- **HTTP caching with ETag support**
  - Static assets cached for 7-30 days depending on type
  - Automatic cache invalidation based on file modification
  - 304 Not Modified responses for unchanged content
  - Configurable cache headers per content type
- **Lazy-loaded images**
  - All images automatically include `loading="lazy"` attribute
  - Deferred loading for off-screen images
  - Improved page load performance
  - Responsive image sizing with CSS
- **Asset minification**
  - Automatic CSS and JavaScript minification in production
  - 30-60% file size reduction
  - Disabled in dev mode for easier debugging
  - On-the-fly minification using tdewolff/minify
- **Image optimization**
  - Dynamic image resizing via query parameters (`?w=800&h=600`)
  - Quality optimization for JPEG and PNG images
  - Automatic max-width limiting (1920px for large images)
  - On-demand processing, no pre-generation needed
- **Automatic sitemap generation**
  - `/sitemap.xml` automatically generated
  - Includes all posts with lastmod dates
  - Includes all static pages
  - Proper priority and changefreq values
  - SEO-optimized for search engines

### 🔒 **Security**

- HTML entity escaping
- XSS protection in templates
- Safe markdown rendering
- Proper HTTP headers
- No user authentication (by design)

### 🎨 **Assets**

- Custom CSS with CSS variables
- JavaScript for interactivity
- SVG favicon support
- Theme toggle script
- Share buttons script
- Static asset serving

## Developer Features

### 📝 **Templates**

- Customizable HTML templates
- Go template syntax
- Template inheritance
- Separate templates for:
  - Home page
  - Posts list
  - Individual post
  - Static page
  - Error page

### 🎨 **Styling**

- CSS with variables
- Dark/light theme support
- Responsive design
- Print styles
- Modular structure

### 🧪 **Development**

- Easy local development (`go run main.go`)
- **Hot reload mode** (`go run main.go -dev`)
  - Automatic file watching with fsnotify
  - Reloads templates on file changes
  - Reloads config.yaml dynamically
  - Watches posts, static, templates, assets directories
  - 500ms debounce to prevent multiple reloads
  - Debug mode logging for troubleshooting
- **Config hot reload in production**
  - Watches config.yaml even in production mode
  - Configuration changes apply instantly
  - No server restart needed
  - Ideal for updating social links, site settings, etc.
- Live reload friendly
- Clear error messages
- Comprehensive logging

## File Support

### 📁 **Folder Structure**

```
podium/
├── posts/          # Blog posts (.md files)
├── static/         # Static pages (.md files)
├── templates/      # HTML templates
├── assets/         # CSS, JS, images
└── bin/           # Compiled binaries
```

### 📄 **Front Matter**

Supported in all markdown files:

- `Tags: tag1, tag2, tag3` - Post tags
- `Date: 2025-11-03` - Publication date
- `Draft: true` - Draft status

### 🎯 **Content Types**

- **Blog Posts**: Dated, tagged, with excerpts and reading time
- **Static Pages**: Timeless content like About, Contact
- Both support full Markdown syntax

## Accessibility

### ♿ **A11y Features**

- Semantic HTML
- ARIA labels on share buttons
- Keyboard navigation support
- Sufficient color contrast
- Touch-friendly tap targets
- Responsive font sizing

## SEO Features

- **Automatic sitemap.xml generation**
  - Dynamically generated at `/sitemap.xml`
  - Includes all posts with publication dates as lastmod
  - Includes all static pages
  - Homepage, posts page, and RSS feed included
  - Proper priority values (1.0 for homepage, 0.9 for posts page, 0.8 for individual posts)
  - Change frequency hints for crawlers
- **robots.txt** with sitemap reference
- **humans.txt** for project credits
- RSS/Atom feed with auto-discovery
- Meta descriptions from site configuration
- Semantic HTML structure
- Clean, readable URLs
- Mobile-friendly design
- Fast load times with optimization
- Proper heading hierarchy
- Image alt text support

## License

MIT License - Open source and free to use!
