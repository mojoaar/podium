package main

import (
	"flag"
	"fmt"
	"html"
	"html/template"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"
	"github.com/russross/blackfriday/v2"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	SiteTitle       string `yaml:"site_title"`
	SiteDescription string `yaml:"site_description"`
	SiteAuthor      string `yaml:"site_author"`
	SiteAuthorURL   string `yaml:"site_author_url"`
	SiteURL         string `yaml:"site_url"`
	HomeIntro       string `yaml:"home_intro"`
	ShowQuickLinks  bool   `yaml:"show_quick_links"`
	Port            int    `yaml:"port"`
	PostsFolder     string `yaml:"posts_folder"`
	StaticFolder    string `yaml:"static_folder"`
	TemplatesFolder string `yaml:"templates_folder"`
	AssetsFolder    string `yaml:"assets_folder"`
	PostsPerPage    int    `yaml:"posts_per_page"`
	FeedItems       int    `yaml:"feed_items"`
	ExcerptLength   int    `yaml:"excerpt_length"`
	ShowSocialLinks bool   `yaml:"show_social_links"`
	SocialTwitter   string `yaml:"social_twitter"`
	SocialBluesky   string `yaml:"social_bluesky"`
	SocialLinkedIn  string `yaml:"social_linkedin"`
	SocialGitHub    string `yaml:"social_github"`
	SocialReddit    string `yaml:"social_reddit"`
	SocialFacebook  string `yaml:"social_facebook"`
	UmamiScriptURL  string `yaml:"umami_script_url"`
	UmamiWebsiteID  string `yaml:"umami_website_id"`
}

// Global config variable
var appConfig Config
var isDevMode bool

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
	Title            string
	Content          template.HTML
	Pages            []PageLink
	SiteTitle        string
	SiteDesc         string
	SiteAuthor       string
	SiteAuthorURL    string
	IsDraft          bool
	CurrentYear      string
	ShowSocialLinks  bool
	SocialTwitter    string
	SocialBluesky    string
	SocialLinkedIn   string
	SocialGitHub     string
	SocialReddit     string
	SocialFacebook   string
	UmamiScriptURL   string
	UmamiWebsiteID   string
}

type PageLink struct {
	Title       string
	Slug        string
	Tags        []string
	Date        string
	PublishDate string
	Excerpt     string
	ReadingTime string
	Featured    bool
}

type Post struct {
	Title            string
	Slug             string
	Content          template.HTML
	Pages            []PageLink
	Tags             []string
	SiteTitle        string
	SiteDesc         string
	SiteAuthor       string
	SiteAuthorURL    string
	Date             string
	PublishDate      string
	IsDraft          bool
	ReadingTime      string
	CurrentYear      string
	Featured         bool
	ShowSocialLinks  bool
	SocialTwitter    string
	SocialBluesky    string
	SocialLinkedIn   string
	SocialGitHub     string
	SocialReddit     string
	SocialFacebook   string
	UmamiScriptURL   string
	UmamiWebsiteID   string
}

// program implements the service.Interface
type program struct {
	router *gin.Engine
	exit   chan struct{}
}

// cacheMiddleware adds appropriate caching headers based on content type
func cacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		
		// In dev mode, disable caching
		if isDevMode {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Next()
			return
		}
		
		// Cache static assets for longer periods
		if strings.HasPrefix(path, "/assets/") {
			// Determine cache duration based on file type
			ext := filepath.Ext(path)
			var maxAge int
			
			switch ext {
			case ".css", ".js":
				maxAge = 86400 * 7 // 7 days for CSS/JS
			case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".webp":
				maxAge = 86400 * 30 // 30 days for images
			case ".woff", ".woff2", ".ttf", ".eot":
				maxAge = 86400 * 365 // 1 year for fonts
			default:
				maxAge = 86400 // 1 day for other assets
			}
			
			c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
			
			// Try to get file modification time for ETag
			filePath := filepath.Join(".", path)
			if info, err := os.Stat(filePath); err == nil {
				etag := fmt.Sprintf("\"%x-%x\"", info.ModTime().Unix(), info.Size())
				c.Header("ETag", etag)
				
				// Check if client has cached version
				if match := c.GetHeader("If-None-Match"); match == etag {
					c.AbortWithStatus(http.StatusNotModified)
					return
				}
			}
		} else {
			// For HTML pages, use shorter cache with revalidation
			c.Header("Cache-Control", "public, max-age=300, must-revalidate")
		}
		
		c.Next()
	}
}

