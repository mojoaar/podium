package main

import (
	"flag"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"
	"github.com/russross/blackfriday/v2"
	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	SiteTitle       string `yaml:"site_title"`
	SiteDescription string `yaml:"site_description"`
	SiteAuthor      string `yaml:"site_author"`
	SiteURL         string `yaml:"site_url"`
	HomeIntro       string `yaml:"home_intro"`
	Port            int    `yaml:"port"`
	PostsFolder     string `yaml:"posts_folder"`
	StaticFolder    string `yaml:"static_folder"`
	TemplatesFolder string `yaml:"templates_folder"`
	AssetsFolder    string `yaml:"assets_folder"`
	PostsPerPage    int    `yaml:"posts_per_page"`
	FeedItems       int    `yaml:"feed_items"`
	ExcerptLength   int    `yaml:"excerpt_length"`
}

// Global config variable
var appConfig Config

// loadConfig reads and parses the config.yaml file
func loadConfig(path string) (Config, error) {
	var config Config
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	return config, err
}

type Page struct {
	Title       string
	Content     template.HTML
	Pages       []PageLink
	SiteTitle   string
	SiteDesc    string
	SiteAuthor  string
	IsDraft     bool
}

type PageLink struct {
	Title       string
	Slug        string
	Tags        []string
	Date        string
	Excerpt     string
	ReadingTime string
}

type Post struct {
	Title       string
	Slug        string
	Content     template.HTML
	Pages       []PageLink
	Tags        []string
	SiteTitle   string
	SiteDesc    string
	SiteAuthor  string
	Date        string
	IsDraft     bool
	ReadingTime string
}

// program implements the service.Interface
type program struct {
	router *gin.Engine
	exit   chan struct{}
}

func (p *program) Start(s service.Service) error {
	log.Println("Podium service starting...")
	go p.run()
	return nil
}

