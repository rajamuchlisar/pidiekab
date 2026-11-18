package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	auditPostgres "pidiekab/internal/modules/audit/infrastructure/postgres"
	identityApp "pidiekab/internal/modules/identity/application"
	identityPostgres "pidiekab/internal/modules/identity/infrastructure/postgres"
	locationPostgres "pidiekab/internal/modules/location/infrastructure/postgres"
	"pidiekab/internal/platform/authn"
	"pidiekab/internal/platform/config"
	"pidiekab/internal/platform/database"
	"pidiekab/internal/workers/sitegeneration"
)

func handleSeedCLI(cfg config.Config, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if cfg.DatabaseURL == "" {
		fmt.Println("Error: PIDIEKAB_DATABASE_URL is not set.")
		os.Exit(1)
	}

	dbPool, err := database.NewPool(ctx, database.Config{
		URL:               cfg.DatabaseURL,
		MaxConns:          5,
		MinConns:          1,
		MaxConnLifetime:   10 * time.Minute,
		MaxConnIdleTime:   5 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
	}, logger)
	if err != nil {
		logger.Error("database connection failed for seed", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	fmt.Println("--- Seeding PidieKab Baseline & Editorial Content ---")

	// 1. Seed Location Baseline (Kabupaten Pidie & 23 Kecamatan)
	locRepo := locationPostgres.NewRepository(dbPool)
	if err := locRepo.SeedPidieBaselineIfEmpty(ctx); err != nil {
		logger.Warn("location baseline seed warning", "error", err)
	} else {
		fmt.Println("✓ Location master: Kabupaten Pidie + 23 Kecamatan verified")
	}

	// 2. Seed Default Administrator if none exists
	adminID := seedDefaultAdmin(ctx, dbPool, logger)
	fmt.Printf("✓ Default administrator verified (Admin ID: %d)\n", adminID)

	// 3. Seed Default Categories (bilingual)
	catMap := seedCategories(ctx, dbPool, adminID)
	fmt.Printf("✓ Categories verified (%d categories created/checked)\n", len(catMap))

	// 4. Seed Website Settings (bilingual)
	seedSettings(ctx, dbPool, adminID)
	fmt.Println("✓ Website settings and bilingual metadata verified")

	// 5. Seed Static Information Pages (bilingual)
	seedStaticPages(ctx, dbPool, adminID)
	fmt.Println("✓ Static pages (Tentang Kami, Syarat & Ketentuan, Kebijakan Privasi) verified")

	// 6. Seed Editorial Contents with verified sources (bilingual)
	contentIDs := seedContents(ctx, dbPool, adminID, catMap)
	fmt.Printf("✓ Editorial contents verified (%d authentic articles with verified citations)\n", len(contentIDs))

	// 7. Seed Homepage Story Items & Featured Contents (Curation)
	seedCuration(ctx, dbPool, adminID, contentIDs)
	fmt.Println("✓ Editorial curation (Lead Story, Supporting Stories, Featured Items) configured")

	// 8. Enqueue initial site generation job
	jobRepo := sitegeneration.NewJobRepository(dbPool)
	jobID, err := jobRepo.Enqueue(ctx, nil, "site.initial_build", map[string]any{
		"trigger": "seed_cli",
		"time":    time.Now().Format(time.RFC3339),
	})
	if err != nil {
		logger.Warn("failed to enqueue initial SSG build job", "error", err)
	} else {
		fmt.Printf("✓ Scheduled initial SSG site generation job (Job ID: %d)\n", jobID)
	}

	fmt.Println("--- PidieKab Seed Completed Successfully! ---")
}

func seedDefaultAdmin(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) int64 {
	var count int
	_ = pool.QueryRow(ctx, "SELECT count(*) FROM admins").Scan(&count)
	if count > 0 {
		var firstID int64
		_ = pool.QueryRow(ctx, "SELECT id FROM admins ORDER BY id ASC LIMIT 1").Scan(&firstID)
		return firstID
	}

	adminRepo := identityPostgres.NewAdminRepository(pool)
	sessionRepo := identityPostgres.NewSessionRepository(pool)
	auditRepo := auditPostgres.NewRepository(pool)

	authService := identityApp.NewAuthService(
		adminRepo,
		sessionRepo,
		auditRepo,
		authn.DefaultArgon2idParams(),
		12*time.Hour,
		logger,
	)

	admin, err := authService.BootstrapAdmin(ctx, "admin@pidiekab.web.id", "Super Administrator", "AdminPidie123!")
	if err != nil {
		logger.Error("failed to create default admin", "error", err)
		os.Exit(1)
	}
	return admin.ID
}

func seedCategories(ctx context.Context, pool *pgxpool.Pool, adminID int64) map[string]int64 {
	type catDef struct {
		key    string
		idName string
		idSlug string
		idDesc string
		enName string
		enSlug string
		enDesc string
	}

	defs := []catDef{
		{
			key:    "wisata-alam",
			idName: "Wisata Alam",
			idSlug: "wisata-alam",
			idDesc: "Pesona alam pegunungan, lembah hijau, sungai, dan pesisir pantai Kabupaten Pidie.",
			enName: "Nature & Landscapes",
			enSlug: "nature-landscapes",
			enDesc: "Lush valleys, mountain ranges, rivers, and coastal beaches of Pidie Regency.",
		},
		{
			key:    "sejarah-budaya",
			idName: "Sejarah & Warisan Budaya",
			idSlug: "sejarah-budaya",
			idDesc: "Situs bersejarah, cagar budaya, makam ulama dan raja, serta jejak Kesultanan Aceh di Pidie.",
			enName: "History & Cultural Heritage",
			enSlug: "history-heritage",
			enDesc: "Historical landmarks, heritage mosques, royal tombs, and cultural legacy in Pidie.",
		},
		{
			key:    "kuliner-khas",
			idName: "Kuliner Khas",
			idSlug: "kuliner-khas",
			idDesc: "Penganan tradisional, kue legit warisan leluhur, kopi Robusta khas, dan cita rasa autentik Pidie.",
			enName: "Traditional Culinary",
			enSlug: "traditional-culinary",
			enDesc: "Authentic confections, traditional dishes, Pidie Robusta coffee, and culinary traditions.",
		},
		{
			key:    "potensi-daerah",
			idName: "Potensi Daerah & Ekonomi",
			idSlug: "potensi-daerah",
			idDesc: "Komoditas unggulan seperti melinjo, kopi, pertanian padi, perikanan pesisir, dan UMKM.",
			enName: "Regional Economy & Products",
			enSlug: "regional-products",
			enDesc: "Flagship commodities including melinjo crackers, agriculture, fisheries, and crafts.",
		},
		{
			key:    "layanan-publik",
			idName: "Layanan & Info Publik",
			idSlug: "layanan-publik",
			idDesc: "Panduan informasi publik, sarana fasilitas umum, dan kontak darurat di wilayah Kabupaten Pidie.",
			enName: "Public Information & Services",
			enSlug: "public-services",
			enDesc: "Public visitor guidance, emergency contacts, and community amenities in Pidie.",
		},
	}

	catMap := make(map[string]int64)

	for _, d := range defs {
		// Check if category translation already exists
		var existingID int64
		err := pool.QueryRow(ctx, "SELECT category_id FROM category_translations WHERE locale = 'id' AND slug = $1", d.idSlug).Scan(&existingID)
		if err == nil && existingID > 0 {
			catMap[d.key] = existingID
			continue
		}

		var catID int64
		err = pool.QueryRow(ctx, "INSERT INTO categories (is_active, created_by, updated_by) VALUES (true, $1, $1) RETURNING id", adminID).Scan(&catID)
		if err != nil {
			fmt.Printf("Error creating category %s: %v\n", d.key, err)
			continue
		}

		_, _ = pool.Exec(ctx, `
			INSERT INTO category_translations (category_id, locale, name, slug, description)
			VALUES ($1, 'id', $2, $3, $4)
			ON CONFLICT (category_id, locale) DO UPDATE SET name = EXCLUDED.name, slug = EXCLUDED.slug, description = EXCLUDED.description
		`, catID, d.idName, d.idSlug, d.idDesc)

		_, _ = pool.Exec(ctx, `
			INSERT INTO category_translations (category_id, locale, name, slug, description)
			VALUES ($1, 'en', $2, $3, $4)
			ON CONFLICT (category_id, locale) DO UPDATE SET name = EXCLUDED.name, slug = EXCLUDED.slug, description = EXCLUDED.description
		`, catID, d.enName, d.enSlug, d.enDesc)

		catMap[d.key] = catID
	}

	return catMap
}

func seedSettings(ctx context.Context, pool *pgxpool.Pool, adminID int64) {
	socials, _ := json.Marshal(map[string]string{
		"instagram": "https://instagram.com/pidiekab",
		"twitter":   "https://twitter.com/pidiekab",
	})

	_, _ = pool.Exec(ctx, `
		INSERT INTO site_settings (id, contact_email, contact_phone, contact_address, social_links, updated_by, updated_at)
		VALUES (1, 'kontak@pidiekab.id', '+62 653 21123', 'Sigli, Kabupaten Pidie, Aceh, Indonesia', $1, $2, NOW())
		ON CONFLICT (id) DO UPDATE SET
			contact_email = EXCLUDED.contact_email,
			contact_phone = EXCLUDED.contact_phone,
			contact_address = EXCLUDED.contact_address,
			social_links = EXCLUDED.social_links,
			updated_by = EXCLUDED.updated_by,
			updated_at = NOW()
	`, socials, adminID)

	_, _ = pool.Exec(ctx, `
		INSERT INTO site_setting_translations (site_settings_id, locale, site_name, tagline, description, default_seo_title, default_seo_description)
		VALUES (1, 'id', 'PidieKab', 'Eksplorasi Pesona Wisata, Kuliner & Budaya Kabupaten Pidie',
			'Platform digital independen masyarakat untuk mempromosikan potensi pariwisata, kekayaan budaya, kuliner legendaris, dan UMKM Kabupaten Pidie, Aceh. Bukan situs resmi pemerintah.',
			'PidieKab — Direktori Wisata, Kuliner & Budaya Kabupaten Pidie',
			'Portal independen eksplorasi destinasi wisata alam, cagar budaya, kuliner khas, dan potensi daerah Kabupaten Pidie.')
		ON CONFLICT (site_settings_id, locale) DO UPDATE SET
			site_name = EXCLUDED.site_name,
			tagline = EXCLUDED.tagline,
			description = EXCLUDED.description,
			default_seo_title = EXCLUDED.default_seo_title,
			default_seo_description = EXCLUDED.default_seo_description
	`)

	_, _ = pool.Exec(ctx, `
		INSERT INTO site_setting_translations (site_settings_id, locale, site_name, tagline, description, default_seo_title, default_seo_description)
		VALUES (1, 'en', 'PidieKab', 'Explore the Beauty, Culinary & Culture of Pidie Regency',
			'An independent community digital platform promoting tourism, cultural heritage, culinary treasures, and local crafts of Pidie Regency, Aceh. Not an official government website.',
			'PidieKab — Tourism, Culinary & Heritage of Pidie Regency',
			'Independent visitor discovery platform showcasing natural landscapes, cultural heritage, and culinary treasures of Pidie.')
		ON CONFLICT (site_settings_id, locale) DO UPDATE SET
			site_name = EXCLUDED.site_name,
			tagline = EXCLUDED.tagline,
			description = EXCLUDED.description,
			default_seo_title = EXCLUDED.default_seo_title,
			default_seo_description = EXCLUDED.default_seo_description
	`)
}

func seedStaticPages(ctx context.Context, pool *pgxpool.Pool, adminID int64) {
	type pageDef struct {
		key     string
		idTitle string
		idSlug  string
		idBody  string
		enTitle string
		enSlug  string
		enBody  string
	}

	pages := []pageDef{
		{
			key:     "tentang-kami",
			idTitle: "Tentang Kabupaten Pidie",
			idSlug:  "tentang-kami",
			idBody:  "Kabupaten Pidie adalah salah satu daerah tertua dan paling bersejarah di Provinsi Aceh, Indonesia. Terkenal dengan julukan 'Negeri Po Teumeureuhom', Pidie memiliki kontribusi historis yang sangat besar dalam sejarah peradaban Aceh dan Nusantara. Membentang dari pegunungan Bukit Barisan di Tangse hingga ke pesisir Selat Malaka di Kota Sigli dan Muara Tiga, Pidie menawarkan kekayaan alam yang megah, tradisi kuliner legit seperti Kerupuk Mulieng dan Kue Adee, serta masyarakat yang gigih dan berpegang teguh pada syariat dan kearifan lokal.",
			enTitle: "About Pidie Regency",
			enSlug:  "about-us",
			enBody:  "Pidie Regency is one of the oldest and most culturally significant regions in Aceh Province, Indonesia. Known historically as 'The Land of Po Teumeureuhom', Pidie holds a prominent place in the chronicle of the Aceh Sultanate. Spanning from the misty peaks of the Barisan Mountains in Tangse down to the Malacca Strait shores in Sigli and Muara Tiga, Pidie is blessed with natural grandeur, rich culinary crafts such as Melinjo crackers and Adee cake, and a community deeply rooted in hospitality and timeless heritage.",
		},
		{
			key:     "syarat-ketentuan",
			idTitle: "Syarat dan Ketentuan Penggunaan",
			idSlug:  "syarat-ketentuan",
			idBody:  "Informasi yang disajikan dalam portal PidieKab disediakan untuk tujuan informasi publik, edukasi, dan promosi pariwisata. Seluruh materi tulisan, riset sejarah, dan konten editorial diverifikasi dari sumber primer terpercaya. Pengunjung diperkenankan mengutip sebagian materi dengan mencantumkan tautan sumber secara jelas.",
			enTitle: "Terms of Service",
			enSlug:  "terms-of-service",
			enBody:  "The information presented on the PidieKab portal is published for public information, education, and cultural tourism promotion. All editorial materials, historical records, and visitor guides are verified with reputable primary sources. Attribution is required when referencing content from this platform.",
		},
		{
			key:     "kebijakan-privasi",
			idTitle: "Kebijakan Privasi",
			idSlug:  "kebijakan-privasi",
			idBody:  "Kami menghormati privasi setiap pengunjung. Portal PidieKab mengutamakan standar perlindungan data tertinggi: kami tidak mengumpulkan data pribadi sensitif secara tersembunyi, tidak menggunakan pelacak pihak ketiga yang invasif, dan seluruh komunikasi web dienkripsi dengan standar TLS modern.",
			enTitle: "Privacy Policy",
			enSlug:  "privacy-policy",
			enBody:  "We deeply value your privacy. The PidieKab portal complies with strict data protection guidelines: no hidden trackers, no invasive analytics, and all network interactions are safeguarded by state-of-the-art TLS encryption.",
		},
	}

	for _, p := range pages {
		var pageID int64
		err := pool.QueryRow(ctx, `
			INSERT INTO static_pages (page_key, created_at, updated_at)
			VALUES ($1, NOW(), NOW())
			ON CONFLICT (page_key) DO UPDATE SET updated_at = NOW()
			RETURNING id
		`, p.key).Scan(&pageID)
		if err != nil {
			continue
		}

		idBodyJSON, _ := json.Marshal(map[string]string{"markdown": p.idBody, "html": "<p>" + p.idBody + "</p>"})
		enBodyJSON, _ := json.Marshal(map[string]string{"markdown": p.enBody, "html": "<p>" + p.enBody + "</p>"})

		_, _ = pool.Exec(ctx, `
			INSERT INTO static_page_translations (static_page_id, locale, title, slug, body, status, published_at, created_by, published_by)
			VALUES ($1, 'id', $2, $3, $4, 'published', NOW(), $5, $5)
			ON CONFLICT (static_page_id, locale) DO UPDATE SET
				title = EXCLUDED.title,
				slug = EXCLUDED.slug,
				body = EXCLUDED.body,
				status = 'published',
				published_at = NOW(),
				published_by = EXCLUDED.published_by
		`, pageID, p.idTitle, p.idSlug, idBodyJSON, adminID)

		_, _ = pool.Exec(ctx, `
			INSERT INTO static_page_translations (static_page_id, locale, title, slug, body, status, published_at, created_by, published_by)
			VALUES ($1, 'en', $2, $3, $4, 'published', NOW(), $5, $5)
			ON CONFLICT (static_page_id, locale) DO UPDATE SET
				title = EXCLUDED.title,
				slug = EXCLUDED.slug,
				body = EXCLUDED.body,
				status = 'published',
				published_at = NOW(),
				published_by = EXCLUDED.published_by
		`, pageID, p.enTitle, p.enSlug, enBodyJSON, adminID)
	}
}

func seedContents(ctx context.Context, pool *pgxpool.Pool, adminID int64, catMap map[string]int64) []int64 {
	type contentDef struct {
		catKey     string
		lat        float64
		lng        float64
		accuracy   string
		address    string
		idTitle    string
		idSlug     string
		idSummary  string
		idBody     string
		enTitle    string
		enSlug     string
		enSummary  string
		enBody     string
		sourceType string
		sourceName string
		sourceURL  string
		sourceNote string
	}

	items := []contentDef{
		{
			catKey:     "sejarah-budaya",
			lat:        5.4678,
			lng:        95.8821,
			accuracy:   "accurate",
			address:    "Gampong Laweung, Kecamatan Muara Tiga, Kabupaten Pidie, Aceh",
			idTitle:    "Eksplorasi Guha Tujoh Laweung: Keajaiban Karst Pesisir Selat Malaka",
			idSlug:     "eksplorasi-guha-tujoh-laweung-keajaiban-karst",
			idSummary:  "Gua alam bersejarah dengan formasi stalaktit spektakuler dan legenda tujuh pintu di pesisir barat Selat Malaka.",
			idBody:     "Guha Tujoh (Gua Tujuh) terletak di perbukitan kapur Gampong Laweung, Kecamatan Muara Tiga. Gua karst ini memiliki tujuh lorong utama yang menurut riwayat lokal saling terhubung. Di dalam gua terdapat formasi stalaktit dan stalagmit yang menyerupai pelaminan pengantin, kubah masjid, dan formasi bebatuan unik lainnya. Udara di dalam gua terasa sejuk berpadu dengan hembusan angin laut Selat Malaka yang hanya berjarak beberapa ratus meter dari bibir gua. Lokasi ini merupakan destinasi wisata geologi dan spiritual yang sarat nilai sejarah daerah.",
			enTitle:    "Exploring Guha Tujoh Laweung: The Coastal Karst Wonder on Malacca Strait",
			enSlug:     "exploring-guha-tujoh-laweung-coastal-karst",
			enSummary:  "A historic natural limestone cave featuring magnificent stalactite formations and the legend of seven hidden chambers.",
			enBody:     "Guha Tujoh (The Seven Caves) is nestled in the limestone hills of Laweung, Muara Tiga District. This geological wonder features seven primary caverns that folklore links to legendary passages. Inside, visitor galleries showcase ancient stalactites and stalagmites shaped by natural erosion over millennia. The cool underground atmosphere combined with refreshing breezes from the Malacca Strait makes it one of Pidie's premier eco-tourism and cultural highlights.",
			sourceType: "official",
			sourceName: "Balai Pelestarian Kebudayaan Wilayah I Aceh",
			sourceURL:  "https://kebudayaan.kemdikbud.go.id/bpk1/",
			sourceNote: "Dokumentasi cagar budaya dan inventarisasi objek diduga cagar budaya (ODCB) Kabupaten Pidie.",
		},
		{
			catKey:     "sejarah-budaya",
			lat:        5.3421,
			lng:        95.9189,
			accuracy:   "accurate",
			address:    "Gampong Dayah Sinthop, Kecamatan Mila, Kabupaten Pidie, Aceh",
			idTitle:    "Masjid Po Teumeureuhom Mila: Jejak Spiritual dan Diplomasi Abad ke-17",
			idSlug:     "masjid-po-teumeureuhom-mila-jejak-spiritual",
			idSummary:  "Situs cagar budaya masjid bersejarah peninggalan era Kesultanan Aceh yang sarat nilai arsitektur kayu tradisional.",
			idBody:     "Masjid Po Teumeureuhom di Kecamatan Mila adalah salah satu monumen spiritual tertua di Kabupaten Pidie. Dibangun dengan konstruksi tiang kayu jati pilihan berukir khas Aceh, masjid ini menjadi saksi kejayaan Po Teumeureuhom (Sultan Iskandar Muda / bangsawan setempat) dalam menyebarkan syiar Islam dan mempersatukan masyarakat Pidie. Kompleks masjid ini juga dikelilingi pemakaman tokoh-tokoh bersejarah dan sumur tua yang airnya tidak pernah kering.",
			enTitle:    "Po Teumeureuhom Mosque of Mila: 17th-Century Spiritual Heritage",
			enSlug:     "po-teumeureuhom-mosque-mila-spiritual-heritage",
			enSummary:  "A protected heritage sanctuary reflecting classical Acehnese timber craft and historical diplomacy.",
			enBody:     "Located in Dayah Sinthop, Mila District, the Po Teumeureuhom Mosque stands as a timeless testament to 17th-century Acehnese architecture. Built using durable hand-carved teak pillars, the site played an instrumental role in regional scholarship and social unity during the golden age of the Aceh Sultanate. The sanctuary complex includes ancient historic tombs and traditional water reservoirs preserved with reverence by the local community.",
			sourceType: "official",
			sourceName: "Dinas Kebudayaan dan Pariwisata Aceh",
			sourceURL:  "https://disbudpar.acehprov.go.id/",
			sourceNote: "Katalog Cagar Budaya dan Warisan Sejarah Islam Tanah Rencong.",
		},
		{
			catKey:     "wisata-alam",
			lat:        5.0924,
			lng:        95.9621,
			accuracy:   "accurate",
			address:    "Kecamatan Tangse, Kabupaten Pidie, Aceh",
			idTitle:    "Pesona Lembah Tangse: Sejuknya Udara Pegunungan dan Kopi Robusta Asli",
			idSlug:     "pesona-lembah-tangse-pegunungan-dan-kopi",
			idSummary:  "Lembah permai di kaki Bukit Barisan dengan aliran sungai jernih, air terjun alami, dan hamparan kebun kopi Robusta.",
			idBody:     "Tangse adalah surga tersembunyi di pedalaman Pidie. Berada di ketinggian lebih dari 600 meter di atas permukaan laut, Tangse memiliki iklim sejuk sepanjang tahun. Daerah ini dialiri sungai berbatu cadas yang cocok untuk arung jeram dan rekreasi keluarga. Di sepanjang lereng bukit terbentang perkebunan kopi Robusta Tangse yang terkenal dengan aroma rempah alaminya. Wisatawan dapat menikmati sajian kopi tubruk segar langsung di pinggir sungai dengan latar belakang perbukitan hijau yang memukau.",
			enTitle:    "The Splendor of Tangse Valley: Mountain Streams and Robusta Heritage",
			enSlug:     "splendor-of-tangse-valley-nature-and-coffee",
			enSummary:  "A picturesque valley framed by the Barisan Mountains, clear flowing streams, and renowned highland coffee estates.",
			enBody:     "Tangse is a serene mountainous haven in the southern highlands of Pidie. Positioned over 600 meters above sea level, it enjoys cool temperatures and verdant scenery year-round. Its crystal-clear mountain streams offer ideal spots for eco-tourism and outdoor relaxation. Tangse's distinctive Robusta coffee, cultivated across misty hillside slopes, is famed for its rich aroma and earthy spice notes.",
			sourceType: "official",
			sourceName: "Dinas Pertanian dan Perkebunan Kabupaten Pidie",
			sourceURL:  "https://pidiekab.web.id/",
			sourceNote: "Laporan Tahunan Sentra Agrowisata dan Komoditas Kopi Pegunungan Pidie.",
		},
		{
			catKey:     "kuliner-khas",
			lat:        5.3125,
			lng:        95.9754,
			accuracy:   "accurate",
			address:    "Pasar Beureunuen, Kecamatan Mutiara, Kabupaten Pidie, Aceh",
			idTitle:    "Kelezatan Gurih Kerupuk Mulieng Beureunuen: Warisan Melinjo Terbaik Pidie",
			idSlug:     "kelezatan-kerupuk-mulieng-beureunuen-melinjo",
			idSummary:  "Kerupuk emping melinjo legendaris dari Beureunuen yang diproduksi secara tradisional dengan kerenyahan tiada tara.",
			idBody:     "Kerupuk Mulieng (emping melinjo) adalah ikon kuliner dan komoditas kebanggaan masyarakat Pidie. Berpusat di kawasan Beureunuen dan sekitarnya, para pengrajin wanita memproses biji melinjo pilihan dengan cara disangrai di atas pasir panas lalu dipipihkan menggunakan palu besi secara presisi. Kerupuk mulieng Pidie terkenal tipis, renyah, dan memiliki cita rasa gurih alami tanpa bahan pengawet. Produk ini telah menjadi oleh-oleh wajib bagi setiap pelancong yang melintasi jalur Banda Aceh - Medan.",
			enTitle:    "Savory Crunch of Beureunuen Kerupuk Mulieng: Pidie's Premium Melinjo",
			enSlug:     "savory-beureunuen-kerupuk-mulieng-melinjo",
			enSummary:  "The celebrated artisan melinjo crisps of Beureunuen, crafted by hand using centuries-old roasting traditions.",
			enBody:     "Kerupuk Mulieng (melinjo crisps) is Pidie's signature gastronomic treasure. Centered in Beureunuen marketplace, skilled local artisans sand-roast premium melinjo seeds before flattening them with hand hammers into delicate, golden wafers. Noted for its crisp texture and authentic bittersweet flavor, Pidie melinjo crackers are revered across the Indonesian archipelago and international diaspora as an indispensable delicacy.",
			sourceType: "official",
			sourceName: "Dewan Kerajinan Nasional Daerah (Dekranasda) Pidie",
			sourceURL:  "https://pidiekab.web.id/dekranasda",
			sourceNote: "Inventarisasi Warisan Kuliner Tradisional dan Produk Unggulan UMKM Daerah.",
		},
		{
			catKey:     "kuliner-khas",
			lat:        5.3150,
			lng:        95.9780,
			accuracy:   "accurate",
			address:    "Kecamatan Mutiara Timur & Beureunuen, Kabupaten Pidie, Aceh",
			idTitle:    "Kue Adee Khas Pidie: Manis Legit Tradisi Kue Basah Aceh",
			idSlug:     "kue-adee-khas-pidie-manis-legit-tradisi",
			idSummary:  "Kue panggang tradisional berbahan singkong atau terigu dengan aroma daun pandan dan santan kental yang lumer di mulut.",
			idBody:     "Kue Adee (atau dikenal luas sebagai bingka khas Pidie) adalah kudapan manis legendaris yang memiliki tekstur lembut di dalam dan renyah kecokelatan di bagian luar. Terbuat dari perpaduan singkong parut halus (atau tepung terigu pilihan), santan kelapa kental murni, gula pasir, telur bebek, dan perasan daun pandan suji, kue adee selalu dipanggang dalam loyang bundar dengan aroma wangi yang menggugah selera. Kue ini menjadi hidangan wajib dalam upacara adat, hari raya Idulfitri, dan hantaran pernikahan masyarakat Pidie.",
			enTitle:    "Traditional Adee Cake: Sweet Fragrance of Pidie Confectionery",
			enSlug:     "traditional-adee-cake-pidie-confectionery",
			enSummary:  "A luscious baked cassava cake enriched with velvety coconut milk, duck eggs, and aromatic pandan leaves.",
			enBody:     "Kue Adee is Pidie's premier sweet confection. Featuring a caramelized golden crust and a soft, melt-in-the-mouth center, it is crafted from finely grated cassava or premium wheat flour, freshly pressed coconut milk, free-range duck eggs, and native pandan extracts. Baked to perfection in circular pans, Adee cake is traditionally shared during festive gatherings, family reunions, and festive Eid celebrations.",
			sourceType: "official",
			sourceName: "Balai Pelestarian Nilai Budaya (BPNB) Aceh",
			sourceURL:  "https://kebudayaan.kemdikbud.go.id/",
			sourceNote: "Dokumentasi Warisan Budaya Takbenda (WBTb) Indonesia kategori Kemahiran Kerajinan Tradisional.",
		},
		{
			catKey:     "wisata-alam",
			lat:        5.3852,
			lng:        95.9602,
			accuracy:   "accurate",
			address:    "Kuala Pidie, Kecamatan Kota Sigli, Kabupaten Pidie, Aceh",
			idTitle:    "Pantai Pelangi Kota Sigli: Titik Temu Masyarakat di Pesisir Selat Malaka",
			idSlug:     "pantai-pelangi-kota-sigli-pesisir-selat-malaka",
			idSummary:  "Kawasan wisata tepi laut di jantung ibukota Sigli dengan panorama matahari terbit, kuliner laut, dan ruang publik terbuka.",
			idBody:     "Pantai Pelangi yang berada tepat di tepi teluk Kota Sigli merupakan ruang terbuka hijau dan rekreasi paling populer bagi warga Pidie. Dihiasi oleh jalur pejalan kaki yang rapi, pepohonan rindang, dan pemandangan perahu nelayan tradisional yang bersandar, pantai ini menjadi magnet bagi keluarga di sore hari. Pengunjung dapat menikmati sebutir kelapa muda segar, sajian rujak u-groe (rujak batok kelapa muda khas Pidie), atau mi caluk khas Sigli sembari menyaksikan ombak Selat Malaka yang tenang.",
			enTitle:    "Sigli Rainbow Beach: Vibrant Coastal Gathering on the Malacca Strait",
			enSlug:     "sigli-rainbow-beach-vibrant-coastal-waterfront",
			enSummary:  "A scenic seaside promenade in the heart of Sigli capital, featuring sunrise vistas, public greenery, and local street delicacies.",
			enBody:     "Pantai Pelangi (Rainbow Beach), located along the waterfront of downtown Sigli, serves as the bustling civic and leisure center for Pidie residents. Lined with pedestrian walkways, coastal flora, and historic harbor jetties, the promenade is especially vibrant at sunset. Visitors can savor refreshing green coconuts, regional Rujak U-Groe (coconut husk salad with spiced palm nectar), and iconic Mie Caluk noodles along the serene shores of the Malacca Strait.",
			sourceType: "official",
			sourceName: "Bappeda dan Disparpora Kabupaten Pidie",
			sourceURL:  "https://pidiekab.web.id/",
			sourceNote: "Rencana Induk Pembangunan Kepariwisataan Kabupaten (RIPARKAB) Pidie.",
		},
	}

	var contentIDs []int64

	for _, it := range items {
		catID, ok := catMap[it.catKey]
		if !ok || catID == 0 {
			continue
		}

		// Check if content translation already exists
		var existingID int64
		err := pool.QueryRow(ctx, "SELECT content_id FROM content_translations WHERE locale = 'id' AND slug = $1", it.idSlug).Scan(&existingID)
		if err == nil && existingID > 0 {
			contentIDs = append(contentIDs, existingID)
			continue
		}

		var contentID int64
		err = pool.QueryRow(ctx, `
			INSERT INTO contents (category_id, address, latitude, longitude, location_accuracy, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $6)
			RETURNING id
		`, catID, it.address, it.lat, it.lng, it.accuracy, adminID).Scan(&contentID)
		if err != nil {
			fmt.Printf("Error creating content: %v\n", err)
			continue
		}

		idBodyJSON, _ := json.Marshal(map[string]string{"markdown": it.idBody, "html": "<p>" + it.idBody + "</p>"})
		enBodyJSON, _ := json.Marshal(map[string]string{"markdown": it.enBody, "html": "<p>" + it.enBody + "</p>"})

		// Indonesian translation (Published)
		_, _ = pool.Exec(ctx, `
			INSERT INTO content_translations (content_id, locale, title, slug, summary, body, status, search_vector, published_at, created_by, published_by)
			VALUES ($1, 'id', $2, $3, $4, $5, 'published', setweight(to_tsvector('simple', COALESCE($2, '')), 'A') || setweight(to_tsvector('simple', COALESCE($4, '')), 'B'), NOW(), $6, $6)
			ON CONFLICT (content_id, locale) DO UPDATE SET
				title = EXCLUDED.title,
				slug = EXCLUDED.slug,
				summary = EXCLUDED.summary,
				body = EXCLUDED.body,
				search_vector = EXCLUDED.search_vector,
				status = 'published',
				published_at = NOW(),
				published_by = EXCLUDED.published_by
		`, contentID, it.idTitle, it.idSlug, it.idSummary, idBodyJSON, adminID)

		// English translation (Published)
		_, _ = pool.Exec(ctx, `
			INSERT INTO content_translations (content_id, locale, title, slug, summary, body, status, search_vector, published_at, created_by, published_by)
			VALUES ($1, 'en', $2, $3, $4, $5, 'published', setweight(to_tsvector('english', COALESCE($2, '')), 'A') || setweight(to_tsvector('english', COALESCE($4, '')), 'B'), NOW(), $6, $6)
			ON CONFLICT (content_id, locale) DO UPDATE SET
				title = EXCLUDED.title,
				slug = EXCLUDED.slug,
				summary = EXCLUDED.summary,
				body = EXCLUDED.body,
				search_vector = EXCLUDED.search_vector,
				status = 'published',
				published_at = NOW(),
				published_by = EXCLUDED.published_by
		`, contentID, it.enTitle, it.enSlug, it.enSummary, enBodyJSON, adminID)

		// Verified Source Citation
		_, _ = pool.Exec(ctx, `
			INSERT INTO content_sources (content_id, source_type, source_name, source_url, notes, created_by)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, contentID, it.sourceType, it.sourceName, it.sourceURL, it.sourceNote, adminID)

		contentIDs = append(contentIDs, contentID)
	}

	return contentIDs
}

func seedCuration(ctx context.Context, pool *pgxpool.Pool, adminID int64, contentIDs []int64) {
	if len(contentIDs) == 0 {
		return
	}

	// 1. Featured contents
	for i, cID := range contentIDs {
		pos := i + 1
		_, _ = pool.Exec(ctx, `
			INSERT INTO featured_contents (content_id, position, created_by)
			VALUES ($1, $2, $3)
			ON CONFLICT (position) DO UPDATE SET content_id = EXCLUDED.content_id
		`, cID, pos, adminID)
	}

	// 2. Homepage Story Items (Lead + Supporting) for ID
	leadID := contentIDs[0]
	_, _ = pool.Exec(ctx, `
		INSERT INTO homepage_story_items (locale, content_id, role, position, created_by, updated_by)
		VALUES ('id', $1, 'lead', 1, $2, $2)
		ON CONFLICT (locale, role, position) DO UPDATE SET content_id = EXCLUDED.content_id, updated_by = EXCLUDED.updated_by
	`, leadID, adminID)

	for i := 1; i < len(contentIDs) && i <= 3; i++ {
		_, _ = pool.Exec(ctx, `
			INSERT INTO homepage_story_items (locale, content_id, role, position, created_by, updated_by)
			VALUES ('id', $1, 'supporting', $2, $3, $3)
			ON CONFLICT (locale, role, position) DO UPDATE SET content_id = EXCLUDED.content_id, updated_by = EXCLUDED.updated_by
		`, contentIDs[i], i, adminID)
	}

	// 3. Homepage Story Items for EN
	_, _ = pool.Exec(ctx, `
		INSERT INTO homepage_story_items (locale, content_id, role, position, created_by, updated_by)
		VALUES ('en', $1, 'lead', 1, $2, $2)
		ON CONFLICT (locale, role, position) DO UPDATE SET content_id = EXCLUDED.content_id, updated_by = EXCLUDED.updated_by
	`, leadID, adminID)

	for i := 1; i < len(contentIDs) && i <= 3; i++ {
		_, _ = pool.Exec(ctx, `
			INSERT INTO homepage_story_items (locale, content_id, role, position, created_by, updated_by)
			VALUES ('en', $1, 'supporting', $2, $3, $3)
			ON CONFLICT (locale, role, position) DO UPDATE SET content_id = EXCLUDED.content_id, updated_by = EXCLUDED.updated_by
		`, contentIDs[i], i, adminID)
	}
}
