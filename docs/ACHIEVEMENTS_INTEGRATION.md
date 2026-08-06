# Achievements integration

Base revision: `feat/stats-tracking-system@c0d93bd55de690b156c65d1c581cf2cd4e1008cf`.

## Data model

The migration creates exactly two tables:

- `achievements`: numeric ID, bilingual copy, color, ordering and a JSON condition;
- `user_achievements`: the idempotent `(user_id, achievement_id)` relation and award time.

Image URLs are intentionally absent. The frontend builds
`/img/achievements/<achievement.slug>.svg` and falls back to `star.svg`.

The condition is a constrained JSON expression tree. A node contains exactly one
of `all`, `any`, or `{field, operator, value}`. Supported fields are `event`,
`test_slug`, `percentage` and `time_spent`; supported operators depend on the
field and are limited to `eq`, `in`, `gte` and `lte`. Depth and node count are
bounded. Conditions never execute code or become SQL fragments.

Apply migrations in this order:

1. `migrations/20260806_achievements.sql`
2. `migrations/20260806_achievements_seed.sql`

## Production wiring

Create one `achievements.Service` using the existing MySQL client and the site's
authoritative test catalog. Pass it to `api.NewManager` through
`api.WithAchievements(service, userIDProvider)`.

`UserIDProvider` is deliberately an interface: the public repository does not
contain the current secret authentication mechanism. Its implementation must
derive the numeric user ID from the existing authenticated request context. Do
not accept `user_id` from JSON, query parameters or a public header.

When configured, successful `POST /tests/submit/` responses include
`new_achievements`. Repeated submissions return an empty array because the
relation has a composite primary key and MySQL uses `ON CONFLICT DO NOTHING`
semantics through GORM. `GET /achievements/?language=ru|en` returns the complete
catalog with `earned` and optional `earned_at`.

Before production deployment, replace `StaticCatalog` with the site's current
test registry so the backend, not the browser, verifies that `test_name` exists.
Keep result validation before award evaluation. If statistics and achievements
must be committed atomically in the production codebase, wrap both writes in the
existing request transaction; that transaction API is not present in this public
snapshot.

The cross-origin test header and permissive CORS policy exist only in
`integration-kit/mock_backend`; they are not part of the Go production patch.