func (p *program) run() {
	// Set Gin to release mode when running as a service
	gin.SetMode(gin.ReleaseMode)
	
	p.router = gin.Default()

	// Load HTML templates
	p.router.LoadHTMLGlob("templates/*")

	// Home route
	p.router.GET("/", func(c *gin.Context) {
		pages := getStaticPages()
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Pages":       pages,
			"SiteTitle":   appConfig.SiteTitle,
			"SiteDesc":    appConfig.SiteDescription,
			"SiteAuthor":  appConfig.SiteAuthor,
			"HomeIntro":   appConfig.HomeIntro,
		})
	})

	// Static pages route
	p.router.GET("/page/:slug", func(c *gin.Context) {
		slug := c.Param("slug")
		content, title, _, _, isDraft, _, err := loadMarkdownFile("static", slug)
		if err != nil {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"Error":      "Page not found",
				"Pages":      getStaticPages(),
				"SiteTitle":  appConfig.SiteTitle,
			})
			return
		}

		// Don't show draft pages
		if isDraft {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"Error":      "Page not found",
				"Pages":      getStaticPages(),
				"SiteTitle":  appConfig.SiteTitle,
			})
			return
		}

		pages := getStaticPages()
		c.HTML(http.StatusOK, "page.html", Page{
			Title:      title,
			Content:    template.HTML(content),
			Pages:      pages,
			SiteTitle:  appConfig.SiteTitle,
			SiteDesc:   appConfig.SiteDescription,
			SiteAuthor: appConfig.SiteAuthor,
			IsDraft:    isDraft,
		})
	})

	// Blog posts list route
	p.router.GET("/posts", func(c *gin.Context) {
		allPosts := getBlogPosts()
		pages := getStaticPages()
		
		// Get page number from query params
		pageStr := c.DefaultQuery("page", "1")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
		
		// Calculate pagination
		postsPerPage := appConfig.PostsPerPage
		totalPosts := len(allPosts)
		totalPages := (totalPosts + postsPerPage - 1) / postsPerPage
		
		// Ensure page is within bounds
		if page > totalPages && totalPages > 0 {
			page = totalPages
		}
		
		// Calculate slice bounds
		start := (page - 1) * postsPerPage
		end := start + postsPerPage
		if end > totalPosts {
			end = totalPosts
		}
		
		// Get posts for current page
		var paginatedPosts []PageLink
		if start < totalPosts {
			paginatedPosts = allPosts[start:end]
		}
		
		c.HTML(http.StatusOK, "posts.html", gin.H{
			"Posts":       paginatedPosts,
			"Pages":       pages,
			"SiteTitle":   appConfig.SiteTitle,
			"CurrentPage": page,
			"TotalPages":  totalPages,
			"HasPrev":     page > 1,
			"HasNext":     page < totalPages,
			"PrevPage":    page - 1,
			"NextPage":    page + 1,
		})
	})

	// Individual blog post route
	p.router.GET("/posts/:slug", func(c *gin.Context) {
		slug := c.Param("slug")
		content, title, tags, date, isDraft, plainText, err := loadMarkdownFile("posts", slug)
		if err != nil {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"Error":     "Post not found",
				"Pages":     getStaticPages(),
				"SiteTitle": appConfig.SiteTitle,
			})
			return
		}

		// Don't show draft posts
		if isDraft {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"Error":     "Post not found",
				"Pages":     getStaticPages(),
				"SiteTitle": appConfig.SiteTitle,
			})
			return
		}

		pages := getStaticPages()
		readingTime := calculateReadingTime(plainText)
		c.HTML(http.StatusOK, "post.html", Post{
			Title:       title,
			Slug:        slug,
			Content:     template.HTML(content),
			Pages:       pages,
			Tags:        tags,
			SiteTitle:   appConfig.SiteTitle,
			SiteDesc:    appConfig.SiteDescription,
			SiteAuthor:  appConfig.SiteAuthor,
			Date:        date,
			IsDraft:     isDraft,
			ReadingTime: readingTime,
		})
	})

	// Tag filtering route
	p.router.GET("/tags/:tag", func(c *gin.Context) {
		tag := c.Param("tag")
		allPosts := getBlogPosts()
		var filteredPosts []PageLink
		
		for _, post := range allPosts {
			for _, postTag := range post.Tags {
				if strings.EqualFold(postTag, tag) {
					filteredPosts = append(filteredPosts, post)
					break
				}
			}
		}
		
		// Get page number from query params
		pageStr := c.DefaultQuery("page", "1")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
		
		// Calculate pagination
		postsPerPage := appConfig.PostsPerPage
		totalPosts := len(filteredPosts)
		totalPages := (totalPosts + postsPerPage - 1) / postsPerPage
		
		// Ensure page is within bounds
		if page > totalPages && totalPages > 0 {
			page = totalPages
		}
		
		// Calculate slice bounds
		start := (page - 1) * postsPerPage
		end := start + postsPerPage
		if end > totalPosts {
			end = totalPosts
		}
		
		// Get posts for current page
		var paginatedPosts []PageLink
		if start < totalPosts {
			paginatedPosts = filteredPosts[start:end]
		}
		
		pages := getStaticPages()
		c.HTML(http.StatusOK, "posts.html", gin.H{
			"Posts":       paginatedPosts,
			"Pages":       pages,
			"Tag":         tag,
			"SiteTitle":   appConfig.SiteTitle,
			"CurrentPage": page,
			"TotalPages":  totalPages,
			"HasPrev":     page > 1,
			"HasNext":     page < totalPages,
			"PrevPage":    page - 1,
			"NextPage":    page + 1,
		})
	})

	// RSS/Atom Feed route
	p.router.GET("/feed.xml", func(c *gin.Context) {
		posts := getBlogPosts()
		
		// Limit to feed_items from config
		feedPosts := posts
		if len(posts) > appConfig.FeedItems {
			feedPosts = posts[:appConfig.FeedItems]
		}
		
		// Build time for the feed (most recent post date or current time)
		buildDate := time.Now().Format(time.RFC1123Z)
		if len(feedPosts) > 0 && feedPosts[0].Date != "" {
			if parsedDate, err := time.Parse("2006-01-02", feedPosts[0].Date); err == nil {
				buildDate = parsedDate.Format(time.RFC1123Z)
			}
		}
		
		c.Header("Content-Type", "application/rss+xml; charset=utf-8")
		c.String(http.StatusOK, generateRSSFeed(feedPosts, buildDate))
	})

	// Sitemap.xml route
	p.router.GET("/sitemap.xml", func(c *gin.Context) {
		posts := getBlogPosts()
		pages := getStaticPages()
		
		c.Header("Content-Type", "application/xml; charset=utf-8")
		c.String(http.StatusOK, generateSitemap(posts, pages))
	})

	// Serve static assets (CSS, JS, images)
	p.router.Static("/assets", "./assets")

	port := fmt.Sprintf(":%d", appConfig.Port)
	log.Printf("Starting Podium server on %s", port)
	if err := p.router.Run(port); err != nil {
		log.Printf("Error starting server: %v", err)
	}
}

