# 06 Search And Discovery

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `public-global-search.md`
- `search-category-filters.md`
- `search-result-ordering.md`
- `search-results-pagination.md`
- `search-result-card-content.md`

---

## Materi sumber: `public-global-search.md`

# Public Search

## Decision SRCH-001 --- Global Search

**Status: Agreed**

PidieKab MVP will provide one global search experience across all
published MVP content categories.

## Searchable Content

The global search covers: - Wisata & Destinasi. - Kuliner & UMKM. -
Budaya & Sejarah.

Search results must respect the publication status and language rules
defined in LNG-001 and LNG-002. Content that is not published in the
requested language must not be presented as though it were available in
that language.

## MVP Search Behaviour

-   One search input is available from the public website navigation or
    another consistently accessible location.
-   Visitors can search without signing in.
-   Results link to the corresponding public detail pages.
-   Results are limited to published content.
-   The interface should identify each result's category.
-   Empty queries should not trigger a broad, unbounded result list.
-   A clear empty state should be shown when no matching content is
    found.
-   Search should work for Indonesian and English content according to
    the active locale.

## Search Scope and Ranking

The initial search implementation should prioritize useful matching
against: - Content title. - Summary/short description. - Main body text
where appropriate. - Relevant category or taxonomy labels.

Ranking, typo tolerance, stemming, autocomplete, and advanced filters
are not yet committed. Keep the MVP implementation simple and evaluate
PostgreSQL full-text search or a straightforward indexed search approach
during technical design.

## Out of Scope for MVP

-   Dedicated search engine infrastructure.
-   AI-powered semantic search.
-   Personalized search ranking.
-   Search analytics dashboard.
-   Complex faceted filtering unless later justified by usability
    testing.

## Open Decisions

-   Whether search results should have category filters.
-   Whether autocomplete/suggestions are needed at launch.
-   Minimum query length and pagination behaviour.
-   Search indexing strategy and language-specific text configuration.

---

## Materi sumber: `search-category-filters.md`

# Search Result Category Filters

## Decision SRCH-002 --- Category Filters in Global Search

**Status: Agreed**

The PidieKab MVP global search results will include category filters.

## Filter Options

The search results page will provide: - Semua. - Wisata & Destinasi. -
Kuliner & UMKM. - Budaya & Sejarah.

The default selection is **Semua**, so visitors initially see matching
results from all published MVP categories.

## Behaviour

1.  Selecting a category narrows the current search results to that
    category.
2.  Selecting **Semua** restores results across all three MVP
    categories.
3.  The active search query is retained when the visitor changes
    category.
4.  Counts per category may be displayed if they can be calculated
    reliably, but are not mandatory for MVP.
5.  Filters only include categories with eligible published content in
    the selected language.
6.  Empty states should distinguish between no results for the query
    overall and no results within the selected category.

## Scope Boundaries

-   These are category filters, not a full faceted-search system.
-   No advanced filters by location, rating, price, opening hours, or
    popularity in MVP.
-   No AI-powered semantic search.
-   Category filtering must respect locale-specific publication rules.

## Open Decisions

-   Should the category filters be displayed as tabs/chips or a dropdown
    on mobile?
-   Should category filters also be available on the Jelajahi Pidie
    landing page, or only on search results?
-   Should visitors be able to sort search results, or should the system
    use a fixed relevance order?

---

## Materi sumber: `search-result-ordering.md`

# Search Result Ordering

## Decision SRCH-003 --- Relevance-Based Search Ordering

**Status: Agreed**

PidieKab MVP global search results will be ordered by search relevance.
Results that best match the visitor's query should appear first.

## MVP Behaviour

-   Relevance is the default and primary ordering for search results.
-   Relevance is calculated from matching content fields, with stronger
    weight for title matches than summary or body matches.
-   Category filtering (SRCH-002) preserves relevance ordering within
    the selected category.
-   The same query and locale should produce a stable, predictable
    order.
-   A deterministic secondary ordering (such as publication date and
    content ID) should be used when relevance scores are equal.
-   Only content published in the active locale is eligible to appear.

## Not Included in This Decision

-   Admin-curated promotion overriding relevance.
-   Newest-first or popularity-based sorting controls for visitors.
-   Personalised ranking.
-   AI/semantic ranking.

## Technical Design Guidance

During technical specification, define a simple, explainable relevance
score. PostgreSQL full-text search and/or indexed text matching may be
evaluated. Title matches should carry more weight than summary and body
matches. Exact scoring weights, text-search configuration per locale,
typo tolerance, and indexing strategy remain technical decisions.

## Open Decisions

-   Should search results support pagination or a "Load more"
    interaction?
-   Should a visitor be able to change sorting, or should relevance
    remain fixed?
-   Should exact title matches receive a distinct visual treatment?

---

## Materi sumber: `search-results-pagination.md`

# Search Results Pagination

## Decision SRCH-004 --- Pagination

**Status: Agreed**

PidieKab MVP search results will use page-based pagination.

## MVP Behaviour

-   Search results are divided into numbered pages.
-   The initial page size is **12 results per page** as the proposed
    default for implementation.
-   Visitors can navigate to previous/next pages and directly to
    available page numbers.
-   The current search query, selected category filter, and active
    locale are preserved when changing pages.
-   Page navigation should use crawlable, shareable URLs where
    practical.
-   Changing the search query or category filter resets the result view
    to page 1.
-   Pagination only covers results eligible for publication in the
    active locale.
-   Empty and out-of-range page states must be handled clearly.

## Why Pagination

Pagination provides predictable navigation, supports direct linking and
search-engine crawlability, and avoids loading an unbounded number of
results at once.

## Technical Considerations

-   Use deterministic ordering so records do not unpredictably move
    between pages when relevance scores are equal.
-   Consider database pagination performance and suitable indexes.
-   Keep query, category, locale, and page state synchronized with the
    URL.
-   Exact URL format and whether offset-based or cursor-based pagination
    is used remain technical design decisions.

## Out of Scope for MVP

-   Infinite scroll.
-   Automatic loading of subsequent pages.
-   User-configurable page size.

## Open Decisions

-   Confirm final page size through responsive UI design and usability
    testing.
-   Define the exact URL/query parameter convention.

---

## Materi sumber: `search-result-card-content.md`

# Search Result Card Content

## Decision SRCH-005 --- Minimal Search Result Information

**Status: Agreed**

Each public search result will display only these information fields:

1.  **Title** --- the content name or title.
2.  **Category** --- Wisata & Destinasi, Kuliner & UMKM, or Budaya &
    Sejarah.
3.  **Location** --- the relevant location information.

Search result cards will not display a thumbnail/image or
description/summary.

## Behaviour and Data Rules

-   The title links to the corresponding published detail page.
-   Category labels must reflect the content's actual MVP category.
-   Location is displayed when relevant and available in the content
    data.
-   The system must not invent or infer a location when it is unknown.
-   For content where a geographic location is not applicable or has not
    been established, the UI should handle the missing location
    consistently without displaying misleading information.
-   Results continue to follow the active locale, publication status,
    relevance ordering, category filter, and pagination decisions.

## Design Intent

Keep search results compact, scannable, and focused on helping visitors
identify the right content quickly. Visual media and longer editorial
descriptions belong on the detail page, not in the search result card.

## Out of Scope

-   Search-result thumbnails.
-   Search-result description excerpts.
-   Ratings, popularity indicators, and promotional badges unless
    separately approved.

---