// minifyAsset minifies CSS and JS files on-the-fly
func minifyAsset(filePath string) ([]byte, error) {
	// Read the original file
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	// In dev mode, return unminified
	if isDevMode {
		return content, nil
	}
	
	// Create minifier
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("application/javascript", js.Minify)
	
	// Determine content type based on extension
	ext := filepath.Ext(filePath)
	var contentType string
	switch ext {
	case ".css":
		contentType = "text/css"
	case ".js":
		contentType = "application/javascript"
	default:
		// Don't minify other file types
		return content, nil
	}
	
	// Minify the content
	minified, err := m.Bytes(contentType, content)
	if err != nil {
		// If minification fails, return original content
		log.Printf("Warning: Failed to minify %s: %v", filePath, err)
		return content, nil
	}
	
	return minified, nil
}

// convertToWebP converts an image to WebP format
func convertToWebP(sourcePath string) ([]byte, error) {
	// Open the source image
	img, err := imaging.Open(sourcePath)
	if err != nil {
		return nil, err
	}
	
	// Create a buffer to write WebP data
	// Note: imaging library doesn't support WebP encoding directly
	// So we'll return the optimized version of the original format
	// For true WebP support, you'd need to use a library like 'chai2010/webp'
	
	// For now, let's just optimize the image by resizing if it's too large
	bounds := img.Bounds()
	maxWidth := 1920
	maxHeight := 1920
	
	if bounds.Dx() > maxWidth || bounds.Dy() > maxHeight {
		img = imaging.Fit(img, maxWidth, maxHeight, imaging.Lanczos)
	}
	
	// Encode back to original format with quality optimization
	ext := strings.ToLower(filepath.Ext(sourcePath))
	var buf []byte
	
	switch ext {
	case ".jpg", ".jpeg":
		// Save as JPEG with 85% quality
		tmpFile, err := ioutil.TempFile("", "optimized-*.jpg")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()
		
		if err := imaging.Save(img, tmpFile.Name(), imaging.JPEGQuality(85)); err != nil {
			return nil, err
		}
		
		buf, err = ioutil.ReadFile(tmpFile.Name())
		if err != nil {
			return nil, err
		}
		
	case ".png":
		tmpFile, err := ioutil.TempFile("", "optimized-*.png")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()
		
		if err := imaging.Save(img, tmpFile.Name()); err != nil {
			return nil, err
		}
		
		buf, err = ioutil.ReadFile(tmpFile.Name())
		if err != nil {
			return nil, err
		}
		
	default:
		// For unsupported formats, return original
		return ioutil.ReadFile(sourcePath)
	}
	
	return buf, nil
}

// resizeImage resizes an image to specified dimensions
func resizeImage(sourcePath string, width, height int) (image.Image, error) {
	img, err := imaging.Open(sourcePath)
	if err != nil {
		return nil, err
	}
	
	if width > 0 && height > 0 {
		// Resize to exact dimensions
		return imaging.Resize(img, width, height, imaging.Lanczos), nil
	} else if width > 0 {
		// Resize by width, maintain aspect ratio
		return imaging.Resize(img, width, 0, imaging.Lanczos), nil
	} else if height > 0 {
		// Resize by height, maintain aspect ratio
		return imaging.Resize(img, 0, height, imaging.Lanczos), nil
	}
	
	return img, nil
}

func (p *program) Start(s service.Service) error {
	log.Println("Podium service starting...")
	go p.run()
	
	// Start config file watcher in production mode
	if !isDevMode {
		go p.watchConfigFile()
	}
	
	return nil
}