func (p *program) Stop(s service.Service) error {
	log.Println("Podium service stopping...")
	close(p.exit)
	return nil
}

func main() {
	var serviceAction string
	flag.StringVar(&serviceAction, "service", "", "Control the system service: install, uninstall, start, stop, restart")
	flag.Parse()

	// Get the executable path for the service
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	// Get the working directory (where the app files are)
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Service configuration
	svcConfig := &service.Config{
		Name:             "Podium",
		DisplayName:      "Podium Web Server",
		Description:      "A lightweight web server for hosting markdown-based websites and blogs",
		WorkingDirectory: workDir,
		Executable:       execPath,
		Option: service.KeyValue{
			"UserService": true, // Install as user service on macOS/Linux
		},
	}

	prg := &program{
		exit: make(chan struct{}),
	}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Handle service control actions
	if serviceAction != "" {
		err := handleServiceAction(s, serviceAction)
		if err != nil {
			log.Fatalf("Failed to %s service: %v", serviceAction, err)
		}
		return
	}

	// Create logger
	logger, err := s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Run the service
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}

func handleServiceAction(s service.Service, action string) error {
	switch action {
	case "install":
		err := s.Install()
		if err != nil {
			return err
		}
		fmt.Println("Service installed successfully!")
		fmt.Println("Use '-service start' to start the service")
		return nil
	case "uninstall":
		err := s.Stop()
		if err != nil {
			log.Printf("Warning: Failed to stop service: %v", err)
		}
		err = s.Uninstall()
		if err != nil {
			return err
		}
		fmt.Println("Service uninstalled successfully!")
		return nil
	case "start":
		err := s.Start()
		if err != nil {
			return err
		}
		fmt.Println("Service started successfully!")
		return nil
	case "stop":
		err := s.Stop()
		if err != nil {
			return err
		}
		fmt.Println("Service stopped successfully!")
		return nil
	case "restart":
		err := s.Restart()
		if err != nil {
			return err
		}
		fmt.Println("Service restarted successfully!")
		return nil
	default:
		return fmt.Errorf("unknown service action: %s", action)
	}
}

