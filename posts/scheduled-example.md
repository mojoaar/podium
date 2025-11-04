Tags: scheduling, example, future
Date: 2025-11-05
PublishDate: 2025-12-01 09:00

# Scheduled Post Example

This post is scheduled to publish on December 1st, 2025 at 9:00 AM!

## How Post Scheduling Works

Add a `PublishDate` field to your post's front matter with a date and time in the future:

```markdown
Tags: announcement
Date: 2025-11-05
PublishDate: 2025-12-01 09:00

# Your Post Title

Your content here...
```

## Date Format

The `PublishDate` field uses the format: `YYYY-MM-DD HH:MM`

- **YYYY-MM-DD** - The date (e.g., 2025-12-01)
- **HH:MM** - 24-hour time (e.g., 09:00 for 9 AM, 14:30 for 2:30 PM)

## Behavior

- Posts with future `PublishDate` are hidden from:
  - Blog posts list
  - Individual post pages
  - Tag pages
  - RSS feed
  - Sitemap
- Once the publish date/time arrives, the post automatically becomes visible
- No server restart needed!

## Use Cases

- **Scheduled announcements** - Prepare content in advance
- **Timed releases** - Coordinate with events or launches
- **Content planning** - Write posts when you have time, publish when you want
- **Time-zone coordination** - Publish at optimal times for your audience

This post will become visible on December 1st, 2025 at 9:00 AM. Until then, it's hidden from public view!
