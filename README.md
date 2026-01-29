# shawnbecker.de

My personal portfolio site. Built with Go because I wanted to learn it properly, and because I was tired of JavaScript build tooling.

**Live:** [shawnbecker.de](https://shawnbecker.de)

## Stack

- **Go** + [Templ](https://templ.guide/) for server-side rendering
- **HTMX** for the contact form (no client-side JS framework)
- **i18n** with English/German support
- Deployed on Hetzner

## Why Go?

After years of React/Next.js, I wanted something simpler. Go's standard library handles routing, the binary just runs, and Templ gives me type-safe templates without the JSX mental model. It's refreshing.

## Local Dev

```bash
# needs Go 1.21+, templ cli, and air (for hot reload)
cp .env.example .env  # add your keys
air
```

## Structure

```
handlers/     # route handlers
views/        # templ components
i18n/         # translations
static/       # css, images
```

The CSS follows a semantic-first approach - most styling comes from HTML elements, not utility classes.

## Contact

The form uses hCaptcha + Resend for email. Nothing fancy, it just works.

---

Built in Berlin, 2025.