// loadMarkdownFile reads and converts a markdown file to HTML
func loadMarkdownFile(folder, slug string) (string, string, []string, string, bool, string, error) {
	filePath := filepath.Join(folder, slug+".md")
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", "", nil, "", false, "", err
	}

	// Parse front matter for metadata (tags, date, draft)
	lines := strings.Split(string(content), "\n")
	var tags []string
	title := slug
	var date string
	var isDraft bool
	var contentStartLine int
	var frontMatterLines []int
	
	// Check for front matter (tags, date, draft)
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
			if contentStartLine == 0 {
				contentStartLine = i
			}
			break
		}
		if strings.HasPrefix(line, "Tags:") || strings.HasPrefix(line, "tags:") {
			tagStr := strings.TrimPrefix(line, "Tags:")
			tagStr = strings.TrimPrefix(tagStr, "tags:")
			tagStr = strings.TrimSpace(tagStr)
			if tagStr != "" {
				tagList := strings.Split(tagStr, ",")
				for _, tag := range tagList {
					tags = append(tags, strings.TrimSpace(tag))
				}
			}
			frontMatterLines = append(frontMatterLines, i)
		}
		if strings.HasPrefix(line, "Date:") || strings.HasPrefix(line, "date:") {
			date = strings.TrimPrefix(line, "Date:")
			date = strings.TrimPrefix(date, "date:")
			date = strings.TrimSpace(date)
			frontMatterLines = append(frontMatterLines, i)
		}
		if strings.HasPrefix(line, "Draft:") || strings.HasPrefix(line, "draft:") {
			draftStr := strings.TrimPrefix(line, "Draft:")
			draftStr = strings.TrimPrefix(draftStr, "draft:")
			draftStr = strings.TrimSpace(draftStr)
			isDraft = strings.ToLower(draftStr) == "true"
			frontMatterLines = append(frontMatterLines, i)
		}
		// Don't process beyond the title
		if i > 15 {
			break
		}
	}

	// Remove front matter lines from content before converting to HTML
	var contentLines []string
	if len(frontMatterLines) > 0 && contentStartLine == 0 {
		// Find the last front matter line
		maxFrontMatter := 0
		for _, lineNum := range frontMatterLines {
			if lineNum > maxFrontMatter {
				maxFrontMatter = lineNum
			}
		}
		contentStartLine = maxFrontMatter + 1
	}
	
	if contentStartLine > 0 && contentStartLine < len(lines) {
		contentLines = lines[contentStartLine:]
	} else {
		contentLines = lines
	}

	contentToRender := strings.Join(contentLines, "\n")

	// Convert markdown to HTML
	html := blackfriday.Run([]byte(contentToRender))
	
	// Get plain text content for excerpts
	plainText := stripHTML(string(html))

	return string(html), title, tags, date, isDraft, plainText, nil
}

// generateRSSFeed creates an RSS 2.0 feed XML string
func generateRSSFeed(posts []PageLink, buildDate string) string {
	var feed strings.Builder
	
	feed.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	feed.WriteString("\n")
	feed.WriteString(`<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">`)
	feed.WriteString("\n<channel>\n")
	
	// Channel metadata
	feed.WriteString(fmt.Sprintf("  <title>%s</title>\n", htmlEscape(appConfig.SiteTitle)))
	feed.WriteString(fmt.Sprintf("  <link>%s</link>\n", appConfig.SiteURL))
	feed.WriteString(fmt.Sprintf("  <description>%s</description>\n", htmlEscape(appConfig.SiteDescription)))
	feed.WriteString("  <language>en-us</language>\n")
	feed.WriteString(fmt.Sprintf("  <lastBuildDate>%s</lastBuildDate>\n", buildDate))
	feed.WriteString(fmt.Sprintf("  <atom:link href=\"%s/feed.xml\" rel=\"self\" type=\"application/rss+xml\" />\n", appConfig.SiteURL))
	
	// Items
	for _, post := range posts {
		feed.WriteString("  <item>\n")
		feed.WriteString(fmt.Sprintf("    <title>%s</title>\n", htmlEscape(post.Title)))
		feed.WriteString(fmt.Sprintf("    <link>%s/posts/%s</link>\n", appConfig.SiteURL, post.Slug))
		feed.WriteString(fmt.Sprintf("    <guid>%s/posts/%s</guid>\n", appConfig.SiteURL, post.Slug))
		
		if post.Date != "" {
			if parsedDate, err := time.Parse("2006-01-02", post.Date); err == nil {
				feed.WriteString(fmt.Sprintf("    <pubDate>%s</pubDate>\n", parsedDate.Format(time.RFC1123Z)))
			}
		}
		
		// Load post content for description
		content, _, _, _, _, _, err := loadMarkdownFile("posts", post.Slug)
		if err == nil {
			// Truncate content for RSS description (first 200 chars)
			description := stripHTML(content)
			if len(description) > 200 {
				description = description[:200] + "..."
			}
			feed.WriteString(fmt.Sprintf("    <description>%s</description>\n", htmlEscape(description)))
		}
		
		// Add tags as categories
		for _, tag := range post.Tags {
			feed.WriteString(fmt.Sprintf("    <category>%s</category>\n", htmlEscape(tag)))
		}
		
		feed.WriteString("  </item>\n")
	}
	
	feed.WriteString("</channel>\n</rss>")
	return feed.String()
}

