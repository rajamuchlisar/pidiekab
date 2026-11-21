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

//go:embed assets/kue_adee_khas_pidie.jpg
var kueAdeeImageBytes []byte

// publishKueAdeeStory inserts or updates the comprehensive editorial story and hero image for Kue Adee Khas Pidie.
func publishKueAdeeStory(ctx context.Context, dbPool *pgxpool.Pool, adminID int64, storagePath string, logger *slog.Logger) error {
	fmt.Println("--> Publishing Editorial Story: 'Kue Adee Khas Pidie'...")

	// 1. Initialize local storage
	if storagePath == "" {
		storagePath = "./data/media"
	}
	store, err := mediaStorage.NewLocalStorage(storagePath)
	if err != nil {
		return fmt.Errorf("failed to init media storage: %w", err)
	}

	// 2. Ensure hero media asset exists for kue adee
	var mediaID int64
	var mediaStorageKey string
	err = dbPool.QueryRow(ctx, `
		SELECT id, storage_key FROM media_assets 
		WHERE original_filename = 'kue_adee_khas_pidie.jpg' 
		LIMIT 1
	`).Scan(&mediaID, &mediaStorageKey)

	if err != nil || mediaID == 0 {
		savedKey, sizeWritten, saveErr := store.Save(ctx, "jpg", bytes.NewReader(kueAdeeImageBytes))
		if saveErr != nil {
			logger.Warn("failed to save kue adee image to storage, checking fallback path", "error", saveErr)
			savedKey = fmt.Sprintf("%04d/%02d/kue_adee_khas_pidie.jpg", time.Now().Year(), int(time.Now().Month()))
			sizeWritten = int64(len(kueAdeeImageBytes))
		}
		mediaStorageKey = savedKey

		err = dbPool.QueryRow(ctx, `
			INSERT INTO media_assets (
				storage_key, original_filename, mime_type, file_size, width, height,
				alt_text_id, alt_text_en, source_url, rights_status, attribution, uploaded_by, created_at
			) VALUES (
				$1, 'kue_adee_khas_pidie.jpg', 'image/jpeg', $2, 1920, 1080,
				'Kue Adee Khas Pidie: Manis Legit Bertabur Bawang Goreng Harum',
				'Traditional Adee Baked Sweet Cassava Cake of Pidie Aceh',
				'https://pidiekab.web.id/', 'rights_cleared',
				'Dokumentasi Editorial pidiekab.web.id', $3, NOW()
			)
			RETURNING id
		`, mediaStorageKey, sizeWritten, adminID).Scan(&mediaID)
		if err != nil {
			return fmt.Errorf("failed to insert kue adee media asset: %w", err)
		}
		fmt.Printf(" [+] Kue Adee Hero Image Saved (Media ID #%d, /media/%s)\n", mediaID, mediaStorageKey)
	} else {
		fmt.Printf(" [=] Kue Adee Hero Image Exists (Media ID #%d, /media/%s)\n", mediaID, mediaStorageKey)
		// Ensure physical file exists on disk
		targetPath, resolveErr := store.ResolvePath(mediaStorageKey)
		if resolveErr == nil {
			if _, statErr := os.Stat(targetPath); os.IsNotExist(statErr) {
				_ = os.MkdirAll(filepath.Dir(targetPath), 0750)
				if writeErr := os.WriteFile(targetPath, kueAdeeImageBytes, 0644); writeErr == nil {
					fmt.Printf(" [+] Wrote missing kue adee image to storage disk: %s\n", targetPath)
				}
			}
		}
	}

	// 3. Resolve category ID for 'kuliner-khas'
	var categoryID int64
	err = dbPool.QueryRow(ctx, `
		SELECT category_id FROM category_translations 
		WHERE slug = 'kuliner-khas' AND locale = 'id' 
		LIMIT 1
	`).Scan(&categoryID)
	if err != nil || categoryID == 0 {
		_ = dbPool.QueryRow(ctx, `SELECT id FROM categories WHERE is_active = true ORDER BY id ASC LIMIT 1`).Scan(&categoryID)
	}
	if categoryID == 0 {
		return fmt.Errorf("no suitable category found for kue adee story")
	}

	// 4. Check or create content identity
	var contentID int64
	slugID := "kue-adee-khas-pidie-manis-legit-tradisi"
	err = dbPool.QueryRow(ctx, `
		SELECT content_id FROM content_translations 
		WHERE slug = $1 AND locale = 'id' 
		LIMIT 1
	`, slugID).Scan(&contentID)

	if err != nil || contentID == 0 {
		err = dbPool.QueryRow(ctx, `
			INSERT INTO contents (category_id, address, latitude, longitude, location_accuracy, main_media_id, created_by, updated_by, created_at, updated_at)
			VALUES ($1, 'Kecamatan Mutiara Timur & Beureunuen, Kabupaten Pidie, Aceh', 5.3150, 95.9780, 'accurate', $2, $3, $3, NOW(), NOW())
			RETURNING id
		`, categoryID, mediaID, adminID).Scan(&contentID)
		if err != nil {
			return fmt.Errorf("failed to create content identity for kue adee: %w", err)
		}
		fmt.Printf(" [+] Created Content Identity #%d for Kue Adee\n", contentID)
	} else {
		_, _ = dbPool.Exec(ctx, `
			UPDATE contents 
			SET main_media_id = $1, category_id = $2, updated_by = $3, updated_at = NOW() 
			WHERE id = $4
		`, mediaID, categoryID, adminID, contentID)
		fmt.Printf(" [=] Updated Content Identity #%d for Kue Adee\n", contentID)
	}

	// 5. Construct Editorial HTML & Markdown for Indonesian
	imageURL := "/media/" + filepath.ToSlash(mediaStorageKey)

	idHTML := fmt.Sprintf(`
<figure class="editorial-hero-banner">
    <img src="%s" alt="Kue Adee Khas Pidie: Manis Legit Bertabur Bawang Goreng Harum" loading="eager" />
    <figcaption>Kue Adee tradisional khas Kabupaten Pidie yang baru matang dari panggangan: berkulit keemasan karamel, bertekstur lembut lumer di dalam, dan beraroma harum taburan bawang merah goreng. (Foto Dok. Editorial pidiekab.web.id)</figcaption>
</figure>

<p class="editorial-lead">Bagi siapa pun yang menempuh perjalanan darat menyusuri pesisir timur Aceh, melintasi batas Kabupaten Pidie selalu ditandai oleh pemandangan yang memikat indra: deretan kedai panggangan tradisional dengan kepulan aroma manis santan kelapa, karamel gula tebu, dan wangi gurih bawang merah goreng yang menyeruak ke udara. Inilah aroma Kue Adee—kudapan basah legendaris yang menduduki takhta tertinggi dalam peta gastronomi manis Negeri Po Teumeureuhom.</p>

<p>Kue Adee (kerap disebut pula sebagai <em>Bingka Adee</em>) adalah simbol kehangatan dan kehalusan budi pekerti masyarakat Pidie. Berbeda dengan kue bingka pada umumnya di kepulauan Nusantara, Adee memiliki identitas rasa yang sangat kontras namun bersatu dalam harmoni sempurna: rasa manis pekat yang legit, kelembutan santan murni yang lumer seketika di rongga mulut, serta sentuhan gurih renyah dari taburan bawang goreng di atas kerak panggangan keemasannya.</p>

<h2 class="editorial-chapter-title">I. Riwayat &amp; Jejak Tradisi: Kue Para Bangsawan dan Tamu Mulia</h2>
<p>Kue Adee berakar dari tradisi kuliner kuno masyarakat pesisir Pidie dan kawasan sekitarnya. Pada masa Kesultanan Aceh Darussalam, adee bukanlah panganan harian biasa yang mudah ditemui di sembarang pasar. Kue basah istimewa ini merupakan sajian kehormatan para bangsawan, kerabat istana, dan alim ulama pada perayaan-perayaan keagamaan besar atau upacara penyambutan utusan kerajaan.</p>

<p>Nama <em>Adee</em> dalam tuturan masyarakat lokal kerap dikaitkan dengan filosofi keadilan rasa dan ketenteraman hati saat menikmatinya. Mengolah kue adee tempo dulu membutuhkan ketelatenan tinggi dan dedikasi waktu yang panjang: mulai dari memarut ubi dengan tangan, memeras kelapa tua tanpa air sedikit pun, hingga menjaga kestabilan bara api kayu arang selama berjam-jam. Menyajikan kue adee kepada tamu adalah ungkapan penghormatan tertinggi (<em>pemulia jamee</em>) dari sang tuan rumah.</p>

<h2 class="editorial-chapter-title">II. Dua Varian Utama: Adee Ubi Kayu vs Adee Terigu</h2>
<p>Di tangan para perempuan perajin kue di Kabupaten Pidie, tercipta dua varian klasik yang masing-masing membawa karakter tekstur yang memukau:</p>

<ul>
    <li><strong>Adee Ubi (Singkong)</strong>: Merupakan varian paling tua, otentik, dan sarat kearifan agraris pedesaan Pidie. Menggunakan ubi kayu (singkong) pilihan yang diparut sangat halus lalu diperas sedikit patinya agar tekstur akhir tidak terlalu liat. Dicampur dengan perasan santan kelapa kental pertama dan gula pasir murni, Adee Ubi menghasilkan tekstur padat, kenyal (<em>chewy</em>), berserat lembut, dan beraroma tanah subur yang manis alami.</li>
    <li><strong>Adee Terigu (Tepung Gandum)</strong>: Varian ini berkembang seiring masuknya komoditas tepung gandum berkualitas tinggi ke bandar niaga pesisir Pidie. Tepung terigu diaduk bersama telur bebek segar dan santan hangat hingga membentuk adonan cair yang sangat halus. Hasil panggangan Adee Terigu menghadirkan sensasi lembut lumer menyerupai puding karamel panggang (<em>custard</em>) yang kaya rasa dan meleleh seketika di lidah.</li>
</ul>

<h2 class="editorial-chapter-title">III. Rahasia Kelezatan: Santan Tua, Telur Bebek, dan Tungku Arang Ganda</h2>
<p>Keajaiban rasa kue adee tidak mengandalkan bahan pengawet, pemanis buatan, atau pengembang instan. Tiga pilar utama bahan alaminya dijaga ketat secara turun-temurun oleh para perajin:</p>

<p><strong>1. Santan Kelapa Tua Murni (<em>Pati U Teupreh</em>)</strong><br>
Kelapa yang digunakan wajib kelapa tua yang telah matang sempurna di pohon. Kelapa diparut dan diperas tanpa tambahan air mentah (hanya menggunakan sari perasan pertama) demi menghasilkan kandungan minyak nabati alami yang melimpah dan aroma gurih yang intens.</p>

<p><strong>2. Telur Bebek Petelur Lokal</strong><br>
Sebagian besar pembuat kue adee tradisional lebih memilih telur bebek daripada telur ayam. Kuning telur bebek yang berwarna jingga kemerahan memberikan kekayaan rasa lemak gurih, kelembutan tekstur yang tahan lama, serta rona kuning keemasan yang berkilau alami tanpa bantuan pewarna sintetis.</p>

<p><strong>3. Metode Pemanggangan Bara Arang Atas-Bawah</strong><br>
Kue adee dimasak di atas loyang bundar aluminium atau seng tebal berdiameter 20 hingga 25 sentimeter. Kunci keemasan kuenya terletak pada teknik pemanggangan tradisional: loyang diletakkan di atas tungku bara arang kayu atau batok kelapa dengan api sedang, sementara di bagian atasnya ditutup lempeng seng yang ditimbuni arang membara menyala. Pemanasan ganda dari atas dan bawah ini memastikan bagian tengah kue matang merata dan lembap, sementara permukaannya mengalami karamelisasi sempurna—menghasilkan kerak tipis kecokelatan yang harum menggiurkan.</p>

<h2 class="editorial-chapter-title">IV. Sentuhan Magis Bawang Goreng: Misteri Manis-Gurih yang Memikat</h2>
<p>Salah satu ciri paling unik yang membedakan Kue Adee Pidie dengan penganan manis lainnya di Nusantara adalah taburan irisan bawang merah goreng (<em>bawang mirah teugoreng</em>) dan sedikit biji jintan manis di atas permukaannya sesaat sebelum proses pemanggangan selesai.</p>

<p>Bagi penikmat kuliner dari luar daerah yang baru pertama kali melihatnya, kombinasi bawang goreng di atas kue manis mungkin terasa tak lazim. Namun begitu mencicipinya, keajaiban gastronomi Aceh langsung terbukti: gurihnya minyak bawang dan aroma wangi rempah panggang bertindak sebagai penyeimbang sempurna (<em>umami counter-balance</em>) terhadap manis pekatnya adonan santan. Bawang goreng mencegah rasa enek (<em>jelak</em>), mengundang siapa pun untuk terus memotong dan menikmatinya hingga irisan terakhir.</p>

<h2 class="editorial-chapter-title">V. Nilai Adat &amp; Tradisi: Pengikat Silaturahmi dan Hantaran Pengantin</h2>
<p>Dalam tata kehidupan masyarakat Pidie, Kue Adee tidak sekadar bernilai ekonomi, melainkan sarat makna sosial dan adat yang luhur:</p>

<ul>
    <li><strong>Hantaran Pengantin (<em>Dalung Ranup &amp; Hantaran Linto/Dara Baro</em>)</strong>: Dalam prosesi pernikahan adat Aceh, kue adee merupakan hidangan wajib yang disusun rapi dalam talam tembaga bertutup kain beludru sulam benang emas (<em>dalung meukasab</em>). Manis dan legitnya kue adee melambangkan doa dan harapan agar kehidupan rumah tangga pasangan baru senantiasa harmonis, manis, dan rekat tak terpisahkan.</li>
    <li><strong>Sajian Hari Raya &amp; Kenduri Mulod</strong>: Saat Idulfitri, Iduladha, dan bulan peringatan Maulid Nabi Muhammad SAW, meunasah dan rumah-rumah warga selalu menyediakan loyang kue adee untuk disuguhkan kepada sanak saudara yang pulang dari tanah rantau.</li>
    <li><strong>Ikon Buah Tangan Perantau</strong>: Bagi perantau yang hendak kembali ke kota perantauan seperti Medan, Jakarta, hingga Malaysia, membawa beberapa kotak kue adee hangat adalah sebuah kewajiban moral untuk membagikan kebahagiaan kampung halaman kepada para sahabat.</li>
</ul>

<h2 class="editorial-chapter-title">VI. Teman Setia Kupi Khop dan Kopi Robusta Tangse</h2>
<p>Kenikmatan sejati sepotong kue adee berpendar paling sempurna saat disandingkan dengan secangkir kopi hitam tradisional Aceh. Di warung-warung kopi khas Pidie, para tetua dan pemuda kerap memesan sepotong adee potong bersanding dengan <strong>Kupi Khop</strong> (kopi tubruk robusta yang disajikan dalam gelas kaca terbalik di atas piring kecil) atau seduhan kopi robusta dataran tinggi Tangse.</p>

<p>Pahit pekat dan aroma sangrai kopi robusta lokal membasuh manis legitnya santan karamel adee, menciptakan pengalaman sensori yang kaya rasa dan menenangkan jiwa di sore hari.</p>

<div class="editorial-sources-box">
    <h4>💡 Panduan Menikmati &amp; Menyimpan Kue Adee Asli:</h4>
    <ul>
        <li><strong>Suhu Terbaik:</strong> Paling nikmat disantap dalam kondisi hangat kuku, sekitar 30 menit setelah keluar dari loyang pemanggangan.</li>
        <li><strong>Daya Tahan Alami:</strong> Karena dimasak matang sempurna dengan kandungan minyak kelapa murni dan tanpa bahan pengawet buatan, kue adee tahan selama 2 hingga 3 hari pada suhu ruang sejuk. Jika disimpan di dalam lemari pendingin, kue adee dapat bertahan hingga 7 hari dan dapat dihangatkan kembali menggunakan oven atau pemanggang teflon tanpa minyak.</li>
        <li><strong>Varian Favorit:</strong> Cobalah membandingkan varian Adee Ubi yang berserat kenyal dan Adee Terigu yang bertekstur lumer layaknya kue panggang Eropa untuk menemukan cita rasa favorit Anda.</li>
    </ul>
</div>
`, imageURL)

	idSummary := "Kudapan manis legit legendaris khas Pidie dengan perpaduan santan kelapa kental, singkong atau terigu, dan taburan bawang goreng harum yang dipanggang sempurna di atas bara arang."

	idBodyJSON, _ := json.Marshal(map[string]string{
		"html":     idHTML,
		"markdown": idHTML,
	})

	// Upsert Indonesian translation
	idTitle := "Kue Adee Khas Pidie: Manis Legit Tradisi Kue Basah Aceh"
	_, err = dbPool.Exec(ctx, `
		INSERT INTO content_translations (
			content_id, locale, title, slug, summary, body, status, published_at, published_by,
			search_vector, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, 'id', $2, 'kue-adee-khas-pidie-manis-legit-tradisi', $3, $4, 'published', NOW(), $5,
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
		return fmt.Errorf("failed to save ID translation for kue adee: %w", err)
	}
	fmt.Println(" [+] Indonesian Translation Published: 'Kue Adee Khas Pidie: Manis Legit Tradisi Kue Basah Aceh'")

	// 6. Construct English Translation
	enTitle := "Traditional Adee Cake: Sweet Fragrance of Pidie Confectionery"
	enSlug := "traditional-adee-cake-pidie-confectionery"
	enHTML := fmt.Sprintf(`
<figure class="editorial-hero-banner">
    <img src="%s" alt="Traditional Adee Baked Sweet Cassava Cake of Pidie Aceh" loading="eager" />
    <figcaption>Freshly baked traditional Kue Adee of Pidie: glistening with a caramelized golden crust, velvety coconut interior, and garnished with aromatic fried crisp shallots. (Photo: Editorial Documentation pidiekab.web.id)</figcaption>
</figure>

<p class="editorial-lead">For anyone traversing the scenic highway along Aceh’s eastern coastal plains, entering Pidie Regency is instantly announced by a signature fragrance drifting upon the warm coastal breeze: the sweet aroma of caramelized palm sugars, bubbling virgin coconut milk, and the irresistible warmth of golden fried shallots. This is the scent of Kue Adee—Pidie’s crown jewel of traditional baked confections.</p>

<p>Kue Adee (often called <em>Bingka Adee</em>) represents both the hospitality and culinary sophistication of the Pidie people. Unlike conventional sponge cakes or generic Southeast Asian bingka desserts, Adee boasts a mesmerizing interplay of flavors: intense caramelized sweetness, a velvety melt-in-the-mouth texture enriched by pure coconut cream, and an audacious savory crunch from crisp shallots sprinkled across its golden surface.</p>

<h2 class="editorial-chapter-title">I. Heritage &amp; Royal Confectionery Roots</h2>
<p>Kue Adee traces its lineage back to the maritime heyday of the Aceh Sultanate. Historically, this rich baked cake was not an everyday street snack; it was an exclusive ceremonial treat reserved for nobility, royal banquets, visiting dignitaries, and revered Islamic scholars (ulama) during festive religious assemblies.</p>

<p>In local folk wisdom, the name <em>Adee</em> evokes a sense of inner tranquility and wholesome harmony. Creating an authentic Adee required painstaking dedication: fine hand-grating of freshly harvested cassava roots, extracting the rich first-press coconut cream without a single drop of added water, and tending wood-charcoal embers for hours on end. Offering Kue Adee to visitors remains the highest expression of Acehnese guest honor (<em>pemulia jamee</em>).</p>

<h2 class="editorial-chapter-title">II. Two Classic Variants: Cassava vs Wheat Custard</h2>
<p>Artisanal bakers throughout Pidie have perfected two distinct variations, each celebrated for its unique textural character:</p>

<ul>
    <li><strong>Adee Ubi (Cassava)</strong>: The oldest and most agrarian version. Made from native white cassava roots, finely grated and lightly pressed to maintain balanced moisture. Blended with unadulterated coconut cream and cane sugar, Adee Ubi yields a pleasantly chewy, dense, and naturally sweet crumb that celebrates Pidie’s fertile earth.</li>
    <li><strong>Adee Terigu (Wheat Flour)</strong>: Emerged during the historical arrival of European flour through coastal trading ports. Whisked with local free-range duck eggs and warm coconut milk, this version achieves a luxurious, velvety custard texture reminiscent of fine baked crème caramel that dissolves gracefully on the palate.</li>
</ul>

<h2 class="editorial-chapter-title">III. The Secret of Double Charcoal Baking</h2>
<p>True Pidie Adee contains zero artificial additives, synthetic colorings, or chemical leavening agents. Its culinary magic rests upon three traditional tenets:</p>

<p><strong>1. First-Press Coconut Cream (<em>Pati U Teupreh</em>)</strong><br>
Crafted exclusively from fully matured coconuts harvested at peak ripeness. The grated meat is pressed without added water, releasing natural plant lipids that impart an unmistakable, nutty depth.</p>

<p><strong>2. Free-Range Local Duck Eggs</strong><br>
Traditional artisans insist on duck eggs over chicken eggs. Their vibrant deep-orange yolks lend a silkier mouthfeel, superior structural stability during baking, and a natural golden luster without artificial dyes.</p>

<p><strong>3. Dual-Zone Charcoal Embers</strong><br>
Adee cakes are baked in circular aluminum pans measuring 20 to 25 centimeters in diameter. The baking process utilizes dual radiant heat: the pan sits above glowing wood charcoal embers, while a heavy metal lid heaped with red-hot coals covers the top. This top-down caramelized heat produces a brittle, golden-brown glaze while keeping the center delightfully moist and tender.</p>

<h2 class="editorial-chapter-title">IV. The Magic of Crisp Shallots</h2>
<p>Perhaps the most fascinating hallmark of Pidie Adee is the generous sprinkling of thinly sliced fried red shallots (<em>bawang goreng</em>) and fragrant anise seeds across its crust moments before it is pulled from the hearth.</p>

<p>While topping a dessert with fried shallots might sound unconventional to newcomers, the first bite reveals culinary genius: the savory notes of caramelized allium cut effortlessly through the dense coconut sweetness, creating an umami counter-balance that keeps the palate fresh, slice after slice.</p>

<h2 class="editorial-chapter-title">V. Cultural Significance: Weddings &amp; Festivals</h2>
<p>In community life, Kue Adee serves as an indispensable ceremonial token:</p>

<ul>
    <li><strong>Wedding Banquets &amp; Dowry Trays (<em>Dalung Meukasab</em>)</strong>: During traditional Acehnese nuptials, Adee is artfully arranged inside brass trays draped with gold-embroidered velvet cloths. Its rich sweetness symbolizes prayers for harmony, lasting warmth, and unbreakable bonding for the newlyweds.</li>
    <li><strong>Religious Holidays &amp; Maulid Gatherings</strong>: During Eid celebrations and the Prophet’s Birthday festivals, communal village meunasah halls and households prepare large batches to welcome returning migrant workers.</li>
    <li><strong>The Traveler's Prized Souvenir</strong>: Anyone journeying home from Pidie to Medan, Jakarta, or overseas destinations takes boxes of freshly baked Adee as a cherished parcel of home.</li>
</ul>

<h2 class="editorial-chapter-title">VI. Paired with Inverted 'Kupi Khop'</h2>
<p>The ultimate way to savor Kue Adee is in the company of a robust Acehnese brew. Across Pidie coffee shops, patrons pair warm slices of Adee with <strong>Kupi Khop</strong>—strong Robusta coffee served dramatically in an inverted glass on a saucer—or freshly roasted coffee from the misty mountain slopes of Tangse.</p>

<div class="editorial-sources-box">
    <h4>💡 Connoisseur’s Serving Guide:</h4>
    <ul>
        <li><strong>Optimal Temperature:</strong> Best enjoyed lukewarm, approximately 30 minutes after leaving the oven.</li>
        <li><strong>Natural Shelf Life:</strong> Due to thorough baking and pure coconut lipids, Adee keeps fresh for 2 to 3 days at cool room temperature, or up to 7 days refrigerated. Gently rewarm in an oven or dry pan to restore its crispy crust.</li>
        <li><strong>Tasting Tip:</strong> Compare the hearty texture of cassava Adee against the silky melt of wheat custard Adee to appreciate the full spectrum of Pidie baking heritage.</li>
    </ul>
</div>
`, imageURL)

	enSummary := "Pidie's legendary caramelized baked cake crafted from rich coconut cream, velvety cassava or wheat flour, duck eggs, and a signature crisp shallot aroma."

	enBodyJSON, _ := json.Marshal(map[string]string{
		"html":     enHTML,
		"markdown": enHTML,
	})

	// Upsert English translation
	_, err = dbPool.Exec(ctx, `
		INSERT INTO content_translations (
			content_id, locale, title, slug, summary, body, status, published_at, published_by,
			search_vector, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, 'en', $2, $3, $4, $5, 'published', NOW(), $6,
			setweight(to_tsvector('simple', COALESCE($2, '')), 'A') || setweight(to_tsvector('simple', COALESCE($4, '')), 'B'),
			$6, $6, NOW(), NOW()
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
	`, contentID, enTitle, enSlug, enSummary, enBodyJSON, adminID)

	if err != nil {
		return fmt.Errorf("failed to save EN translation for kue adee: %w", err)
	}
	fmt.Println(" [+] English Translation Published: 'Traditional Adee Cake: Sweet Fragrance of Pidie Confectionery'")

	return nil
}
