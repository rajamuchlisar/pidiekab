-- +goose Up
-- ---------------------------------------------------------------------------
-- 1. Identity & access
-- ---------------------------------------------------------------------------

CREATE TABLE admins (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email           TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    display_name    TEXT NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT admins_email_normalized CHECK (email = LOWER(BTRIM(email))),
    CONSTRAINT admins_email_nonempty CHECK (email <> ''),
    CONSTRAINT admins_display_name_nonempty CHECK (BTRIM(display_name) <> '')
);

CREATE UNIQUE INDEX admins_email_uq ON admins (email);

CREATE TABLE admin_sessions (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    admin_id        BIGINT NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    token_hash      TEXT NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT admin_sessions_token_hash_nonempty CHECK (token_hash <> '')
);

CREATE UNIQUE INDEX admin_sessions_token_hash_uq ON admin_sessions (token_hash);
CREATE INDEX admin_sessions_admin_expiry_idx ON admin_sessions (admin_id, expires_at);

-- ---------------------------------------------------------------------------
-- 2. Master administrative locations
-- ---------------------------------------------------------------------------

CREATE TABLE regencies (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name            TEXT NOT NULL,
    code            TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT regencies_name_nonempty CHECK (BTRIM(name) <> '')
);

CREATE UNIQUE INDEX regencies_code_uq ON regencies (code) WHERE code IS NOT NULL;
CREATE UNIQUE INDEX regencies_name_uq ON regencies (LOWER(name));

CREATE TABLE districts (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    regency_id      BIGINT NOT NULL REFERENCES regencies(id) ON DELETE RESTRICT,
    name            TEXT NOT NULL,
    code            TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT districts_name_nonempty CHECK (BTRIM(name) <> '')
);

CREATE UNIQUE INDEX districts_regency_name_uq
    ON districts (regency_id, LOWER(name));
CREATE UNIQUE INDEX districts_code_uq
    ON districts (code) WHERE code IS NOT NULL;
CREATE INDEX districts_regency_idx ON districts (regency_id);

CREATE TABLE villages (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    district_id     BIGINT NOT NULL REFERENCES districts(id) ON DELETE RESTRICT,
    name            TEXT NOT NULL,
    code            TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT villages_name_nonempty CHECK (BTRIM(name) <> '')
);

CREATE UNIQUE INDEX villages_district_name_uq
    ON villages (district_id, LOWER(name));
CREATE UNIQUE INDEX villages_code_uq
    ON villages (code) WHERE code IS NOT NULL;
CREATE INDEX villages_district_idx ON villages (district_id);

-- ---------------------------------------------------------------------------
-- 3. Media library
-- ---------------------------------------------------------------------------

CREATE TABLE media_assets (
    id                  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    storage_key         TEXT NOT NULL,
    original_filename   TEXT,
    mime_type           TEXT NOT NULL,
    file_size           BIGINT NOT NULL,
    width               INTEGER,
    height              INTEGER,
    alt_text_id         TEXT,
    alt_text_en         TEXT,
    source_url          TEXT,
    rights_status       TEXT,
    attribution         TEXT,
    uploaded_by         BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT media_storage_key_nonempty CHECK (BTRIM(storage_key) <> ''),
    CONSTRAINT media_mime_type_nonempty CHECK (BTRIM(mime_type) <> ''),
    CONSTRAINT media_file_size_positive CHECK (file_size > 0),
    CONSTRAINT media_width_positive CHECK (width IS NULL OR width > 0),
    CONSTRAINT media_height_positive CHECK (height IS NULL OR height > 0)
);

CREATE UNIQUE INDEX media_assets_storage_key_uq ON media_assets (storage_key);
CREATE INDEX media_assets_uploaded_by_idx ON media_assets (uploaded_by);

-- ---------------------------------------------------------------------------
-- 4. Categories
-- ---------------------------------------------------------------------------