func (p *program) run() {
	// Set Gin mode based on dev mode
	if isDevMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	
	p.router = gin.Default()

	// Load HTML templates (they will auto-reload in debug mode)
	p.router.LoadHTMLGlob("templates/*")

	// Add caching middleware
	p.router.Use(cacheMiddleware())

	// Home route
	p.router.GET("/", func(c *gin.Context) {
		pages := getStaticPages()
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Pages":          pages,
			"SiteTitle":      appConfig.SiteTitle,
			"SiteDesc":       appConfig.SiteDescription,
			"SiteAuthor":     appConfig.SiteAuthor,
			"SiteAuthorURL":  appConfig.SiteAuthorURL,
			"HomeIntro":      appConfig.HomeIntro,
			"ShowQuickLinks": appConfig.ShowQuickLinks,
			"CurrentYear":    getCurrentYear(),
			"ShowSocialLinks": appConfig.ShowSocialLinks,
			"SocialTwitter":   appConfig.SocialTwitter,
			"SocialBluesky":   appConfig.SocialBluesky,
			"SocialLinkedIn":  appConfig.SocialLinkedIn,
			"SocialGitHub":    appConfig.SocialGitHub,
			"SocialReddit":    appConfig.SocialReddit,
			"SocialFacebook":  appConfig.SocialFacebook,
			"UmamiScriptURL":  appConfig.UmamiScriptURL,
			"UmamiWebsiteID":  appConfig.UmamiWebsiteID,
		})
	})

	// Static pages route
	p.router.GET("/page/:slug", func(c *gin.Context) {
		slug := c.Param("slug")
		content, title, _, _, _, isDraft, _, _, err := loadMarkdownFile("static", slug)
		if err != nil {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"Error":           "Page not found",
				"Pages":           getStaticPages(),
				"SiteTitle":       appConfig.SiteTitle,
				"SiteAuthor":      appConfig.SiteAuthor,
				"SiteAuthorURL":   appConfig.SiteAuthorURL,
				"CurrentYear":     getCurrentYear(),
				"ShowSocialLinks": appConfig.ShowSocialLinks,
				"SocialTwitter":   appConfig.SocialTwitter,
				"SocialBluesky":   appConfig.SocialBluesky,
				"SocialLinkedIn":  appConfig.SocialLinkedIn,
				"SocialGitHub":    appConfig.SocialGitHub,
				"SocialReddit":    appConfig.SocialReddit,
				"SocialFacebook":  appConfig.SocialFacebook,
				"UmamiScriptURL":  appConfig.UmamiScriptURL,
				"UmamiWebsiteID":  appConfig.UmamiWebsiteID,
			})
			return
		}

		// Don't show draft pages
		if isDraft {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"Error":           "Page not found",
				"Pages":           getStaticPages(),
				"SiteTitle":       appConfig.SiteTitle,
				"SiteAuthor":      appConfig.SiteAuthor,
				"SiteAuthorURL":   appConfig.SiteAuthorURL,
				"CurrentYear":     getCurrentYear(),
				"ShowSocialLinks": appConfig.ShowSocialLinks,
				"SocialTwitter":   appConfig.SocialTwitter,
				"SocialBluesky":   appConfig.SocialBluesky,
				"SocialLinkedIn":  appConfig.SocialLinkedIn,
				"SocialGitHub":    appConfig.SocialGitHub,
				"SocialReddit":    appConfig.SocialReddit,
				"SocialFacebook":  appConfig.SocialFacebook,
				"UmamiScriptURL":  appConfig.UmamiScriptURL,
				"UmamiWebsiteID":  appConfig.UmamiWebsiteID,
			})
			return
		}

		pages := getStaticPages()
		c.HTML(http.StatusOK, "page.html", Page{
			Title:           title,
			Content:         template.HTML(content),
			Pages:           pages,
			SiteTitle:       appConfig.SiteTitle,
			SiteDesc:        appConfig.SiteDescription,
			SiteAuthor:      appConfig.SiteAuthor,
			SiteAuthorURL:   appConfig.SiteAuthorURL,
			IsDraft:         isDraft,
			CurrentYear:     getCurrentYear(),
			ShowSocialLinks: appConfig.ShowSocialLinks,
			SocialTwitter:   appConfig.SocialTwitter,
			SocialBluesky:   appConfig.SocialBluesky,
			SocialLinkedIn:  appConfig.SocialLinkedIn,
			SocialGitHub:    appConfig.SocialGitHub,
			SocialReddit:    appConfig.SocialReddit,
			SocialFacebook:  appConfig.SocialFacebook,
			UmamiScriptURL:  appConfig.UmamiScriptURL,
			UmamiWebsiteID:  appConfig.UmamiWebsiteID,
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
			"Posts":           paginatedPosts,
			"Pages":           pages,
			"SiteTitle":       appConfig.SiteTitle,
			"SiteAuthor":      appConfig.SiteAuthor,
			"SiteAuthorURL":   appConfig.SiteAuthorURL,
			"CurrentPage":     page,
			"TotalPages":      totalPages,
			"HasPrev":         page > 1,
			"HasNext":         page < totalPages,
			"PrevPage":        page - 1,
			"NextPage":        page + 1,
			"CurrentYear":     getCurrentYear(),
			"ShowSocialLinks": appConfig.ShowSocialLinks,
			"SocialTwitter":   appConfig.SocialTwitter,
			"SocialBluesky":   appConfig.SocialBluesky,
			"SocialLinkedIn":  appConfig.SocialLinkedIn,
			"SocialGitHub":    appConfig.SocialGitHub,
			"SocialReddit":    appConfig.SocialReddit,
			"SocialFacebook":  appConfig.SocialFacebook,
			"UmamiScriptURL":  appConfig.UmamiScriptURL,
			"UmamiWebsiteID":  appConfig.UmamiWebsiteID,
		})
	})

	// Individual blog post route
	p.router.GET("/posts/:slug", func(c *gin.Context) {
		slug := c.Param("slug")
		content, title, tags, date, publishDate, isDraft, isFeatured, plainText, err := loadMarkdownFile("posts", slug)
		if err != nil {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"Error":           "Post not found",
				"Pages":           getStaticPages(),
				"SiteTitle":       appConfig.SiteTitle,
				"SiteAuthor":      appConfig.SiteAuthor,
				"SiteAuthorURL":   appConfig.SiteAuthorURL,
				"CurrentYear":     getCurrentYear(),
				"ShowSocialLinks": appConfig.ShowSocialLinks,
				"SocialTwitter":   appConfig.SocialTwitter,
				"SocialBluesky":   appConfig.SocialBluesky,
				"SocialLinkedIn":  appConfig.SocialLinkedIn,
				"SocialGitHub":    appConfig.SocialGitHub,
				"SocialReddit":    appConfig.SocialReddit,
				"SocialFacebook":  appConfig.SocialFacebook,
				"UmamiScriptURL":  appConfig.UmamiScriptURL,
				"UmamiWebsiteID":  appConfig.UmamiWebsiteID,
			})
			return
		}

		// Don't show draft posts
		if isDraft {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"Error":           "Post not found",
				"Pages":           getStaticPages(),
				"SiteTitle":       appConfig.SiteTitle,
				"SiteAuthor":      appConfig.SiteAuthor,
				"SiteAuthorURL":   appConfig.SiteAuthorURL,
				"CurrentYear":     getCurrentYear(),
				"ShowSocialLinks": appConfig.ShowSocialLinks,
				"SocialTwitter":   appConfig.SocialTwitter,
				"SocialBluesky":   appConfig.SocialBluesky,
				"SocialLinkedIn":  appConfig.SocialLinkedIn,
				"SocialGitHub":    appConfig.SocialGitHub,
				"SocialReddit":    appConfig.SocialReddit,
				"SocialFacebook":  appConfig.SocialFacebook,
				"UmamiScriptURL":  appConfig.UmamiScriptURL,
				"UmamiWebsiteID":  appConfig.UmamiWebsiteID,
			})
			return
		}

		// Check if post is scheduled for future publication
		if publishDate != "" {
			pubTime, err := time.Parse("2006-01-02 15:04", publishDate)
			if err == nil && time.Now().Before(pubTime) {
				// Post is scheduled for the future, don't show it yet
				c.HTML(http.StatusNotFound, "error.html", gin.H{
					"Error":           "Post not found",
					"Pages":           getStaticPages(),
					"SiteTitle":       appConfig.SiteTitle,
					"SiteAuthor":      appConfig.SiteAuthor,
					"SiteAuthorURL":   appConfig.SiteAuthorURL,
					"CurrentYear":     getCurrentYear(),
					"ShowSocialLinks": appConfig.ShowSocialLinks,
					"SocialTwitter":   appConfig.SocialTwitter,
					"SocialBluesky":   appConfig.SocialBluesky,
					"SocialLinkedIn":  appConfig.SocialLinkedIn,
					"SocialGitHub":    appConfig.SocialGitHub,
					"SocialReddit":    appConfig.SocialReddit,
					"SocialFacebook":  appConfig.SocialFacebook,
					"UmamiScriptURL":  appConfig.UmamiScriptURL,
					"UmamiWebsiteID":  appConfig.UmamiWebsiteID,
				})
				return
			}
		}

		pages := getStaticPages()
		readingTime := calculateReadingTime(plainText)
		c.HTML(http.StatusOK, "post.html", Post{
			Title:           title,
			Slug:            slug,
			Content:         template.HTML(content),
			Pages:           pages,
			Tags:            tags,
			SiteTitle:       appConfig.SiteTitle,
			SiteDesc:        appConfig.SiteDescription,
			SiteAuthor:      appConfig.SiteAuthor,
			SiteAuthorURL:   appConfig.SiteAuthorURL,
			Date:            date,
			PublishDate:     publishDate,
			IsDraft:         isDraft,
			ReadingTime:     readingTime,
			CurrentYear:     getCurrentYear(),
			Featured:        isFeatured,
			ShowSocialLinks: appConfig.ShowSocialLinks,
			SocialTwitter:   appConfig.SocialTwitter,
			SocialBluesky:   appConfig.SocialBluesky,
			SocialLinkedIn:  appConfig.SocialLinkedIn,
			SocialGitHub:    appConfig.SocialGitHub,
			SocialReddit:    appConfig.SocialReddit,
			SocialFacebook:  appConfig.SocialFacebook,
			UmamiScriptURL:  appConfig.UmamiScriptURL,
			UmamiWebsiteID:  appConfig.UmamiWebsiteID,
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
			"Posts":           paginatedPosts,
			"Pages":           pages,
			"Tag":             tag,
			"SiteTitle":       appConfig.SiteTitle,
			"SiteAuthor":      appConfig.SiteAuthor,
			"SiteAuthorURL":   appConfig.SiteAuthorURL,
			"CurrentPage":     page,
			"TotalPages":      totalPages,
			"HasPrev":         page > 1,
			"HasNext":         page < totalPages,
			"PrevPage":        page - 1,
			"NextPage":        page + 1,
			"CurrentYear":     getCurrentYear(),
			"ShowSocialLinks": appConfig.ShowSocialLinks,
			"SocialTwitter":   appConfig.SocialTwitter,
			"SocialBluesky":   appConfig.SocialBluesky,
			"SocialLinkedIn":  appConfig.SocialLinkedIn,
			"SocialGitHub":    appConfig.SocialGitHub,
			"SocialReddit":    appConfig.SocialReddit,
			"SocialFacebook":  appConfig.SocialFacebook,
			"UmamiScriptURL":  appConfig.UmamiScriptURL,
			"UmamiWebsiteID":  appConfig.UmamiWebsiteID,
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

	// Serve robots.txt
	p.router.GET("/robots.txt", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		content := fmt.Sprintf("# robots.txt for Podium\n# https://www.robotstxt.org/\n\nUser-agent: *\nAllow: /\n\n# Sitemaps\nSitemap: %s/sitemap.xml\n", appConfig.SiteURL)
		c.String(http.StatusOK, content)
	})

	// Serve humans.txt
	p.router.GET("/humans.txt", func(c *gin.Context) {
		content, err := ioutil.ReadFile("humans.txt")
		if err != nil {
			c.String(http.StatusNotFound, "humans.txt not found")
			return
		}
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusOK, string(content))
	})

	// Serve static assets (CSS, JS, images) with minification for CSS/JS
	p.router.GET("/assets/*filepath", func(c *gin.Context) {
		reqPath := c.Param("filepath")
		fullPath := filepath.Join("./assets", reqPath)
		
		// Check if file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		
		// Get file extension
		ext := filepath.Ext(fullPath)
		
		// For CSS and JS, serve minified version
		if ext == ".css" || ext == ".js" {
			minified, err := minifyAsset(fullPath)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			
			// Set appropriate content type
			if ext == ".css" {
				c.Header("Content-Type", "text/css; charset=utf-8")
			} else if ext == ".js" {
				c.Header("Content-Type", "application/javascript; charset=utf-8")
			}
			
			c.Data(http.StatusOK, c.GetHeader("Content-Type"), minified)
			return
		}
		
		// For images, check if optimization is requested
		isImage := ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
		if isImage {
			// Check for resize parameters
			widthStr := c.Query("w")
			heightStr := c.Query("h")
			optimize := c.Query("optimize") == "true"
			
			if widthStr != "" || heightStr != "" {
				// Resize requested
				width, _ := strconv.Atoi(widthStr)
				height, _ := strconv.Atoi(heightStr)
				
				resized, err := resizeImage(fullPath, width, height)
				if err != nil {
					log.Printf("Error resizing image: %v", err)
					c.File(fullPath)
					return
				}
				
				// Save to temp file and serve
				tmpFile, err := ioutil.TempFile("", "resized-*"+ext)
				if err != nil {
					c.File(fullPath)
					return
				}
				defer os.Remove(tmpFile.Name())
				defer tmpFile.Close()
				
				if err := imaging.Save(resized, tmpFile.Name()); err != nil {
					c.File(fullPath)
					return
				}
				
				c.File(tmpFile.Name())
				return
			} else if optimize && !isDevMode {
				// Optimize image
				optimized, err := convertToWebP(fullPath)
				if err != nil {
					log.Printf("Error optimizing image: %v", err)
					c.File(fullPath)
					return
				}
				
				// Determine content type
				contentType := "image/jpeg"
				switch ext {
				case ".png":
					contentType = "image/png"
				case ".gif":
					contentType = "image/gif"
				}
				
				c.Data(http.StatusOK, contentType, optimized)
				return
			}
		}
		
		// For other files, serve normally
		c.File(fullPath)
	})

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

