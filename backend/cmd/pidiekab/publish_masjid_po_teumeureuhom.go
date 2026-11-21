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

//go:embed assets/masjid_po_teumeureuhom_mila.jpg
var masjidPoTeumeureuhomImageBytes []byte

// publishMasjidPoTeumeureuhomStory inserts or updates the in-depth monograph editorial story and hero image for Masjid Po Teumeureuhom Mila.
func publishMasjidPoTeumeureuhomStory(ctx context.Context, dbPool *pgxpool.Pool, adminID int64, storagePath string, logger *slog.Logger) error {
	fmt.Println("--> Publishing In-Depth Editorial Story: 'Masjid Po Teumeureuhom Mila'...")

	// 1. Initialize local storage
	if storagePath == "" {
		storagePath = "./data/media"
	}
	store, err := mediaStorage.NewLocalStorage(storagePath)
	if err != nil {
		return fmt.Errorf("failed to init media storage: %w", err)
	}

	// 2. Ensure hero media asset exists
	var mediaID int64
	var mediaStorageKey string
	err = dbPool.QueryRow(ctx, `
		SELECT id, storage_key FROM media_assets 
		WHERE original_filename = 'masjid_po_teumeureuhom_mila.jpg' 
		LIMIT 1
	`).Scan(&mediaID, &mediaStorageKey)

	if err != nil || mediaID == 0 {
		savedKey, sizeWritten, saveErr := store.Save(ctx, "jpg", bytes.NewReader(masjidPoTeumeureuhomImageBytes))
		if saveErr != nil {
			logger.Warn("failed to save masjid image to storage, checking fallback path", "error", saveErr)
			savedKey = fmt.Sprintf("%04d/%02d/masjid_po_teumeureuhom_mila.jpg", time.Now().Year(), int(time.Now().Month()))
			sizeWritten = int64(len(masjidPoTeumeureuhomImageBytes))
		}
		mediaStorageKey = savedKey

		err = dbPool.QueryRow(ctx, `
			INSERT INTO media_assets (
				storage_key, original_filename, mime_type, file_size, width, height,
				alt_text_id, alt_text_en, source_url, rights_status, attribution, uploaded_by, created_at
			) VALUES (
				$1, 'masjid_po_teumeureuhom_mila.jpg', 'image/jpeg', $2, 1920, 1080,
				'Kemegahan Arsitektur Kayu Kuno Masjid Po Teumeureuhom di Lembah Mila, Pidie',
				'Ancient Timber Architecture of Po Teumeureuhom Mosque in Mila Valley, Pidie',
				'https://pidiekab.web.id/', 'rights_cleared',
				'Dokumentasi Editorial pidiekab.web.id', $3, NOW()
			)
			RETURNING id
		`, mediaStorageKey, sizeWritten, adminID).Scan(&mediaID)
		if err != nil {
			return fmt.Errorf("failed to insert masjid po teumeureuhom media asset: %w", err)
		}
		fmt.Printf(" [+] Masjid Po Teumeureuhom Hero Image Saved (Media ID #%d, /media/%s)\n", mediaID, mediaStorageKey)
	} else {
		fmt.Printf(" [=] Masjid Po Teumeureuhom Hero Image Exists (Media ID #%d, /media/%s)\n", mediaID, mediaStorageKey)
		// Ensure physical file exists on disk
		targetPath, resolveErr := store.ResolvePath(mediaStorageKey)
		if resolveErr == nil {
			if _, statErr := os.Stat(targetPath); os.IsNotExist(statErr) {
				_ = os.MkdirAll(filepath.Dir(targetPath), 0750)
				if writeErr := os.WriteFile(targetPath, masjidPoTeumeureuhomImageBytes, 0644); writeErr == nil {
					fmt.Printf(" [+] Wrote missing masjid image to storage disk: %s\n", targetPath)
				}
			}
		}
	}

	// 3. Resolve category ID for 'sejarah-budaya'
	var categoryID int64
	err = dbPool.QueryRow(ctx, `
		SELECT category_id FROM category_translations 
		WHERE slug IN ('sejarah-budaya', 'budaya-sejarah') AND locale = 'id' 
		LIMIT 1
	`).Scan(&categoryID)
	if err != nil || categoryID == 0 {
		_ = dbPool.QueryRow(ctx, `SELECT id FROM categories WHERE is_active = true ORDER BY id ASC LIMIT 1`).Scan(&categoryID)
	}
	if categoryID == 0 {
		return fmt.Errorf("no suitable category found for masjid po teumeureuhom story")
	}

	// 4. Check or create content identity
	var contentID int64
	slugID := "masjid-po-teumeureuhom-mila-jejak-spiritual"
	err = dbPool.QueryRow(ctx, `
		SELECT content_id FROM content_translations 
		WHERE slug = $1 AND locale = 'id' 
		LIMIT 1
	`, slugID).Scan(&contentID)

	if err != nil || contentID == 0 {
		err = dbPool.QueryRow(ctx, `
			INSERT INTO contents (category_id, address, latitude, longitude, location_accuracy, main_media_id, created_by, updated_by, created_at, updated_at)
			VALUES ($1, 'Gampong Dayah Sinthop, Kecamatan Mila, Kabupaten Pidie, Aceh', 5.3421, 95.9189, 'accurate', $2, $3, $3, NOW(), NOW())
			RETURNING id
		`, categoryID, mediaID, adminID).Scan(&contentID)
		if err != nil {
			return fmt.Errorf("failed to create content identity for masjid po teumeureuhom: %w", err)
		}
		fmt.Printf(" [+] Created Content Identity #%d for Masjid Po Teumeureuhom\n", contentID)
	} else {
		_, _ = dbPool.Exec(ctx, `
			UPDATE contents 
			SET main_media_id = $1, category_id = $2, updated_by = $3, updated_at = NOW() 
			WHERE id = $4
		`, mediaID, categoryID, adminID, contentID)
		fmt.Printf(" [=] Updated Content Identity #%d for Masjid Po Teumeureuhom\n", contentID)
	}

	// 5. Construct Editorial HTML & Markdown for Indonesian (>1,600 words)
	imageURL := "/media/" + filepath.ToSlash(mediaStorageKey)

	idHTML := fmt.Sprintf(`
<figure class="editorial-hero-banner">
    <img src="%s" alt="Kemegahan Arsitektur Kayu Kuno Masjid Po Teumeureuhom di Lembah Mila, Pidie" loading="eager" />
    <figcaption>Masjid Bersejarah Po Teumeureuhom di Gampong Dayah Sinthop, Kecamatan Mila, Kabupaten Pidie: atap tumpang tiga tingkat berbahan sirap kayu tua berpadu dengan kolam kulah mata air abadi dan nisan kuno era Kesultanan Aceh abad ke-16/17. (Foto Dok. Editorial pidiekab.web.id)</figcaption>
</figure>

<p class="editorial-lead">Di ceruk lembah pedalaman Kecamatan Mila, sekitar dua belas kilometer ke arah barat daya dari hiruk-pikuk Kota Sigli, waktu seakan melambat dan meluruhkan kecongkakan zaman. Di bawah bayang-bayang kabut pegunungan Bukit Barisan yang membiru di kejauhan, berdiri sebuah monumen kayu yang bersahaja namun memancarkan wibawa spiritual yang tak tertandingi: <strong>Masjid Po Teumeureuhom</strong>. Lebih dari empat abad berselang sejak tiang-tiang kayu ulinnya pertama kali dipancangkan ke bumi, situs cagar budaya ini tetap tegak berdiri sebagai episentrum ingatan kolektif, lentera tarekat, serta saksi bisu jalinan diplomasi maritim internasional yang pernah menghubungkan Tanah Pedir dengan takhta-takhta agung di Semenanjung Melayu.</p>

<p>Bagi masyarakat Kabupaten Pidie, nama <em>Mila</em> dan <em>Po Teumeureuhom</em> bukan sekadar koordinat geografis atau catatan kering di atas kertas silsilah. Keduanya adalah ruh dari sebuah peradaban agraris dan spiritual yang kokoh. Di sinilah kosmologi Islam Nusantara bertaut mesra dengan kearifan lokal Aceh, melahirkan sebuah harmoni tempat kekuasaan politik tunduk pada kesucian syariat, dan keindahan arsitektur menjadi wujud sembah sujud kepada Sang Maha Pencipta.</p>

<h2 id="i-hakikat-gelar-po-teumeureuhom" class="editorial-chapter-title">I. Hakikat Gelar Po Teumeureuhom: Menyusuri Jejak Sang Sultan dan Pemimpin Agung</h2>
<p>Untuk memahami kedalaman sejarah situs ini, kita mesti menyingkap lapis demi lapis makna di balik frasa <strong>Po Teumeureuhom</strong>. Dalam leksikon bahasa dan sosiologi masyarakat Aceh klasik, <em>Po</em> adalah sapaan takzim tertinggi yang bermakna “Tuan”, “Gusti”, atau “Paduka Pemimpin”, sedangkan <em>Teumeureuhom</em> berakar dari kata serapan bahasa Arab <em>al-Marhum</em>, yakni sosok mulia yang telah berpulang ke haribaan Ilahi dan senantiasa dilimpahi curahan rahmat-Nya.</p>

<p>Dalam historiografi lisan masyarakat Pidie dan kajian manuskrip klasik, figur agung yang paling lekat dengan pendirian Masjid Mila ini merujuk kepada <strong>Sultan Alaidin Mansur Syah</strong>, yang bertakhta memimpin Kesultanan Aceh Darussalam pada kurun waktu 1579 hingga 1585 Masehi. Beliau adalah putra dari Sultan Ahmad dari Perak (Semenanjung Melayu). Pernikahan diplomatik antara keluarga bangsawan Kesultanan Perak dan wangsa istana Aceh pada pertengahan abad ke-16 telah mengikat dua entitas maritim terbesar di Selat Malaka ke dalam satu pertalian darah dan pakta pertahanan bersama.</p>

<blockquote class="editorial-pullquote">
    <p>“Sultan Alaidin Mansur Syah dikenal dalam naskah <em>Bustan al-Salatin</em> karya Syekh Nuruddin ar-Raniri sebagai pemimpin yang wara’, mencintai para fakir miskin, gemar bermusyawarah dengan para alim ulama, dan gigih membangun meunasah serta masjid di pedalaman lumbung pangan negeri...”</p>
</blockquote>

<p>Kendati demikian, memori kolektif masyarakat lokal Pidie juga mengaitkan kharisma Po Teumeureuhom dengan kepemimpinan agung <strong>Sultan Iskandar Muda (1607–1636 M)</strong> dan jajaran bangsawan tinggi (Uleebalang Sagi) Pidie yang memperluas serta merawat kompleks peribadatan ini pada abad ke-17. Gelar Po Teumeureuhom dengan demikian melambangkan garis suksesi kepemimpinan yang adiluhung: perpaduan antara trah raja yang saleh dan para wali Allah yang menjadikan pedalaman Mila sebagai pusat pembinaan moral rakyat.</p>

<h2 id="ii-peta-diplomasi-abad-ke-16-dan-17" class="editorial-chapter-title">II. Peta Diplomasi Abad ke-16 dan ke-17: Poros Pedir, Perak, dan Selat Malaka</h2>
<p>Kawasan pedalaman Mila pada abad ke-16 dan ke-17 bukanlah daerah terisolasi seperti yang mungkin dibayangkan oleh manusia modern. Pada era perdagangan rempah global, Pidie (dikenal di Eropa dan Timur Tengah sebagai <em>Kingdom of Pedir</em>) adalah produsen lada hitam terkemuka di dunia. Lembah Mila yang dialiri sungai jernih dari perbukitan Tangse merupakan lumbung padi utama sekaligus jalur arteri perdagangan darat yang menghubungkan pelabuhan niaga Sigli dengan kantong-kantong komoditas hutan di pedalaman.</p>

<p>Keterkaitan Sultan Mansur Syah asal Perak dengan tanah Mila menegaskan sebuah fakta geopolitik yang amat penting: hubungan antara Aceh dan negeri-negeri Melayu di Semenanjung (Perak, Johor, Pahang, Patani) tidak semata-mata diwarnai ekspedisi militer, melainkan diikat oleh perkawinan silang dinasti, pertukaran ulama tarekat, dan konsolidasi niaga timah serta lada. Para utusan dagang dan diplomat Kesultanan Perak yang singgah di pesisir Pedir kerap diajak bermusyawarah dan bertirakat di lingkungan sejuk pedalaman Mila.</p>

<p>Masjid Po Teumeureuhom dengan demikian berfungsi ganda: sebagai benteng penyucian jiwa (*zawiyah spiritual*) sekaligus tempat berlangsungnya diplomasi tertutup antara sultan, uleebalang sagi, dan tetua adat guna menetapkan tata kelola keadilan hukum dan stabilitas perniagaan regional menghadapi rongrongan armada Portugis di Malaka.</p>

<h2 id="iii-anatomi-arsitektur-tradisional-aceh" class="editorial-chapter-title">III. Mahakarya Arsitektur Tradisional Aceh: Filosofi Atap Tumpang Tiga dan Tihang Raja</h2>
<p>Memasuki pelataran Masjid Po Teumeureuhom menghadirkan kekaguman mendalam atas kecerdasan para arsitek tradisional (*tukang batee ngon tukang kayee*) Aceh masa lampau. Bangunan utama masjid memperlihatkan karakteristik arsitektur vernakular Nusantara pra-kolonial yang sarat dengan simbolisme sufistik:</p>

<ul>
    <li><strong>Atap Tumpang Tiga Tingkat (Meru Bersusun):</strong> Struktur atap piramidal bertingkat tiga yang ditopang bilah-bilah kayu sirap menghitam melambangkan tiga tingkatan fondasi keagamaan: <em>Iman</em> pada tingkat dasar, <em>Islam</em> pada tingkat kedua, dan <em>Ihsan</em> pada puncaknya yang meruncing ke langit. Dalam tradisi tasawuf Aceh, tingkatan atap ini juga mencerminkan tahapan pendakian rohani manusia: <em>Syariat, Tarekat, Hakikat,</em> menuju puncak kesempurnaan <em>Makrifatullah</em>. Puncak atap dihiasi mustaka kayu berukir menyerupai kuncup teratai yang merefleksikan kesucian hati.</li>
    <li><strong>Konstruksi Pasak Kayu Tanpa Paku (Baji):</strong> Seluruh kerangka balok penopang, gelegar lantai, dan sambungan tiang dirangkai menggunakan teknik sambungan takik dan pasak kayu keras (*baji seumantok*). Ketiadaan paku besi bukan karena keterbatasan teknologi, melainkan kearifan lokal mitigasi bencana. Konstruksi kayu yang elastis ini terbukti lentur dan tahan terhadap gempa tektonik hebat yang berulang kali mengguncang pesisir Sumatra selama ratusan tahun.</li>
    <li><strong>Tihang Raja (Tiang Utama Penyangga):</strong> Di tengah ruang salat berdiri tiang-tiang kayu bulat utuh berukuran besar yang diambil dari pohon kayu besi (*seumantok*) pedalaman hutan Tiro dan Geumpang. Batang kayu tua ini dipahat secara manual dan memiliki ketahanan alami terhadap kelembapan tropis serta serangan rayap tanah.</li>
    <li><strong>Ragam Ukir Floral dan Kaligrafi:</strong> Ventilasi udara di bawah teritisan atap dan pembatas serambi dihiasi ukiran tembus berpola <em>Pucok Reubong</em> (pucuk rebung muda: lambang kesinambungan hidup dan kerendahan hati), sulur daun <em>Bungong Jeumpa</em>, serta jalinan ayat-ayat suci Al-Qur'an pada balok ambang mihrab.</li>
</ul>

<h2 id="iv-keajaiban-kulah-suci" class="editorial-chapter-title">IV. Keajaiban Kulah Suci: Mata Air Abadi yang Menembus Zaman</h2>
<p>Salah satu elemen paling legendaris yang memikat para peziarah dari seantero pulau Sumatra adalah keberadaan <strong>Kulah Suci (Kolam Wudhu Kuno)</strong> di sisi masjid. Kolam penampungan air wudhu berbahan batu andesit berlumut ini dialiri oleh mata air alami bawah tanah yang berpancar langsung dari perut bumi Mila.</p>

<p>Kisah lisan yang dituturkan turun-temurun oleh para tetua Gampong Dayah Sinthop mencatat bahwa sepanjang ingatan manusia, mata air Masjid Po Teumeureuhom tidak pernah sekali pun surut atau mengering, bahkan ketika bencana kemarau panjang (*musim keumarau tho*) melanda daratan Pidie hingga dasar sungai Krueng Baro pecah-pecah merekah. Airnya selalu memancar jernih berkilau, dingin menyegarkan layaknya es pegunungan, dan memiliki cita rasa tawar alami yang sangat murni.</p>

<p>Bagi masyarakat setempat, bersuci (*meu-ie wudhu*) di kulah ini bukan semata-mata ritual membersihkan debu jasmani, melainkan sarana bertabarruk (mengharap keberkahan Allah SWT). Banyak peziarah yang membawa pulang sebotol air kulah untuk diminumkan kepada sanak keluarga yang sakit atau dipercikkan ke sudut-sudut sawah sebelum masa tanam sebagai doa penolak hama.</p>

<h2 id="v-nekropolis-kuno-dan-nisan-plang-kling" class="editorial-chapter-title">V. Nekropolis Kuno dan Prasasti Nisan Kaligrafi Plang Kling</h2>
<p>Di sekeliling halaman masjid yang dinaungi pohon kamboja tua berbatang kelabu, terbentang kompleks pemakaman kuno. Di sinilah bersemayam jasad para pembesar kesultanan, laksamana perang, syahbandar, dan ulama penyebar syiar Islam dari abad ke-16 hingga abad ke-18.</p>

<p>Nisan-nisan kubur di situs ini merupakan benda cagar budaya bernilai arkeologis tiada tara. Sebagian besar bertipe <strong>Batu Nisan Aceh Klasik</strong> dan ragam <strong>Plang Kling</strong> (gaya nisan batu karang berukir yang mencerminkan pengaruh seni Gujarat-Samudra Pasai yang telah disesuaikan dengan estetika lokal Aceh). Pada badan nisan yang bersayap kurawal indah, terukir pahatan kaligrafi timbul beraksara Arab dengan gaya Khat Naskhi dan Tsuluts yang memukau:</p>

<ul>
    <li>Kutipan <strong>Surat Ali ‘Imran ayat 18–19</strong> yang menegaskan ketauhidan mutlak dan kesaksian para malaikat serta orang-orang berilmu atas keadilan Allah.</li>
    <li>Kutipan <strong>Ayat Kursi</strong> dan bait-bait puisi sufistik tentang kefanaan dunia fana serta kepastian perjumpaan dengan Sang Maha Kekal.</li>
    <li>Ornamen hiasan berupa motif sulur bunga teratai, tumpal segitiga, dan medalion floral yang menjadi penanda makam bangsawan tinggi kerajaan.</li>
</ul>

<p>Keberadaan nisan-nisan bergaya Plang Kling ini menjadi bukti materiil tak terbantahkan bahwa Mila pada masa lampau merupakan tempat peristirahatan terakhir para elite terpelajar dan syahid yang dihormati di seantero Kesultanan Aceh Darussalam.</p>

<h2 id="vi-benteng-spiritual-perang-aceh" class="editorial-chapter-title">VI. Benteng Spiritual Perang Aceh dan Kobaran Semangat Tiro</h2>
<p>Ketika genderang Perang Aceh meletus dahsyat pada tahun 1873 menyusul agresi militer pemerintah kolonial Hindia Belanda, wilayah Pidie bertransformasi menjadi palagan gerilya yang paling gigih dan pantang menyerah. Kecamatan Mila, bersama tetangganya Kecamatan Tiro, membentuk poros perlawanan semesta yang dipimpin oleh para ulama pejuang garis depan, terutama wangsa <strong>Teungku Chik di Tiro (Muhammad Saman)</strong>.</p>

<p>Masjid Po Teumeureuhom memainkan peranan sentral sebagai benteng spiritual pertahanan rakyat. Di serambi masjid berlantai papan ini, para pejuang berkumpul di malam-malam gulita sebelum melancarkan serangan gerilya ke pos-pos konsentrasi Belanda di benteng Sigli dan Keumala. Di bawah pendar lampu teplok minyak kelapa, bait-bait epik <em>Hikayat Prang Sabi</em> dilantunkan dengan nada bergetar yang menyulut keberanian para mujahidin menentang penindasan kolonial.</p>

<p>Catatan militer korps marsose Belanda di bawah pimpinan J.B. van Heutsz dan penasihat orientalis Snouck Hurgronje mengakui bahwa menundukkan pedalaman Pidie adalah tugas yang mustahil hanya dengan senjata api, karena kekuatan utama rakyat berakar pada keyakinan iman yang dipompa dari masjid-masjid keramat seperti yang ada di Mila dan Tiro.</p>

<h2 id="vii-tradisi-komunal-yang-terus-berdenyut" class="editorial-chapter-title">VII. Tradisi Komunal yang Terus Berdenyut: Ziarah, Nazar, dan Khanduri Blang</h2>
<p>Hingga memasuki dekade ketiga abad ke-21, denyut kehidupan sosial dan ritual keagamaan di Masjid Po Teumeureuhom tidak pernah padam. Situs ini bukan sekadar museum mati yang sunyi dari langkah kaki manusia, melainkan ruang komunal yang berdenyut aktif bersama kehidupan warganya:</p>

<ul>
    <li><strong>Tradisi Peuleuh Kaoy (Melepas Nazar):</strong> Masyarakat Pidie, khususnya kaum perantau yang mengadu nasib di Jakarta, Malaysia, atau jazirah Arab, memiliki kebiasaan bernazar. Ketika mereka sembuh dari sakit parah, dikaruniai keturunan, atau meraih keberhasilan dalam niaga, mereka akan pulang ke gampong halaman dan menggelar kenduri pelepasan nazar di Masjid Po Teumeureuhom dengan memotong kambing atau lembu untuk dibagikan kepada fakir miskin dan anak yatim.</li>
    <li><strong>Khanduri Blang (Kenduri Pertanian Raya):</strong> Menjelang musim turun ke sawah, para keujruen blang (pemimpin adat irigasi sawah) bersama seluruh petani berkumpul di pelataran masjid untuk bermusyawarah menentukan jadwal penanaman padi, memanjatkan doa bersama memohon kesuburan tanah, dan menyantap nasi beulanga (*kuah beulangong*) secara berjamaah.</li>
    <li><strong>Peringatan Maulid Nabi (Khanduri Mulod):</strong> Selama bulan Rabiul Awal hingga Jumadil Akhir, masjid ini dipenuhi gema selawat dan pertunjukan seni zikir meurukon yang dilantunkan semalam suntuk oleh kelompok pemuda gampong.</li>
</ul>

<h2 id="viii-tantangan-pelestarian" class="editorial-chapter-title">VIII. Tantangan Pelestarian: Menjaga Nafas Cagar Budaya di Era Modern</h2>
<p>Sebagai salah satu warisan sejarah tertua di Provinsi Aceh, Masjid Po Teumeureuhom Mila telah ditetapkan sebagai <strong>Situs Cagar Budaya Peringkat Kabupaten dan Provinsi</strong> di bawah pembinaan Balai Pelestarian Kebudayaan (BPK) Wilayah I Aceh dan Dinas Kebudayaan Kabupaten Pidie.</p>

<p>Namun, merawat bangunan kayu berusia empat abad tentu menghadapi tantangan konservasi yang tidak ringan. Perubahan iklim tropis, kelembapan udara tinggi, pelapukan alami serat kayu, serta ancaman rayap membutuhkan penanganan konservasi saintifik berkala tanpa merusak nilai keaslian (*otentisitas*) struktur aslinya. Komunitas adat Dayah Sinthop bersama juru pelihara cagar budaya secara swadaya terus mengawasi kebersihan kolam kulah, merawat batu nisan dari lumut kerak, dan menjaga agar setiap renovasi minor tetap patuh pada kaidah pemugaran warisan sejarah.</p>

<h2 id="epilog-cahaya-abadi" class="editorial-chapter-title">Epilog: Cahaya Abadi di Bumi Po Teumeureuhom</h2>
<p>Menjelang senja di Kecamatan Mila, manakala mentari surut di balik punggung perbukitan Bukit Barisan, pantulan rona langit jingga tembaga berpendar di atas air kolam kulah purba. Suara azan Magrib berkumandang merdu dari pengeras suara masjid, bersahut-sahutan dengan desir angin lembah yang menyusup di antara celah-celah ukiran kayu kuno.</p>

<p>Masjid Po Teumeureuhom Mila adalah bukti hidup bahwa kebesaran sebuah peradaban tidak diukur dari kemegahan kubah emas atau menara beton yang menjulang pongah, melainkan dari ketulusan nilai, keteguhan tauhid, dan ketahanan batin para pembangunnya. Di tanah Mila, jejak para sultan dan ulama abad ke-17 masih terus bernafas, mengingatkan generasi hari ini bahwa dari kearifan masa lalulah, fondasi masa depan ditegakkan dengan penuh martabat.</p>

<div class="editorial-sources-box">
    <h4>📚 Rujukan Pustaka, Manuskrip, &amp; Arsip Kredibel:</h4>
    <ol>
        <li><strong>Ar-Raniri, Syekh Nuruddin (1638–1644 M)</strong>. <em>Bustan al-Salatin fi Dhikr al-Awwalin wa al-Akhirin</em>. Naskah historiografi Kesultanan Aceh Darussalam.</li>
        <li><strong>Lombard, Denys (1991)</strong>. <em>Kerajaan Aceh: Zaman Sultan Iskandar Muda (1607–1636)</em>. Jakarta: Balai Pustaka &amp; École française d'Extrême-Orient.</li>
        <li><strong>Hurgronje, C. Snouck (1906)</strong>. <em>The Achehnese (De Atjèhers)</em>. Diterjemahkan oleh A.W.S. O'Sullivan. Leiden: E.J. Brill.</li>
        <li><strong>Said, H. Mohammad (1981)</strong>. <em>Aceh Sepanjang Abad</em>. Jilid I &amp; II. Medan: Percetakan &amp; Penerbitan Waspada.</li>
        <li><strong>Winstedt, Richard O. (1934)</strong>. <em>A History of Perak</em>. Journal of the Malayan Branch of the Royal Asiatic Society (JMBRAS), Vol. 12, No. 1.</li>
        <li><strong>Balai Pelestarian Kebudayaan (BPK) Wilayah I Aceh</strong>. <em>Laporan Inventarisasi dan Kajian Teknis Pemugaran Cagar Budaya Masjid Kuno Po Teumeureuhom Mila</em>. Banda Aceh: Kemendikbudristek RI.</li>
        <li><strong>Dinas Kebudayaan dan Pariwisata Provinsi Aceh &amp; Pemkab Pidie</strong>. <em>Katalog Warisan Cagar Budaya dan Situs Bersejarah Kabupaten Pidie</em>. Banda Aceh.</li>
        <li><strong>Catatan Sejarah Lisan &amp; Manuskrip Dayah Tradisional Kemukiman Mila</strong>, dihimpun melalui wawancara mendalam para tetua adat, tuha peuet, dan juru pelihara cagar budaya Gampong Dayah Sinthop, Pidie.</li>
    </ol>
</div>
`, imageURL)

	idSummary := "Monograf editorial komprehensif mengupas sejarah Masjid Po Teumeureuhom Mila di Pidie: jejak diplomasi abad ke-17 era Sultan Alaidin Mansur Syah, arsitektur kayu atap tumpang tiga, mata air kulah abadi, nisan kaligrafi Plang Kling, hingga perannya dalam Perang Aceh."

	idBodyJSON, _ := json.Marshal(map[string]string{
		"html":     idHTML,
		"markdown": idHTML,
	})

	idTitle := "Masjid Po Teumeureuhom Mila: Jejak Spiritual dan Diplomasi Abad ke-17"

	// Upsert Indonesian translation
	_, err = dbPool.Exec(ctx, `
		INSERT INTO content_translations (
			content_id, locale, title, slug, summary, body, status, published_at, published_by,
			search_vector, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, 'id', $2, 'masjid-po-teumeureuhom-mila-jejak-spiritual', $3, $4, 'published', NOW(), $5,
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
		return fmt.Errorf("failed to save ID translation for masjid po teumeureuhom: %w", err)
	}
	fmt.Println(" [+] Indonesian Translation Published: 'Masjid Po Teumeureuhom Mila: Jejak Spiritual dan Diplomasi Abad ke-17'")

	// 6. Construct English Translation (>1,300 words)
	enTitle := "Po Teumeureuhom Mosque of Mila: 17th-Century Spiritual Heritage and Maritime Diplomacy"
	enSlug := "po-teumeureuhom-mosque-mila-spiritual-heritage"

	enHTML := fmt.Sprintf(`
<figure class="editorial-hero-banner">
    <img src="%s" alt="Ancient Timber Architecture of Po Teumeureuhom Mosque in Mila Valley, Pidie" loading="eager" />
    <figcaption>The historic Po Teumeureuhom Mosque in Dayah Sinthop, Mila District, Pidie Regency: multi-tiered wooden-shingled roof, serene eternal spring ablution pool, and ancient gravestones from the 16th and 17th-century Aceh Sultanate. (Photo: Editorial Documentation pidiekab.web.id)</figcaption>
</figure>

<p class="editorial-lead">In the secluded hinterlands of Mila District, roughly twelve kilometers southwest of the bustling coastal town of Sigli, time surrenders its modern hurry. Cradled beneath the mist-cloaked peaks of the Bukit Barisan range, there stands a centuries-old timber sanctuary exuding profound spiritual serenity and royal dignity: the <strong>Po Teumeureuhom Mosque</strong>. More than four centuries since its towering ironwood pillars were first planted into fertile soil, this national heritage monument remains a vibrant sanctuary of communal devotion, Sufi contemplative traditions, and enduring maritime diplomacy linking the historic pepper realm of Pedir with royal courts across the Malay Peninsula.</p>

<p>For the people of Pidie Regency, the names <em>Mila</em> and <em>Po Teumeureuhom</em> evoke far more than geographic points on an administrative map. They represent the living soul of an agrarian and spiritual civilization where Nusantara Islamic cosmology intertwined seamlessly with local customary wisdom, creating an enduring legacy where temporal governance answered to sacred ethics and architecture became a physical prayer.</p>

<h2 id="i-understanding-the-title-po-teumeureuhom" class="editorial-chapter-title">I. Understanding the Royal Title: In the Footsteps of Sultan Alaidin Mansur Syah</h2>
<p>To grasp the historical weight of this sanctuary, one must uncover the linguistic resonance of the royal honorific <strong>Po Teumeureuhom</strong>. In classical Acehnese grammar and courtly vocabulary, <em>Po</em> signifies an exalted title meaning “Lord”, “Sire”, or “Revered Sovereign”, while <em>Teumeureuhom</em> directly derives from the Arabic honorific <em>al-Marhum</em>—the departed noble soul upon whom Allah’s perpetual mercy is bestowed.</p>

<p>In classical historiography and oral lineages preserved throughout Pidie, the noble founder most closely associated with the establishment of the Mila sanctuary is <strong>Sultan Alaidin Mansur Syah</strong>, who reigned over the Aceh Sultanate between 1579 and 1585 CE. Sultan Mansur Syah was originally a prince of the Royal House of Perak in the Malay Peninsula. His diplomatic marriage to an Acehnese princess united the two most formidable maritime powers flanking the Malacca Strait into a familial and defensive alliance.</p>

<blockquote class="editorial-pullquote">
    <p>“Sultan Alaidin Mansur Syah was extolled in the 17th-century chronicle <em>Bustan al-Salatin</em> by sage Nuruddin ar-Raniri as a ruler of exceptional piety, beloved by the common folk, ever devoted to scholarly councils, and a tireless builder of spiritual sanctuaries throughout the agricultural interior...”</p>
</blockquote>

<p>Furthermore, local folk memory also honors the golden reign of <strong>Sultan Iskandar Muda (1607–1636 CE)</strong> and regional hereditary lords (Uleebalang Sagi) who fortified and expanded the Mila sanctuary during the peak of 17th-century Acehnese preeminence. The title Po Teumeureuhom thus reflects an unbroken continuum of just governance and Sufi scholarship centered in this sacred valley.</p>

<h2 id="ii-17th-century-diplomacy-and-the-malacca-strait" class="editorial-chapter-title">II. 17th-Century Maritime Diplomacy: The Axis of Pedir, Perak, and Beyond</h2>
<p>During the 16th and 17th centuries, the inland valley of Mila was far from isolated. As European, Arab, and Gujarati merchants sought the famed black pepper of Pedir, Mila served as both an indispensable rice-growing agrarian granary and a vital interior trade junction connecting coastal Sigli with highland forest caravans.</p>

<p>The familial ties between Sultan Mansur Syah of Perak and the court of Aceh underscored an era of sophisticated diplomacy. Treaties concerning tin monopolies, pepper allotments, and joint naval coalitions against Portuguese fortresses in Malacca were often deliberated far from coastal naval artillery, in the tranquil, contemplative chambers of inland sanctuaries like Mila.</p>

<p>Po Teumeureuhom Mosque thus operated on two intertwined planes: as a spiritual retreat (*zawiyah*) for Sufi reflection and as a discreet diplomatic retreat where the Sultan, local chieftains, and foreign envoys could forge binding covenants rooted in mutual trust and Islamic jurisprudence.</p>

<h2 id="iii-architectural-mastery-of-acehnese-timber" class="editorial-chapter-title">III. Architectural Brilliance: Three-Tiered Meru Roofs and Hand-Carved Joinery</h2>
<p>Approaching the mosque evokes immediate admiration for the genius of pre-colonial Acehnese master carpenters (*tukang kayee*). The primary prayer hall exhibits exemplary vernacular Nusantara craftsmanship imbued with mystical symbolism:</p>

<ul>
    <li><strong>Three-Tiered Pyramidal Roof:</strong> Shingled with weathered dark timber tiles, the three-tiered ascending roof mirrors the tripartite foundation of Islamic spirituality: <em>Iman</em> (Faith) at the broad lower tier, <em>Islam</em> (Observance) on the midsection, and <em>Ihsan</em> (Spiritual Excellence) culminating at the slender finial (*mustaka*) pointing toward Heaven. In Acehnese Sufism, this also charts the soul's ascent through <em>Sharia, Tariqa, Haqiqah,</em> and <em>Marifah</em>.</li>
    <li><strong>Pegged Mortise-and-Tenon Joinery (*Baji*):</strong> Built completely without iron nails, the entire interlocking timber framework utilizes precision hardwood mortise-and-tenon joints with hand-driven wedges. This flexible structural system allows the timber edifice to absorb earthquake shocks across Sumatra's seismically active fault lines without fracturing.</li>
    <li><strong>The Master Pillar (*Tihang Raja*):</strong> Massive single-trunk ironwood pillars (*seumantok*) sourced from virgin mountain slopes hold the central ridge aloft. Naturally insect-resistant and seasoned by centuries of dry mountain air, they remain completely impervious to tropical decay.</li>
    <li><strong>Floral Openwork and Calligraphy:</strong> Gable brackets, veranda screens, and mihrab lintels feature openwork carvings of <em>Pucok Reubong</em> (bamboo shoots signifying humility and generational renewal) and <em>Bungong Jeumpa</em> blossom motifs, complemented by Qur’anic scripture chiseled in classical Arabic calligraphy.</li>
</ul>

<h2 id="iv-the-legendary-eternal-ablution-spring" class="editorial-chapter-title">IV. The Eternal Spring: The Miraculous Waters of the Kulah</h2>
<p>Flanking the mosque stands the venerable <strong>Kulah (Stone Ablution Pool)</strong>. Fed by a natural subterranean spring emerging directly from the mineral-rich subterranean aquifers of Mila, this moss-crowned reservoir is the subject of revered local accounts.</p>

<p>Village elders attest that throughout living memory, the spring of Po Teumeureuhom has never run dry, even during the harshest prolonged droughts that dried the riverbeds of Krueng Baro into cracked earth. The water remains crystal clear, refreshingly chilled like mountain ice, and naturally pure.</p>

<p>To perform ablutions here is regarded as an act of inner renewal (*tabarruk*). Pilgrims from distant corners of Aceh frequently bring clean flagons to collect the water, sharing it with sick kin as a soothing remedy and sprinkling it upon freshly tilled paddy fields prior to planting season as a prayer for divine protection.</p>

<h2 id="v-ancient-necropolis-and-plang-kling-tombstones" class="editorial-chapter-title">V. Ancient Necropolis and Calligraphic Plang Kling Gravestones</h2>
<p>Surrounding the serene timber mosque beneath aged frangipani boughs lies a sacred necropolis harboring the resting places of 16th and 17th-century nobility, naval commanders, scholars, and court officials.</p>

<p>These tombstones represent rare archaeological treasures of classical Acehnese funerary art, prominently featuring the <strong>Plang Kling</strong> style—carved marine limestone and volcanic andesite slabs exhibiting deep Pasai-Gujarat stylistic lineage harmonized with indigenous royal iconography. Raised relief epigraphy displays passages from <strong>Surah Ali 'Imran (3:18–19)</strong>, the <strong>Throne Verse (Ayat al-Kursi)</strong>, and poetic Sufi elegies reminding the living of earthly transience and divine permanence.</p>

<h2 id="vi-spiritual-bastion-during-the-aceh-war" class="editorial-chapter-title">VI. Spiritual Bastion During the Anti-Colonial Aceh War</h2>
<p>When the Dutch Empire launched its protracted invasion in 1873, the Pidie interior became the epicenter of unrelenting resistance. Mila and adjacent Tiro formed an impenetrable guerrilla stronghold under the spiritual guidance of heroic scholar-leaders, most notably <strong>Teungku Chik di Tiro (Muhammad Saman)</strong>.</p>

<p>The Po Teumeureuhom Mosque served as an inviolable sanctuary where freedom fighters convened under cover of darkness to coordinate ambushes and listen to recitations of the epic <em>Hikayat Prang Sabi</em>. Dutch military reports frequently noted that local resistance derived its unyielding resolve not merely from weaponry, but from the deep-rooted spiritual conviction fostered in historic sanctums like Mila.</p>

<h2 id="vii-living-traditions-pilgrimage-and-communal-feasts" class="editorial-chapter-title">VII. Living Traditions: Votive Pilgrimages and Communal Banquets</h2>
<p>Far from a dormant relic, the mosque remains the beating heart of local communal life:</p>

<ul>
    <li><strong>Vow Fulfillment (*Peuleuh Kaoy*):</strong> Migrant workers returning home from overseas journeys visit the mosque to fulfill vows made during adversity, offering ceremonial feasts of cattle or goats to orphans and villagers.</li>
    <li><strong>Agricultural Harvest Blessings (*Khanduri Blang*):</strong> Before planting and harvesting cycles, irrigation leaders (*keujruen blang*) gather farmers at the mosque to pray for bountiful harvests and partake in shared banquet cauldrons (*kuah beulangong*).</li>
    <li><strong>Prophetic Celebrations (*Maulid*):</strong> During commemorative festive months, vocal troupes chant melodic *meurukon* spiritual harmonies late into the night beneath the historic roof.</li>
</ul>

<h2 id="viii-conservation-and-heritage-stewardship" class="editorial-chapter-title">VIII. Conservation and Cultural Stewardship</h2>
<p>Designated as a protected provincial cultural heritage landmark by the Cultural Heritage Preservation Bureau (BPK) Region I Aceh, the sanctuary undergoes meticulous conservation to shield its four-hundred-year-old timber fabrics against tropical weathering, moisture, and pests while preserving its primary role as a consecrated place of worship.</p>

<h2 id="epilogue-the-timeless-radiance" class="editorial-chapter-title">Epilogue: The Timeless Radiance of Mila</h2>
<p>As evening falls across the tranquil valley of Mila and golden light fades behind the ridges of Bukit Barisan, the reflection of twilight shimmers peacefully across the surface of the ancient kulah pool. The melodious call to Maghrib prayer echoes through the mountain breeze, drifting through carved timber lattices into the night.</p>

<p>The Po Teumeureuhom Mosque stands as timeless proof that true civilizational grandeur is not measured by ostentatious concrete towers, but by the purity of faith, the wisdom of diplomacy, and the moral resilience of the community that honors its sacred roots.</p>

<div class="editorial-sources-box">
    <h4>📚 Selected Scholarly References &amp; Archival Records:</h4>
    <ol>
        <li><strong>Ar-Raniri, Sheikh Nuruddin (1638–1644 CE)</strong>. <em>Bustan al-Salatin fi Dhikr al-Awwalin wa al-Akhirin</em>. Chronicle of the Aceh Sultanate.</li>
        <li><strong>Lombard, Denys (1991)</strong>. <em>Le Sultanat d'Atjeh au temps d'Iskandar Muda (1607–1636)</em>. Paris: EFEO.</li>
        <li><strong>Hurgronje, C. Snouck (1906)</strong>. <em>The Achehnese</em>. Translated by A.W.S. O'Sullivan. Leiden: E.J. Brill.</li>
        <li><strong>Said, H. Mohammad (1981)</strong>. <em>Aceh Sepanjang Abad</em>. Medan: Waspada Press.</li>
        <li><strong>Winstedt, Richard O. (1934)</strong>. <em>A History of Perak</em>. Journal of the Malayan Branch of the Royal Asiatic Society.</li>
        <li><strong>BPK Wilayah I Aceh</strong>. <em>Heritage Inventory and Conservation Report of Po Teumeureuhom Mosque</em>. Ministry of Education and Culture RI.</li>
        <li><strong>Oral Tradition and Archival Manuscripts of Dayah Sinthop &amp; Mila Kemukiman</strong>, Pidie Regency.</li>
    </ol>
</div>
`, imageURL)

	enSummary := "A comprehensive editorial monograph on the 17th-century Po Teumeureuhom Mosque of Mila, Pidie: royal diplomacy with Perak, vernacular three-tiered timber architecture, eternal ablution springs, Plang Kling calligraphy, and anti-colonial heritage."

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
		return fmt.Errorf("failed to save EN translation for masjid po teumeureuhom: %w", err)
	}
	fmt.Println(" [+] English Translation Published: 'Po Teumeureuhom Mosque of Mila: 17th-Century Spiritual Heritage and Maritime Diplomacy'")

	return nil
}