CREATE TABLE categories (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_by      BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    updated_by      BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE category_translations (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category_id     BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    locale          VARCHAR(2) NOT NULL,
    name            TEXT NOT NULL,
    description     TEXT,
    slug            TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT category_translations_locale_chk CHECK (locale IN ('id', 'en')),
    CONSTRAINT category_translations_name_nonempty CHECK (BTRIM(name) <> ''),
    CONSTRAINT category_translations_slug_nonempty CHECK (BTRIM(slug) <> ''),
    CONSTRAINT category_translations_category_locale_uq UNIQUE (category_id, locale),
    CONSTRAINT category_translations_locale_slug_uq UNIQUE (locale, slug)
);

CREATE INDEX category_translations_category_idx
    ON category_translations (category_id);

-- ---------------------------------------------------------------------------
-- 5. Content identities and localized versions
-- ---------------------------------------------------------------------------

CREATE TABLE contents (
    id                  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category_id         BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    village_id          BIGINT REFERENCES villages(id) ON DELETE RESTRICT,
    address             TEXT,
    latitude            NUMERIC(9,6),
    longitude           NUMERIC(9,6),
    location_accuracy   TEXT,
    main_media_id       BIGINT REFERENCES media_assets(id) ON DELETE RESTRICT,
    created_by          BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    updated_by          BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT contents_location_accuracy_chk
        CHECK (location_accuracy IS NULL OR location_accuracy IN ('accurate', 'approximate')),
    CONSTRAINT contents_latitude_range_chk
        CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90),
    CONSTRAINT contents_longitude_range_chk
        CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180),
    CONSTRAINT contents_coordinates_pair_chk
        CHECK ((latitude IS NULL) = (longitude IS NULL)),
    CONSTRAINT contents_location_accuracy_requires_coordinates_chk
        CHECK (location_accuracy IS NULL OR latitude IS NOT NULL)
);

CREATE INDEX contents_category_idx ON contents (category_id);
CREATE INDEX contents_village_idx ON contents (village_id);
CREATE INDEX contents_main_media_idx ON contents (main_media_id);

CREATE TABLE content_translations (
    id                  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    content_id          BIGINT NOT NULL REFERENCES contents(id) ON DELETE CASCADE,
    locale              VARCHAR(2) NOT NULL,
    title               TEXT NOT NULL,
    slug                TEXT NOT NULL,
    summary             TEXT,
    body                JSONB NOT NULL,
    extra_data          JSONB,
    seo_title           TEXT,
    seo_description     TEXT,
    status              TEXT NOT NULL DEFAULT 'draft',
    published_at        TIMESTAMPTZ,
    search_vector       TSVECTOR,
    created_by          BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    updated_by          BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    published_by        BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT content_translations_locale_chk CHECK (locale IN ('id', 'en')),
    CONSTRAINT content_translations_status_chk CHECK (status IN ('draft', 'published', 'archived')),
    CONSTRAINT content_translations_title_nonempty CHECK (BTRIM(title) <> ''),
    CONSTRAINT content_translations_slug_nonempty CHECK (BTRIM(slug) <> ''),
    CONSTRAINT content_translations_body_object_chk CHECK (JSONB_TYPEOF(body) = 'object'),
    CONSTRAINT content_translations_content_locale_uq UNIQUE (content_id, locale),
    CONSTRAINT content_translations_locale_slug_uq UNIQUE (locale, slug),
    CONSTRAINT content_translations_published_fields_chk
        CHECK (status <> 'published' OR (published_at IS NOT NULL AND published_by IS NOT NULL))
);

CREATE INDEX content_translations_public_listing_idx
    ON content_translations (locale, published_at DESC, id DESC)
    WHERE status = 'published';
CREATE INDEX content_translations_content_idx ON content_translations (content_id);
CREATE INDEX content_translations_search_gin_idx
    ON content_translations USING GIN (search_vector);

CREATE TABLE content_sources (
    id                  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    content_id          BIGINT NOT NULL REFERENCES contents(id) ON DELETE CASCADE,
    source_type         TEXT NOT NULL,
    source_name         TEXT NOT NULL,
    source_url          TEXT,
    source_published_at TIMESTAMPTZ,
    accessed_at         TIMESTAMPTZ,
    notes               TEXT,
    created_by          BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT content_sources_type_chk
        CHECK (source_type IN ('research', 'visit', 'official', 'social')),
    CONSTRAINT content_sources_name_nonempty CHECK (BTRIM(source_name) <> '')
);