// htmlEscape escapes special characters for XML/HTML
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// generateSitemap creates an XML sitemap for all posts and pages
func generateSitemap(posts []PageLink, pages []PageLink) string {
	var sitemap strings.Builder
	
	sitemap.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sitemap.WriteString("\n")
	sitemap.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	sitemap.WriteString("\n")
	
	// Add homepage
	sitemap.WriteString("  <url>\n")
	sitemap.WriteString(fmt.Sprintf("    <loc>%s</loc>\n", appConfig.SiteURL))
	sitemap.WriteString("    <changefreq>daily</changefreq>\n")
	sitemap.WriteString("    <priority>1.0</priority>\n")
	sitemap.WriteString("  </url>\n")
	
	// Add posts list page
	sitemap.WriteString("  <url>\n")
	sitemap.WriteString(fmt.Sprintf("    <loc>%s/posts</loc>\n", appConfig.SiteURL))
	sitemap.WriteString("    <changefreq>daily</changefreq>\n")
	sitemap.WriteString("    <priority>0.9</priority>\n")
	sitemap.WriteString("  </url>\n")
	
	// Add individual blog posts
	for _, post := range posts {
		sitemap.WriteString("  <url>\n")
		sitemap.WriteString(fmt.Sprintf("    <loc>%s/posts/%s</loc>\n", appConfig.SiteURL, post.Slug))
		
		if post.Date != "" {
			if parsedDate, err := time.Parse("2006-01-02", post.Date); err == nil {
				sitemap.WriteString(fmt.Sprintf("    <lastmod>%s</lastmod>\n", parsedDate.Format("2006-01-02")))
			}
		}
		
		sitemap.WriteString("    <changefreq>monthly</changefreq>\n")
		sitemap.WriteString("    <priority>0.8</priority>\n")
		sitemap.WriteString("  </url>\n")
	}
	
	// Add static pages
	for _, page := range pages {
		sitemap.WriteString("  <url>\n")
		sitemap.WriteString(fmt.Sprintf("    <loc>%s/page/%s</loc>\n", appConfig.SiteURL, page.Slug))
		sitemap.WriteString("    <changefreq>monthly</changefreq>\n")
		sitemap.WriteString("    <priority>0.7</priority>\n")
		sitemap.WriteString("  </url>\n")
	}
	
	sitemap.WriteString("</urlset>\n")
	return sitemap.String()
}

// stripHTML removes HTML tags from a string (simple implementation)
func stripHTML(s string) string {
	// Simple regex-free approach: remove everything between < and >
	var result strings.Builder
	inTag := false
	for _, char := range s {
		if char == '<' {
			inTag = true
			continue
		}
		if char == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(char)
		}
	}
	// Decode HTML entities like &rsquo; &amp; etc.
	decoded := html.UnescapeString(result.String())
	return strings.TrimSpace(decoded)
}

// generateExcerpt creates a truncated excerpt from plain text
func generateExcerpt(text string, maxLength int) string {
	// Clean up whitespace
	text = strings.TrimSpace(text)
	text = strings.Join(strings.Fields(text), " ")
	
	if len(text) <= maxLength {
		return text
	}
	
	// Truncate and add ellipsis
	excerpt := text[:maxLength]
	
	// Try to break at the last space to avoid cutting words
	lastSpace := strings.LastIndex(excerpt, " ")
	if lastSpace > maxLength-50 { // Only use the space if it's not too far back
		excerpt = excerpt[:lastSpace]
	}
	
	return excerpt + "..."
}

