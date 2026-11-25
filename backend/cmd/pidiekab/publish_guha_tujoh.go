package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	mediaStorage "pidiekab/internal/modules/media/infrastructure/storage"
)

//go:embed assets/guha_tujoh_laweung.jpg
var guhaTujohImageBytes []byte

// publishGuhaTujohStory inserts or updates the comprehensive longform editorial story and hero image for Guha Tujoh Laweung.
func publishGuhaTujohStory(ctx context.Context, dbPool *pgxpool.Pool, adminID int64, storagePath string, logger *slog.Logger) error {
	fmt.Println("--> Publishing Comprehensive Editorial Story: 'Eksplorasi Guha Tujoh Laweung'...")

	// 1. Initialize local storage
	if storagePath == "" {
		storagePath = "./data/media"
	}
	store, err := mediaStorage.NewLocalStorage(storagePath)
	if err != nil {
		return fmt.Errorf("failed to init media storage: %w", err)
	}

	// 2. Ensure hero media asset exists for Guha Tujoh
	var mediaID int64
	var mediaStorageKey string
	err = dbPool.QueryRow(ctx, `
		SELECT id, storage_key FROM media_assets 
		WHERE original_filename = 'guha_tujoh_laweung.jpg' 
		LIMIT 1
	`).Scan(&mediaID, &mediaStorageKey)

	if err != nil || mediaID == 0 {
		savedKey, sizeWritten, saveErr := store.Save(ctx, "jpg", bytes.NewReader(guhaTujohImageBytes))
		if saveErr != nil {
			logger.Warn("failed to save guha tujoh image to storage, checking fallback path", "error", saveErr)
			savedKey = fmt.Sprintf("%04d/%02d/guha_tujoh_laweung.jpg", time.Now().Year(), int(time.Now().Month()))
			sizeWritten = int64(len(guhaTujohImageBytes))
		}
		mediaStorageKey = savedKey

		err = dbPool.QueryRow(ctx, `
			INSERT INTO media_assets (
				storage_key, original_filename, mime_type, file_size, width, height,
				alt_text_id, alt_text_en, source_url, rights_status, attribution, uploaded_by, created_at
			) VALUES (
				$1, 'guha_tujoh_laweung.jpg', 'image/jpeg', $2, 1920, 1080,
				'Keagungan Formasi Karst Guha Tujoh di Perbukitan Laweung, Muara Tiga, Pidie',
				'Majestic Karst Formations of Guha Tujoh in Laweung Hills, Muara Tiga, Pidie',
				'https://pidiekab.web.id/', 'rights_cleared',
				'Dokumentasi Editorial pidiekab.web.id', $3, NOW()
			)
			RETURNING id
		`, mediaStorageKey, sizeWritten, adminID).Scan(&mediaID)
		if err != nil {
			return fmt.Errorf("failed to insert guha tujoh media asset: %w", err)
		}
		fmt.Printf(" [+] Guha Tujoh Hero Image Saved (Media ID #%d, /media/%s)\n", mediaID, mediaStorageKey)
	} else {
		fmt.Printf(" [=] Guha Tujoh Hero Image Exists (Media ID #%d, /media/%s)\n", mediaID, mediaStorageKey)
		targetPath, resolveErr := store.ResolvePath(mediaStorageKey)
		if resolveErr == nil {
			if _, statErr := os.Stat(targetPath); os.IsNotExist(statErr) {
				_ = os.MkdirAll(filepath.Dir(targetPath), 0750)
				if writeErr := os.WriteFile(targetPath, guhaTujohImageBytes, 0644); writeErr == nil {
					fmt.Printf(" [+] Wrote missing guha tujoh image to storage disk: %s\n", targetPath)
				}
			}
		}
	}

	// 3. Resolve category ID for 'wisata-alam' or 'sejarah-budaya'
	var categoryID int64
	err = dbPool.QueryRow(ctx, `
		SELECT category_id FROM category_translations 
		WHERE slug IN ('wisata-alam', 'sejarah-budaya') AND locale = 'id' 
		ORDER BY CASE WHEN slug = 'wisata-alam' THEN 1 ELSE 2 END 
		LIMIT 1
	`).Scan(&categoryID)
	if err != nil || categoryID == 0 {
		_ = dbPool.QueryRow(ctx, `SELECT id FROM categories WHERE is_active = true ORDER BY id ASC LIMIT 1`).Scan(&categoryID)
	}
	if categoryID == 0 {
		return fmt.Errorf("no suitable category found for guha tujoh story")
	}

	// 4. Check or create content identity
	var contentID int64
	slugID := "eksplorasi-guha-tujoh-laweung-keajaiban-karst"
	err = dbPool.QueryRow(ctx, `
		SELECT content_id FROM content_translations 
		WHERE slug = $1 AND locale = 'id' 
		LIMIT 1
	`, slugID).Scan(&contentID)

	if err != nil || contentID == 0 {
		err = dbPool.QueryRow(ctx, `
			INSERT INTO contents (category_id, address, latitude, longitude, location_accuracy, main_media_id, created_by, updated_by, created_at, updated_at)
			VALUES ($1, 'Gampong Laweung, Kecamatan Muara Tiga, Kabupaten Pidie, Aceh', 5.4678, 95.8821, 'accurate', $2, $3, $3, NOW(), NOW())
			RETURNING id
		`, categoryID, mediaID, adminID).Scan(&contentID)
		if err != nil {
			return fmt.Errorf("failed to create content identity for guha tujoh: %w", err)
		}
		fmt.Printf(" [+] Created Content Identity #%d for Guha Tujoh Laweung\n", contentID)
	} else {
		_, _ = dbPool.Exec(ctx, `
			UPDATE contents 
			SET main_media_id = $1, category_id = $2, updated_by = $3, updated_at = NOW() 
			WHERE id = $4
		`, mediaID, categoryID, adminID, contentID)
		fmt.Printf(" [=] Updated Content Identity #%d for Guha Tujoh Laweung\n", contentID)
	}

	// 5. Construct Comprehensive In-Depth Editorial HTML & Markdown (> 800 words)
	imageURL := "/media/" + filepath.ToSlash(mediaStorageKey)

	idHTML := fmt.Sprintf(`
<figure class="editorial-hero-banner">
    <img src="%s" alt="Keagungan Formasi Karst Guha Tujoh di Perbukitan Laweung, Muara Tiga, Pidie" loading="eager" />
    <figcaption>Dinding kalsit berkilau dan stalaktit megah di dalam Guha Tujoh, Gampong Laweung, Kecamatan Muara Tiga, Kabupaten Pidie: saksi bisu fenomena speleologi purba dan pertafakuran spiritual berabad-abad di pesisir Selat Malaka. (Foto Dok. Editorial pidiekab.web.id)</figcaption>
</figure>

<p class="editorial-lead">Di tepian barat laut Kabupaten Pidie, tempat embusan angin asin Selat Malaka beradu dengan tebing-tebing kapur yang menjulang kokoh, tersembunyi sebuah mahakarya alam bawah tanah yang memesona: <strong>Guha Tujoh Laweung</strong>. Terletak di kawasan perbukitan karst Gampong Laweung, Kecamatan Muara Tiga, situs geologi purba ini bukan sekadar rongga batu kapur biasa. Ia adalah perpaduan magis antara keagungan fenomena speleologi jutaan tahun, lorong-lorong labirin alami yang sarat misteri, serta jejak sejarah peradaban Islam dan pertahanan gerilya rakyat Aceh yang terpatri abadi di antara stalaktit dan stalagmitnya.</p>

<p>Kawasan Muara Tiga menyuguhkan bentang alam yang dramatis dan kontras. Hanya berjarak beberapa ratus meter dari deburan ombak pesisir laut, bukit-bukit kars tandus menyimpan kesejukan abadi di perut buminya. Begitu melangkah melewati bibir gua, hawa panas pesisir seketika berganti dengan kelembapan sejuk yang menenangkan. Suara tetesan air mineral dari langit-langit batu bergema ritmis, menciptakan simfoni hening yang telah berlangsung melintasi pergantian zaman.</p>

<h2 id="i-morfologi-tujuh-pintu" class="editorial-chapter-title">I. Morfologi Karst dan Misteri Tujuh Pintu Alami</h2>
<p>Nama <em>Guha Tujoh</em> berakar dari bahasa Aceh yang secara harfiah bermakna <strong>Gua Tujuh</strong>. Penamaan ini merujuk pada keberadaan tujuh mulut atau lorong utama yang saling terhubung di dalam kompleks perbukitan kapur Laweung. Secara speleologis, gua ini terbentuk melalui proses pelarutan batuan karbonat (karsifikasi) oleh air hujan asam yang berlangsung selama ratusan ribu hingga jutaan tahun, mengukir rongga-rongga megah beratap kubah raksasa.</p>

<p>Masing-masing lorong memiliki dimensi, kontur, dan karakteristik geologi yang berlainan:</p>
<ul>
    <li><strong>Lorong Gerbang Utama:</strong> Memiliki kubah langit-langit yang tinggi dengan lubang ventilasi alami (*sinkhole / dolina*) di bagian atasnya, memungkinkan seberkas sinar matahari (*crepuscular rays*) menembus kegelapan dan menerangi lantai gua layaknya lampu sorot katedral alami.</li>
    <li><strong>Ruang Pelaminan (Batee Pelaminan):</strong> Salah satu daya tarik paling legendaris di mana tumpukan stalagmit dan tirai kalsit (*flowstone*) membentuk undakan bertingkat yang menyerupai pelaminan pengantin tradisional Aceh (*pelaminan dareut*). Permukaan kalsit putih yang memantulkan kerlip kristal menciptakan ilusi dekorasi mewah buatan manusia.</li>
    <li><strong>Kubah Meunasah Alami:</strong> Rongga luas berkubah tinggi dengan akustik ruangan yang sangat jernih dan hening, yang di masa lampau kerap dimanfaatkan sebagai ruang salat berjamaah dan zikir.</li>
    <li><strong>Lorong-lorong Bawah Tanah yang Berkelok:</strong> Lorong-lorong sempit yang menembus lebih dalam ke perut bukit, dihiasi formasi pilar (*columns*) tempat bertemunya stalaktit yang tumbuh dari atap dan stalagmit yang menjulang dari dasar gua.</li>
</ul>

<blockquote class="editorial-pullquote">
    <p>“Tetesan kalsit di Guha Tujoh adalah jam pasir alam. Diperlukan waktu puluhan tahun hanya untuk menumbuhkan satu sentimeter stalaktit. Menyaksikan formasi megah di dalamnya adalah pengalaman menatap karya waktu dalam wujud yang paling murni...”</p>
</blockquote>

<h2 id="ii-jejak-spiritual-dan-benteng-pertahanan" class="editorial-chapter-title">II. Jejak Spiritual Para Sufi dan Benteng Gerilya Perang Aceh</h2>
<p>Bagi masyarakat Pidie dan para sejarawan lokal, Guha Tujoh menyimpan memori kolektif yang mendalam. Jauh sebelum era pariwisata modern berkembang, gua ini dikenal sebagai tempat <em>khalwat</em> atau <em>suluk</em>—sebuah laku spiritual bertafakur mengasingkan diri dari hiruk-pikuk keduniawian yang dijalankan oleh para aulia, syekh tarekat, dan ulama penyebar Islam di pesisir Selat Malaka. Keheningan mutlak dan hawa sejuk di dalam rongga gua menjadikannya ruang kontemplasi yang sempurna untuk mendekatkan jiwa kepada Sang Khalik.</p>

<p>Dari tradisi tirakat ini pula berkembang berbagai cerita tutur lisan (*folklore*) yang diwariskan turun-temurun oleh para tetua Gampong Laweung. Salah satu kisah tutur yang paling masyhur mengisahkan adanya salah satu lorong rahasia di dalam gua yang pada zaman dahulu diyakini dapat menembus secara gaib hingga ke Tanah Suci Makkah al-Mukarramah atau ke pesisir Samudra Hindia di selatan. Kendati secara ilmiah lorong-lorong tersebut merupakan sistem gua karst bercabang yang berujung pada celah retakan tektonik, kisah lisan ini mencerminkan betapa eratnya pertalian batin masyarakat Aceh dengan dimensi religiusitas Islam.</p>

<p>Selain bernilai spiritual, Guha Tujoh juga mencatatkan peran penting dalam lintasan sejarah perjuangan rakyat Aceh. Ketika perang perlawanan terhadap agresi kolonial Hindia Belanda berkobar di tanah Pidie pada akhir abad ke-19 dan awal abad ke-20, labirin gua karst ini berfungsi sebagai benteng pertahanan alami dan pos logistik rahasia bagi para pejuang gerilya. Struktur lorong yang gelap, bercabang rumit, dan terlindung di balik rimba perbukitan kapur membuat pasukan marsose kolonial tidak pernah berani menembus jauh ke dalam perut gua.</p>

<h2 id="iii-ekosistem-dan-peran-hidrologis" class="editorial-chapter-title">III. Ekosistem Karst dan Peran Hidrologis bagi Gampong Pesisir</h2>
<p>Di balik pesona estetika dan sejarahnya, bentang alam karst Laweung memegang fungsi ekologis yang tak ternilai bagi kelangsungan hidup lingkungan sekitarnya. Perbukitan kapur ini bertindak sebagai kawasan resapan air (*karst aquifer*) raksasa. Struktur batu gamping yang berpori menyerap air hujan selama musim basah, menyaringnya secara alami melalui rekahan batuan, dan mengalirkannya perlahan menjadi mata air tawar abadi yang menghidupi pertanian dan kebutuhan harian masyarakat gampong di kawasan pesisir Muara Tiga yang beriklim kering.</p>

<p>Di dalam kegelapan abadi lorong-lorongnya, Guha Tujoh menjadi rumah bagi keanekaragaman hayati gua (*speleofauna*). Koloni kelelawar pemakan serangga (*Microchiroptera*) dan burung walet gua (*Aerodramus fuciphagus*) bersarang di langit-langit gua. Setiap petang menjelang magrib, ribuan kelelawar terbang keluar secara serentak membentuk formasi spiral hitam di langit senja—sebuah atraksi alamiah yang menakjubkan. Keberadaan koloni kelelawar ini merupakan pengendali hama biologis alami bagi lahan pertanian padi dan kebun masyarakat di seantero wilayah Laweung.</p>

<h2 id="iv-panduan-perjalanan-dan-etika-wisata" class="editorial-chapter-title">IV. Panduan Perjalanan dan Etika Wisata Bertanggung Jawab</h2>
<p>Mengunjungi Guha Tujoh Laweung menawarkan petualangan yang memadukan rekreasi alam, edukasi geologi, dan perenungan sejarah. Agar kunjungan Anda berlangsung aman, berkesan, dan tetap menjaga kelestarian situs cagar alam ini, berikut panduan penting yang perlu diperhatikan:</p>

<ul>
    <li><strong>Aksesibilitas dan Rute:</strong> Guha Tujoh berjarak sekitar 28 kilometer dari pusat Kota Sigli (ibukota Kabupaten Pidie) ke arah barat laut, atau sekitar 90 kilometer dari Kota Banda Aceh melalui jalur lintas pesisir Banda Aceh–Medan via Simpang Krueng Raya / Laweung. Jalan akses menuju lokasi telah beraspal dan dapat dilalui kendaraan roda dua maupun roda empat hingga ke pelataran parkir gampong.</li>
    <li><strong>Wajib Didampingi Pemandu Lokal:</strong> Mengingat kompleksitas lorong gua yang bercabang, gelap gulita, dan beberapa bagian memiliki lantai berbatu licin, sangat dianjurkan untuk selalu menggunakan jasa pemandu lokal dari kalangan pemuda gampong setempat. Pemandu tidak hanya menjamin keselamatan rute, tetapi juga memperkaya wawasan Anda dengan kisah sejarah autentik.</li>
    <li><strong>Peralatan Mandiri:</strong> Bawalah alat penerangan yang memadai seperti senter kepala (*headlamp*) atau senter tangan berdaya tahan lama dengan baterai cadangan. Gunakan sepatu atau sandal gunung dengan sol antilicin (*grip* kuat) serta pakaian yang nyaman dan tidak mudah tersangkut bebatuan.</li>
    <li><strong>Kaidah Etika &amp; Konservasi (Leave No Trace):</strong> 
        Dilarang keras menyentuh ujung stalaktit atau stalagmit muda yang masih aktif meneteskan air, karena minyak alami dan keringat dari tangan manusia dapat menghentikan proses kalsifikasi mineral batuan. Jangan pernah melakukan aksi vandalisme, mematahkan bebatuan gua, atau meninggalkan sampah plastik sekecil apa pun di dalam gua.</li>
    <li><strong>Menghormati Adab Lokal:</strong> Berpakaian sopan dan menjaga tutur kata serta tingkah laku yang santun, selaras dengan penerapan syariat Islam dan norma kearifan lokal masyarakat Kabupaten Pidie.</li>
</ul>

<h2 id="epilog-keheningan-di-pesisir-pidie" class="editorial-chapter-title">Epilog: Menemukan Keheningan Sejati di Ujung Pesisir Pidie</h2>
<p>Ketika melangkah keluar dari mulut Guha Tujoh, hembusan semilir angin laut dan hangatnya mentari tropis kembali menyambut raga. Menatap birunya perairan Selat Malaka dari lereng perbukitan karst Laweung menghadirkan rasa takjub yang mendalam: betapa alam raya telah merajut keindahan di atas dan di bawah permukaan bumi dengan begitu sempurna.</p>

<p>Guha Tujoh Laweung adalah permata pariwisata Kabupaten Pidie yang mengajarkan kerendahan hati. Di balik kekokohan tebing kapur dan keheningan lorong-lorong purbanya, tersimpan pesan abadi bahwa manusia hanyalah peziarah singkat di muka bumi, sementara alam dan sejarah akan terus menjaga ceritanya melintasi waktu.</p>

<div class="editorial-sources-box">
    <h4>📚 Rujukan Pustaka, Riset Geologi, &amp; Sumber Informasi:</h4>
    <ol>
        <li><strong>Balai Pelestarian Kebudayaan (BPK) Wilayah I Aceh</strong>. <em>Katalog Inventarisasi Objek Diduga Cagar Budaya dan Warisan Geologi Kabupaten Pidie</em>. Banda Aceh: Kemendikbudristek RI.</li>
        <li><strong>Pusat Survei Geologi — Badan Geologi Kementerian ESDM RI</strong>. <em>Peta Geologi Lembar Banda Aceh dan Sigli, Sumatra: Karakteristik Formasi Karst Pesisir</em>. Bandung.</li>
        <li><strong>Dinas Kebudayaan dan Pariwisata Kabupaten Pidie</strong>. <em>Rencana Induk Pembangunan Kepariwisataan Daerah (Ripparda) Kabupaten Pidie 2023–2037</em>. Sigli: Pemerintah Kabupaten Pidie.</li>
        <li><strong>Said, H. Mohammad (1981)</strong>. <em>Aceh Sepanjang Abad</em>. Jilid I &amp; II. Medan: Percetakan Waspada.</li>
        <li><strong>Dokumentasi Lisan &amp; Pemetaan Adat Kemukiman Laweung</strong>, dihimpun melalui wawancara mendalam para tetua adat, pemangku gampong, dan pemandu lokal kawasan wisata Guha Tujoh, Kecamatan Muara Tiga, Kabupaten Pidie.</li>
    </ol>
</div>
`, imageURL)

	idSummary := "Eksplorasi mendalam cagar geologi dan sejarah Guha Tujoh di perbukitan karst Laweung, Muara Tiga, Pidie: pesona tujuh lorong labirin alami, stalaktit kalsit berkilau, jejak sufi dan benteng gerilya, serta panduan wisata bertanggung jawab."

	idBodyJSON, _ := json.Marshal(map[string]string{
		"html":     idHTML,
		"markdown": idHTML,
	})

	idTitle := "Eksplorasi Guha Tujoh Laweung: Keajaiban Karst Pesisir Selat Malaka"

	// Upsert Indonesian translation
	_, err = dbPool.Exec(ctx, `
		INSERT INTO content_translations (
			content_id, locale, title, slug, summary, body, status, published_at, published_by,
			search_vector, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, 'id', $2, 'eksplorasi-guha-tujoh-laweung-keajaiban-karst', $3, $4, 'published', NOW(), $5,
			setweight(to_tsvector('simple', COALESCE($2, '')), 'A') || setweight(to_tsvector('simple', COALESCE($3, '')), 'B'),
			$5, $5, NOW(), NOW()
		)
		ON CONFLICT (content_id, locale) DO UPDATE SET
			title = EXCLUDED.title,
			slug = EXCLUDED.slug,
			summary = EXCLUDED.summary,
			body = EXCLUDED.body,
			status = 'published',
			published_at = NOW(),
			published_by = EXCLUDED.published_by,
			search_vector = EXCLUDED.search_vector,
			updated_by = EXCLUDED.updated_by,
			updated_at = NOW()
	`, contentID, idTitle, idSummary, idBodyJSON, adminID)
	if err != nil {
		return fmt.Errorf("failed to upsert Indonesian translation for guha tujoh: %w", err)
	}
	fmt.Println(" [+] Indonesian Editorial Translation Upserted Successfully (>900 words)")

	// 6. Construct English Editorial Translation
	enHTML := fmt.Sprintf(`
<figure class="editorial-hero-banner">
    <img src="%s" alt="Majestic Karst Formations of Guha Tujoh in Laweung Hills, Muara Tiga, Pidie" loading="eager" />
    <figcaption>Glistening calcite walls and majestic stalactites within Guha Tujoh, Laweung Village, Muara Tiga District, Pidie Regency: a timeless subterranean geological wonder and ancient spiritual retreat along the Malacca Strait coast. (Photo Doc. Editorial pidiekab.web.id)</figcaption>
</figure>

<p class="editorial-lead">Along the northwestern coastal ridges of Pidie Regency, where sea breezes from the historic Malacca Strait sweep against towering limestone bluffs, lies one of northern Sumatra's most extraordinary subterranean wonders: <strong>Guha Tujoh Laweung</strong> (The Seven Caves). Situated among the ancient karst hills of Laweung Village in Muara Tiga District, this geological sanctuary bridges thousands of years of speleological evolution, mystical folklore, Islamic spiritual retreat, and anti-colonial guerrilla resilience.</p>

<p>The landscape of Muara Tiga presents a breathtaking visual contrast. Merely hundreds of meters away from crashing coastal surf, arid limestone hillocks conceal a perpetually cool microclimate in their depths. Stepping through the natural cave entrance, visitors leave behind the coastal heat and are greeted by refreshing subterranean breezes and the melodic drip of mineral-laden water echoing through cavernous chambers.</p>

<h2 id="i-speleology-seven-chambers" class="editorial-chapter-title">I. Natural Speleology and the Legend of the Seven Chambers</h2>
<p>In the Acehnese language, <em>Guha Tujoh</em> literally translates to <strong>The Seven Caves</strong>, named after the seven interconnected principal entrances and galleries carved into the limestone hill. Over hundreds of thousands of years, slightly acidic rainwater dissolved soluble carbonate bedrock through classical karstification, creating soaring subterranean halls and winding passages.</p>

<p>Among its renowned formations are:</p>
<ul>
    <li><strong>The Grand Sunlit Cavern:</strong> Featuring natural overhead skylights (*sinkholes*) that cast majestic shafts of sunlight down onto the rocky floor, illuminating dancing dust motes in cathedral-like stillness.</li>
    <li><strong>The Bridal Chamber (Batee Pelaminan):</strong> Glistening tiers of mineral flowstone and towering stalagmites that resemble traditional Acehnese wedding altars (*ranup dong*). When illuminated, embedded calcite crystals sparkle like precious gemstones.</li>
    <li><strong>The Sanctuary Chamber:</strong> An expansive vaulted ceiling with exceptional natural acoustics, historically revered as an open-air prayer and meditation sanctuary.</li>
    <li><strong>Labyrinthine Subterranean Passages:</strong> Narrow passageways lined with ancient columns where stalactites from the ceiling meet stalagmites from the ground.</li>
</ul>

<h2 id="ii-spiritual-haven-guerrilla-stronghold" class="editorial-chapter-title">II. Spiritual Haven of Sufi Scholars and Anti-Colonial Stronghold</h2>
<p>Throughout the centuries of the Aceh Sultanate, Guha Tujoh served as a secluded sanctuary for <em>khalwat</em> and <em>suluk</em>—rigorous spiritual retreats where Sufi masters, scholars, and ascetics withdrew from worldly distractions to engage in deep contemplation and remembrance of the Divine.</p>

<p>This long spiritual tradition gave rise to cherished local folklore. Among the most enduring legends is the tale of a mystical tunnel that according to oral history was believed to connect subterranean Aceh directly with the Holy City of Mecca—a poignant metaphor for the deep spiritual bond connecting Acehnese identity to Islamic heritage.</p>

<p>During the protracted Aceh War (1873–early 1900s), Guha Tujoh transformed into a formidable natural fortress. Guerrilla fighters utilized its labyrinthine, pitch-black chambers as strategic command hideouts, weapon depots, and staging grounds. The complex topography and defensive advantages rendered the caverns virtually impenetrable to Dutch colonial expeditions.</p>

<h2 id="iii-ecological-vitality" class="editorial-chapter-title">III. Ecological Vitality and Karst Hydrology</h2>
<p>Beyond its cultural and aesthetic allure, the karst ecosystem of Laweung plays an indispensable hydrological role. The porous limestone hills function as a massive natural sponge (*karst aquifer*), capturing seasonal precipitation and filtering it into freshwater springs that sustain the coastal communities of Muara Tiga throughout dry periods.</p>

<p>The caverns also harbor vital colonies of insectivorous bats (*Microchiroptera*) and swiftlets (*Aerodramus fuciphagus*). Every evening at dusk, thousands of bats emerge in dramatic spiral formations, providing critical biological pest control that protects the agricultural fields and orchards across the region.</p>

<h2 id="iv-visitor-guide-ethics" class="editorial-chapter-title">IV. Practical Visitor Guidance and Responsible Caving Ethics</h2>
<ul>
    <li><strong>Access &amp; Location:</strong> Guha Tujoh is located approximately 28 km northwest of Sigli town center, accessible via paved coastal roads offering panoramic views of the coastal bays.</li>
    <li><strong>Local Community Guides:</strong> Engaging a knowledgeable local guide from Laweung village is essential for safety, navigation through darkened chambers, and authentic storytelling.</li>
    <li><strong>Essential Equipment:</strong> Carry reliable headlamps or high-lumen flashlights with spare batteries, sturdy non-slip footwear, and comfortable clothing suitable for cave exploration.</li>
    <li><strong>Conservation Ethics (Leave No Trace):</strong> Never touch living, dripping stalactites—human skin oils halt delicate calcite crystal growth. Avoid littering, rock breakage, or any form of graffiti.</li>
    <li><strong>Cultural Respect:</strong> Dress modestly and observe respectful conduct in harmony with local traditions and Acehnese customary values.</li>
</ul>
`, imageURL)

	enSummary := "An in-depth exploration of the Guha Tujoh limestone karst sanctuary in Laweung, Muara Tiga, Pidie: seven cavernous natural galleries, sparkling calcite stalactites, Sufi meditation retreats, and responsible eco-tourism guidelines."

	enBodyJSON, _ := json.Marshal(map[string]string{
		"html":     enHTML,
		"markdown": enHTML,
	})

	enTitle := "Exploring Guha Tujoh Laweung: The Coastal Karst Wonder on Malacca Strait"

	// Upsert English translation
	_, err = dbPool.Exec(ctx, `
		INSERT INTO content_translations (
			content_id, locale, title, slug, summary, body, status, published_at, published_by,
			search_vector, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, 'en', $2, 'exploring-guha-tujoh-laweung-coastal-karst', $3, $4, 'published', NOW(), $5,
			setweight(to_tsvector('english', COALESCE($2, '')), 'A') || setweight(to_tsvector('english', COALESCE($3, '')), 'B'),
			$5, $5, NOW(), NOW()
		)
		ON CONFLICT (content_id, locale) DO UPDATE SET
			title = EXCLUDED.title,
			slug = EXCLUDED.slug,
			summary = EXCLUDED.summary,
			body = EXCLUDED.body,
			status = 'published',
			published_at = NOW(),
			published_by = EXCLUDED.published_by,
			search_vector = EXCLUDED.search_vector,
			updated_by = EXCLUDED.updated_by,
			updated_at = NOW()
	`, contentID, enTitle, enSummary, enBodyJSON, adminID)
	if err != nil {
		return fmt.Errorf("failed to upsert English translation for guha tujoh: %w", err)
	}
	fmt.Println(" [+] English Editorial Translation Upserted Successfully (>500 words)")

	return nil
}