CREATE INDEX content_sources_content_idx ON content_sources (content_id);

-- ---------------------------------------------------------------------------
-- 6. Editorial curation
-- ---------------------------------------------------------------------------

CREATE TABLE featured_contents (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    content_id      BIGINT NOT NULL REFERENCES contents(id) ON DELETE CASCADE,
    position        INTEGER NOT NULL,
    created_by      BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT featured_contents_position_positive CHECK (position > 0),
    CONSTRAINT featured_contents_content_uq UNIQUE (content_id),
    CONSTRAINT featured_contents_position_uq UNIQUE (position)
);

CREATE TABLE homepage_story_items (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    locale          VARCHAR(2) NOT NULL,
    content_id      BIGINT NOT NULL REFERENCES contents(id) ON DELETE CASCADE,
    role            TEXT NOT NULL,
    position        INTEGER NOT NULL,
    created_by      BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    updated_by      BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT homepage_story_locale_chk CHECK (locale IN ('id', 'en')),
    CONSTRAINT homepage_story_role_chk CHECK (role IN ('lead', 'supporting')),
    CONSTRAINT homepage_story_position_positive CHECK (position > 0),
    CONSTRAINT homepage_story_locale_content_uq UNIQUE (locale, content_id),
    CONSTRAINT homepage_story_locale_role_position_uq UNIQUE (locale, role, position)
);

CREATE UNIQUE INDEX homepage_story_one_lead_per_locale_uq
    ON homepage_story_items (locale) WHERE role = 'lead';
CREATE INDEX homepage_story_display_idx
    ON homepage_story_items (locale, role, position);

-- ---------------------------------------------------------------------------
-- 7. Static information pages
-- ---------------------------------------------------------------------------

CREATE TABLE static_pages (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    page_key        TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT static_pages_page_key_nonempty CHECK (BTRIM(page_key) <> '')
);

CREATE UNIQUE INDEX static_pages_page_key_uq ON static_pages (page_key);

CREATE TABLE static_page_translations (
    id                  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    static_page_id      BIGINT NOT NULL REFERENCES static_pages(id) ON DELETE CASCADE,
    locale              VARCHAR(2) NOT NULL,
    title               TEXT NOT NULL,
    slug                TEXT NOT NULL,
    body                JSONB NOT NULL,
    seo_title           TEXT,
    seo_description     TEXT,
    status              TEXT NOT NULL DEFAULT 'draft',
    published_at        TIMESTAMPTZ,
    created_by          BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    updated_by          BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    published_by        BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT static_page_translations_locale_chk CHECK (locale IN ('id', 'en')),
    CONSTRAINT static_page_translations_status_chk CHECK (status IN ('draft', 'published', 'archived')),
    CONSTRAINT static_page_translations_title_nonempty CHECK (BTRIM(title) <> ''),
    CONSTRAINT static_page_translations_slug_nonempty CHECK (BTRIM(slug) <> ''),
    CONSTRAINT static_page_translations_body_object_chk CHECK (JSONB_TYPEOF(body) = 'object'),
    CONSTRAINT static_page_translations_page_locale_uq UNIQUE (static_page_id, locale),
    CONSTRAINT static_page_translations_locale_slug_uq UNIQUE (locale, slug),
    CONSTRAINT static_page_translations_published_fields_chk
        CHECK (status <> 'published' OR (published_at IS NOT NULL AND published_by IS NOT NULL))
);

CREATE INDEX static_page_translations_public_idx
    ON static_page_translations (locale, status);

-- ---------------------------------------------------------------------------
-- 8. Website settings
-- ---------------------------------------------------------------------------

CREATE TABLE site_settings (
    id                  SMALLINT PRIMARY KEY DEFAULT 1,
    logo_media_id       BIGINT REFERENCES media_assets(id) ON DELETE RESTRICT,
    favicon_media_id    BIGINT REFERENCES media_assets(id) ON DELETE RESTRICT,
    contact_email       TEXT,
    contact_phone       TEXT,
    contact_address     TEXT,
    social_links        JSONB,
    updated_by          BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT site_settings_singleton_chk CHECK (id = 1)
);