// calculateReadingTime estimates reading time based on word count
// Average reading speed: 200-250 words per minute (using 225)
func calculateReadingTime(text string) string {
	words := len(strings.Fields(text))
	minutes := words / 225
	
	if minutes < 1 {
		return "< 1 min read"
	} else if minutes == 1 {
		return "1 min read"
	}
	
	return fmt.Sprintf("%d min read", minutes)
}

// getStaticPages scans the static folder and returns all available pages
func getStaticPages() []PageLink {
	var pages []PageLink

	files, err := ioutil.ReadDir("static")
	if err != nil {
		log.Printf("Error reading static folder: %v", err)
		return pages
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			slug := strings.TrimSuffix(file.Name(), ".md")
			
			// Read file to get title and draft status
			_, title, _, _, isDraft, _, err := loadMarkdownFile("static", slug)
			if err != nil {
				continue
			}

			// Skip draft pages
			if isDraft {
				continue
			}

			pages = append(pages, PageLink{
				Title: title,
				Slug:  slug,
			})
		}
	}

	return pages
}

// getBlogPosts scans the posts folder and returns all available posts
func getBlogPosts() []PageLink {
	var posts []PageLink

	files, err := ioutil.ReadDir("posts")
	if err != nil {
		log.Printf("Error reading posts folder: %v", err)
		return posts
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			slug := strings.TrimSuffix(file.Name(), ".md")
			
			// Read file to get title, tags, date, draft status, and content for excerpt
			_, title, tags, date, isDraft, plainText, err := loadMarkdownFile("posts", slug)
			if err != nil {
				continue
			}

			// Skip draft posts
			if isDraft {
				continue
			}
			
			// Generate excerpt and reading time
			excerpt := generateExcerpt(plainText, appConfig.ExcerptLength)
			readingTime := calculateReadingTime(plainText)

			posts = append(posts, PageLink{
				Title:       title,
				Slug:        slug,
				Tags:        tags,
				Date:        date,
				Excerpt:     excerpt,
				ReadingTime: readingTime,
			})
		}
	}

	// Sort posts by date (newest first)
	sort.Slice(posts, func(i, j int) bool {
		dateI, errI := time.Parse("2006-01-02", posts[i].Date)
		dateJ, errJ := time.Parse("2006-01-02", posts[j].Date)
		
		// If either date is invalid, put it at the end
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}
		
		return dateI.After(dateJ)
	})

	return posts
}

// createFoldersIfNotExist creates necessary folders on startup
func init() {
	// Load config first
	var err error
	appConfig, err = loadConfig("config.yaml")
	if err != nil {
		// Use defaults if config file is missing
		log.Printf("Warning: Could not load config.yaml, using defaults: %v", err)
		appConfig = Config{
			SiteTitle:       "Podium",
			SiteDescription: "A simple and elegant blogging platform",
			SiteAuthor:      "Morten Johansen",
			SiteURL:         "http://localhost:8080",
			Port:            8080,
			PostsFolder:     "posts",
			StaticFolder:    "static",
			TemplatesFolder: "templates",
			AssetsFolder:    "assets",
			PostsPerPage:    10,
			FeedItems:       20,
		}
	}

	// Set defaults for optional fields if not provided
	if appConfig.PostsPerPage == 0 {
		appConfig.PostsPerPage = 10
	}
	if appConfig.FeedItems == 0 {
		appConfig.FeedItems = 20
	}
	if appConfig.SiteURL == "" {
		appConfig.SiteURL = fmt.Sprintf("http://localhost:%d", appConfig.Port)
	}

	folders := []string{appConfig.StaticFolder, appConfig.PostsFolder, appConfig.TemplatesFolder, appConfig.AssetsFolder}
	for _, folder := range folders {
		if _, err := os.Stat(folder); os.IsNotExist(err) {
			err := os.MkdirAll(folder, 0755)
			if err != nil {
				log.Fatalf("Failed to create %s folder: %v", folder, err)
			}
		}
	}
}