// watchConfigFile watches config.yaml for changes in production mode
func (p *program) watchConfigFile() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Warning: Failed to create config file watcher: %v", err)
		return
	}
	defer watcher.Close()

	// Watch config.yaml
	configFile := "config.yaml"
	if err := watcher.Add(configFile); err != nil {
		log.Printf("Warning: Failed to watch config file: %v", err)
		return
	}
	
	log.Printf("Watching config file: %s (hot reload enabled)", configFile)

	// Debounce timer to avoid multiple rapid reloads
	debounceTimer := time.NewTimer(0)
	<-debounceTimer.C // Drain the timer

	for {
		select {
		case <-p.exit:
			// Stop watching when service is stopping
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Only react to write events
			if event.Op&fsnotify.Write == fsnotify.Write {
				// Debounce: wait 500ms before reloading
				debounceTimer.Reset(500 * time.Millisecond)
				go func() {
					<-debounceTimer.C
					log.Printf("Config file changed - reloading...")
					
					// Reload config
					config, err := loadConfig(configFile)
					if err != nil {
						log.Printf("Error: Failed to reload config: %v", err)
					} else {
						appConfig = config
						log.Println("✓ Config reloaded successfully")
					}
				}()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Config watcher error: %v", err)
		}
	}
}

func main() {
	var serviceAction string
	var devMode bool
	flag.StringVar(&serviceAction, "service", "", "Control the system service: install, uninstall, start, stop, restart")
	flag.BoolVar(&devMode, "dev", false, "Enable development mode with hot reload")
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

	// If dev mode is enabled, run with hot reload instead of as a service
	if devMode {
		log.Println("Starting Podium in development mode with hot reload...")
		runWithHotReload()
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

// runWithHotReload starts the server with file watching for auto-reload
func runWithHotReload() {
	// Set dev mode flag
	isDevMode = true
	
	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Failed to create file watcher:", err)
	}
	defer watcher.Close()

	// Watch templates, assets, posts, static, and config
	watchDirs := []string{"templates", "assets", "posts", "static"}
	watchFiles := []string{"config.yaml"}

	for _, dir := range watchDirs {
		if err := watcher.Add(dir); err != nil {
			log.Printf("Warning: Failed to watch directory %s: %v", dir, err)
		} else {
			log.Printf("Watching directory: %s", dir)
		}
	}

	for _, file := range watchFiles {
		if err := watcher.Add(file); err != nil {
			log.Printf("Warning: Failed to watch file %s: %v", file, err)
		} else {
			log.Printf("Watching file: %s", file)
		}
	}

	// Start the server in a goroutine
	go func() {
		prg := &program{exit: make(chan struct{})}
		prg.Start(nil)
	}()

	// Watch for file changes
	log.Println("Hot reload enabled - server will restart when files change")
	debounceTimer := time.NewTimer(0)
	<-debounceTimer.C // Drain the timer

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Only react to write and create events
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				// Debounce: wait 500ms before reloading
				debounceTimer.Reset(500 * time.Millisecond)
				go func() {
					<-debounceTimer.C
					log.Printf("File changed: %s - reloading templates and config...", event.Name)
					
					// Reload config
					config, err := loadConfig("config.yaml")
					if err != nil {
						log.Printf("Warning: Failed to reload config: %v", err)
					} else {
						appConfig = config
						log.Println("✓ Config reloaded")
					}
					
					// Templates are reloaded automatically by Gin on each request in dev mode
					log.Println("✓ Changes detected - templates will reload on next request")
				}()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
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
func loadMarkdownFile(folder, slug string) (string, string, []string, string, string, bool, bool, string, error) {
	filePath := filepath.Join(folder, slug+".md")
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", "", nil, "", "", false, false, "", err
	}

	// Parse front matter for metadata (tags, date, publishDate, featured, draft)
	lines := strings.Split(string(content), "\n")
	var tags []string
	title := slug
	var date string
	var publishDate string
	var isDraft bool
	var isFeatured bool
	var contentStartLine int
	var frontMatterLines []int
	
	// Check for front matter (tags, date, publishDate, featured, draft)
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
		if strings.HasPrefix(line, "PublishDate:") || strings.HasPrefix(line, "publishDate:") || strings.HasPrefix(line, "publish_date:") {
			publishDate = strings.TrimPrefix(line, "PublishDate:")
			publishDate = strings.TrimPrefix(publishDate, "publishDate:")
			publishDate = strings.TrimPrefix(publishDate, "publish_date:")
			publishDate = strings.TrimSpace(publishDate)
			frontMatterLines = append(frontMatterLines, i)
		}
		if strings.HasPrefix(line, "Featured:") || strings.HasPrefix(line, "featured:") {
			featuredStr := strings.TrimPrefix(line, "Featured:")
			featuredStr = strings.TrimPrefix(featuredStr, "featured:")
			featuredStr = strings.TrimSpace(featuredStr)
			isFeatured = strings.ToLower(featuredStr) == "true"
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
		if i > 20 {
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
	
	// Add lazy loading to images
	htmlWithLazyLoad := addLazyLoadingToImages(string(html))
	
	// Get plain text content for excerpts
	plainText := stripHTML(htmlWithLazyLoad)

	return htmlWithLazyLoad, title, tags, date, publishDate, isDraft, isFeatured, plainText, nil
}

// addLazyLoadingToImages adds loading="lazy" attribute to all img tags for better performance
func addLazyLoadingToImages(htmlContent string) string {
	// Replace <img with <img loading="lazy" if not already present
	// Also make images responsive by adding width and height auto-sizing
	htmlContent = strings.ReplaceAll(htmlContent, "<img ", `<img loading="lazy" `)
	
	// If loading="lazy" was already present (unlikely but possible), avoid duplicates
	htmlContent = strings.ReplaceAll(htmlContent, `loading="lazy" loading="lazy"`, `loading="lazy"`)
	
	return htmlContent
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
		content, _, _, _, _, _, _, _, err := loadMarkdownFile("posts", post.Slug)
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
	sitemap.WriteString(fmt.Sprintf("    <lastmod>%s</lastmod>\n", time.Now().Format("2006-01-02")))
	sitemap.WriteString("    <changefreq>daily</changefreq>\n")
	sitemap.WriteString("    <priority>1.0</priority>\n")
	sitemap.WriteString("  </url>\n")
	
	// Add posts list page
	sitemap.WriteString("  <url>\n")
	sitemap.WriteString(fmt.Sprintf("    <loc>%s/posts</loc>\n", appConfig.SiteURL))
	sitemap.WriteString(fmt.Sprintf("    <lastmod>%s</lastmod>\n", time.Now().Format("2006-01-02")))
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
			} else {
				sitemap.WriteString(fmt.Sprintf("    <lastmod>%s</lastmod>\n", time.Now().Format("2006-01-02")))
			}
		} else {
			sitemap.WriteString(fmt.Sprintf("    <lastmod>%s</lastmod>\n", time.Now().Format("2006-01-02")))
		}
		
		sitemap.WriteString("    <changefreq>monthly</changefreq>\n")
		sitemap.WriteString("    <priority>0.8</priority>\n")
		sitemap.WriteString("  </url>\n")
	}
	
	// Add static pages
	for _, page := range pages {
		sitemap.WriteString("  <url>\n")
		sitemap.WriteString(fmt.Sprintf("    <loc>%s/page/%s</loc>\n", appConfig.SiteURL, page.Slug))
		sitemap.WriteString(fmt.Sprintf("    <lastmod>%s</lastmod>\n", time.Now().Format("2006-01-02")))
		sitemap.WriteString("    <changefreq>monthly</changefreq>\n")
		sitemap.WriteString("    <priority>0.7</priority>\n")
		sitemap.WriteString("  </url>\n")
	}
	
	// Add RSS feed
	sitemap.WriteString("  <url>\n")
	sitemap.WriteString(fmt.Sprintf("    <loc>%s/feed.xml</loc>\n", appConfig.SiteURL))
	sitemap.WriteString(fmt.Sprintf("    <lastmod>%s</lastmod>\n", time.Now().Format("2006-01-02")))
	sitemap.WriteString("    <changefreq>daily</changefreq>\n")
	sitemap.WriteString("    <priority>0.5</priority>\n")
	sitemap.WriteString("  </url>\n")
	
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

// getCurrentYear returns the current year as a string
func getCurrentYear() string {
	return fmt.Sprintf("%d", time.Now().Year())
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
			_, title, _, _, _, isDraft, _, _, err := loadMarkdownFile("static", slug)
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
	var featuredPosts []PageLink

	files, err := ioutil.ReadDir("posts")
	if err != nil {
		log.Printf("Error reading posts folder: %v", err)
		return posts
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			slug := strings.TrimSuffix(file.Name(), ".md")
			
			// Read file to get title, tags, date, publishDate, draft status, featured, and content for excerpt
			_, title, tags, date, publishDate, isDraft, isFeatured, plainText, err := loadMarkdownFile("posts", slug)
			if err != nil {
				continue
			}

			// Skip draft posts
			if isDraft {
				continue
			}

			// Check if post is scheduled for future publication
			if publishDate != "" {
				pubTime, err := time.Parse("2006-01-02 15:04", publishDate)
				if err == nil && time.Now().Before(pubTime) {
					// Post is scheduled for the future, skip it
					continue
				}
			}
			
			// Generate excerpt and reading time
			excerpt := generateExcerpt(plainText, appConfig.ExcerptLength)
			readingTime := calculateReadingTime(plainText)

			postLink := PageLink{
				Title:       title,
				Slug:        slug,
				Tags:        tags,
				Date:        date,
				PublishDate: publishDate,
				Excerpt:     excerpt,
				ReadingTime: readingTime,
				Featured:    isFeatured,
			}

			// Separate featured and regular posts
			if isFeatured {
				featuredPosts = append(featuredPosts, postLink)
			} else {
				posts = append(posts, postLink)
			}
		}
	}

	// Sort featured posts by date (newest first)
	sort.Slice(featuredPosts, func(i, j int) bool {
		dateI, errI := time.Parse("2006-01-02", featuredPosts[i].Date)
		dateJ, errJ := time.Parse("2006-01-02", featuredPosts[j].Date)
		
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}
		
		return dateI.After(dateJ)
	})

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

	// Combine featured posts first, then regular posts
	allPosts := append(featuredPosts, posts...)
	return allPosts
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