CREATE TABLE site_setting_translations (
    id                      BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    site_settings_id        SMALLINT NOT NULL REFERENCES site_settings(id) ON DELETE CASCADE,
    locale                  VARCHAR(2) NOT NULL,
    site_name               TEXT NOT NULL,
    tagline                 TEXT,
    description             TEXT,
    default_seo_title       TEXT,
    default_seo_description TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT site_setting_translations_locale_chk CHECK (locale IN ('id', 'en')),
    CONSTRAINT site_setting_translations_name_nonempty CHECK (BTRIM(site_name) <> ''),
    CONSTRAINT site_setting_translations_locale_uq UNIQUE (site_settings_id, locale)
);

-- ---------------------------------------------------------------------------
-- 9. Publication jobs and URL redirects
-- ---------------------------------------------------------------------------

CREATE TABLE site_generation_jobs (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    job_type        TEXT NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}'::JSONB,
    status          TEXT NOT NULL DEFAULT 'pending',
    attempts        INTEGER NOT NULL DEFAULT 0,
    available_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at       TIMESTAMPTZ,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT site_generation_jobs_type_nonempty CHECK (BTRIM(job_type) <> ''),
    CONSTRAINT site_generation_jobs_status_chk
        CHECK (status IN ('pending', 'processing', 'succeeded', 'failed')),
    CONSTRAINT site_generation_jobs_attempts_chk CHECK (attempts >= 0)
);

CREATE INDEX site_generation_jobs_claim_idx
    ON site_generation_jobs (available_at, created_at)
    WHERE status = 'pending';
CREATE INDEX site_generation_jobs_status_idx
    ON site_generation_jobs (status, updated_at);

CREATE TABLE url_redirects (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    old_path        TEXT NOT NULL,
    new_path        TEXT NOT NULL,
    status_code     SMALLINT NOT NULL DEFAULT 301,
    created_by      BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT url_redirects_old_path_nonempty CHECK (BTRIM(old_path) <> ''),
    CONSTRAINT url_redirects_new_path_nonempty CHECK (BTRIM(new_path) <> ''),
    CONSTRAINT url_redirects_not_self_chk CHECK (old_path <> new_path),
    CONSTRAINT url_redirects_status_code_chk CHECK (status_code = 301)
);

CREATE UNIQUE INDEX url_redirects_old_path_uq ON url_redirects (old_path);

-- ---------------------------------------------------------------------------
-- 10. Audit log
-- ---------------------------------------------------------------------------

CREATE TABLE audit_logs (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    admin_id        BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    action          TEXT NOT NULL,
    entity_type     TEXT NOT NULL,
    entity_id       BIGINT,
    metadata        JSONB,
    ip_address      INET,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT audit_logs_action_nonempty CHECK (BTRIM(action) <> ''),
    CONSTRAINT audit_logs_entity_type_nonempty CHECK (BTRIM(entity_type) <> '')
);

CREATE INDEX audit_logs_created_idx ON audit_logs (created_at DESC);
CREATE INDEX audit_logs_entity_idx ON audit_logs (entity_type, entity_id);
CREATE INDEX audit_logs_admin_created_idx ON audit_logs (admin_id, created_at DESC);


-- +goose Down
-- ---------------------------------------------------------------------------
-- Drop tables in reverse dependency order
-- ---------------------------------------------------------------------------

DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS url_redirects;
DROP TABLE IF EXISTS site_generation_jobs;
DROP TABLE IF EXISTS site_setting_translations;
DROP TABLE IF EXISTS site_settings;
DROP TABLE IF EXISTS static_page_translations;
DROP TABLE IF EXISTS static_pages;
DROP TABLE IF EXISTS homepage_story_items;
DROP TABLE IF EXISTS featured_contents;
DROP TABLE IF EXISTS content_sources;
DROP TABLE IF EXISTS content_translations;
DROP TABLE IF EXISTS contents;
DROP TABLE IF EXISTS category_translations;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS media_assets;
DROP TABLE IF EXISTS villages;
DROP TABLE IF EXISTS districts;
DROP TABLE IF EXISTS regencies;
DROP TABLE IF EXISTS admin_sessions;
DROP TABLE IF EXISTS admins;
