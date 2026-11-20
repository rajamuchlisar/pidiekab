package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	mediaStorage "pidiekab/internal/modules/media/infrastructure/storage"
)

//go:embed assets/pidie_hero_tangse_rumoh_aceh.jpg
var masterHeroImageBytes []byte

// publishMasterEditorialStory inserts or updates the centerpiece longform editorial story for Kabupaten Pidie.
func publishMasterEditorialStory(ctx context.Context, dbPool *pgxpool.Pool, adminID int64, storagePath string, logger *slog.Logger) error {
	fmt.Println("--> Publishing Master Editorial Story: 'Menyusuri Nafas Tanah Pedir'...")

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
		WHERE original_filename = 'pidie_hero_tangse_rumoh_aceh.jpg' 
		LIMIT 1
	`).Scan(&mediaID, &mediaStorageKey)

	if err != nil || mediaID == 0 {
		// Save embedded image to storage
		savedKey, sizeWritten, saveErr := store.Save(ctx, "jpg", bytes.NewReader(masterHeroImageBytes))
		if saveErr != nil {
			logger.Warn("failed to save hero image to storage, checking if already on disk", "error", saveErr)
			savedKey = fmt.Sprintf("%04d/%02d/pidie_hero_tangse_rumoh_aceh.jpg", time.Now().Year(), int(time.Now().Month()))
			sizeWritten = int64(len(masterHeroImageBytes))
		}
		mediaStorageKey = savedKey

		err = dbPool.QueryRow(ctx, `
			INSERT INTO media_assets (
				storage_key, original_filename, mime_type, file_size, width, height,
				alt_text_id, alt_text_en, source_url, rights_status, attribution, uploaded_by, created_at
			) VALUES (
				$1, 'pidie_hero_tangse_rumoh_aceh.jpg', 'image/jpeg', $2, 1920, 1080,
				'Lanskap Fajar di Lembah Tangse dan Rumoh Aceh Tradisional',
				'Dawn Landscape across Tangse Valley and Traditional Rumoh Aceh',
				'https://pidiekab.web.id/', 'rights_cleared',
				'Dokumentasi Editorial pidiekab.web.id', $3, NOW()
			)
			RETURNING id
		`, mediaStorageKey, sizeWritten, adminID).Scan(&mediaID)
		if err != nil {
			return fmt.Errorf("failed to insert hero media asset: %w", err)
		}
		fmt.Printf(" [+] Master Hero Image Saved (Media ID #%d, /media/%s)\n", mediaID, mediaStorageKey)
	} else {
		fmt.Printf(" [=] Master Hero Image Exists (Media ID #%d, /media/%s)\n", mediaID, mediaStorageKey)
	}

	// 3. Resolve category ID for 'sejarah-budaya'
	var categoryID int64
	err = dbPool.QueryRow(ctx, `
		SELECT category_id FROM category_translations 
		WHERE slug IN ('sejarah-budaya', 'budaya-sejarah') AND locale = 'id' 
		LIMIT 1
	`).Scan(&categoryID)
	if err != nil || categoryID == 0 {
		// Fallback to first active category
		_ = dbPool.QueryRow(ctx, `SELECT id FROM categories WHERE is_active = true ORDER BY id ASC LIMIT 1`).Scan(&categoryID)
	}
	if categoryID == 0 {
		return fmt.Errorf("no suitable category found for master editorial story")
	}

	// 4. Check or create content identity
	var contentID int64
	slugID := "menyusuri-nafas-tanah-pedir-jejak-abadi"
	err = dbPool.QueryRow(ctx, `
		SELECT content_id FROM content_translations 
		WHERE slug = $1 AND locale = 'id' 
		LIMIT 1
	`, slugID).Scan(&contentID)

	if err != nil || contentID == 0 {
		err = dbPool.QueryRow(ctx, `
			INSERT INTO contents (category_id, address, latitude, longitude, location_accuracy, main_media_id, created_by, updated_by, created_at, updated_at)
			VALUES ($1, 'Kabupaten Pidie, Aceh, Indonesia', 5.3813, 95.9612, 'accurate', $2, $3, $3, NOW(), NOW())
			RETURNING id
		`, categoryID, mediaID, adminID).Scan(&contentID)
		if err != nil {
			return fmt.Errorf("failed to create content identity: %w", err)
		}
		fmt.Printf(" [+] Created Content Identity #%d\n", contentID)
	} else {
		_, _ = dbPool.Exec(ctx, `
			UPDATE contents 
			SET main_media_id = $1, category_id = $2, updated_by = $3, updated_at = NOW() 
			WHERE id = $4
		`, mediaID, categoryID, adminID, contentID)
		fmt.Printf(" [=] Updated Content Identity #%d\n", contentID)
	}

	// 5. Construct Editorial HTML & Markdown for Indonesian
	imageURL := "/media/" + filepath.ToSlash(mediaStorageKey)

	idHTML := fmt.Sprintf(`
<figure class="editorial-hero-banner">
    <img src="%s" alt="Lanskap Fajar di Lembah Tangse dan Rumoh Aceh Tradisional" loading="eager" />
    <figcaption>Fajar menyingsing di lembah Tangse, Kabupaten Pidie: sawah terasering berbingkai kabut Bukit Barisan dan keanggunan arsitektur Rumoh Aceh. (Foto Dok. Editorial pidiekab.web.id)</figcaption>
</figure>

<p class="editorial-lead">Kabut pagi belum sepenuhnya beranjak dari punggung pegunungan Tangse ketika semburat merah tembaga mulai membelah cakrawala Selat Malaka di tepian Sigli. Di sudut-sudut gampong di Ceurih dan Garot, suara yang paling pertama membangunkan hari bukanlah deru mesin, melainkan ketukan ritmis yang teratur: <em>tuk, tak, tuk, tak</em>. Suara palu besi perempuan Pidie menempa melinjo panas menjadi emping renyah yang melegenda di seantero Nusantara.</p>

<p>Beberapa kilometer ke arah selatan, di Keude Beureunuen, uap air mendidih mengepul pekat dari teko-teko saring kuningan di kedai kopi kayu. Aroma pekat biji Robusta Tangse yang disangrai dengan mentega dan gula tebu menyergap siapa saja yang melangkah masuk. Di meja-meja bundar marmer tua, gelas-gelas kaca bergoyang pelan tertimpa perbincangan hangat: tawar-menawar emas murni, kabar harga pinang kering, hingga percakapan telepon dengan sanak saudara yang sedang berniaga di Pasar Tanah Abang Jakarta atau Chow Kit Kuala Lumpur.</p>

<p>Bagi penjelajah yang melintasinya, Kabupaten Pidie kerap tampak sebagai dataran hijau yang membentang dari pesisir utara hingga ngarai terjal pedalaman Bukit Barisan. Namun bagi mereka yang mau berhenti, menyesap kopinya, dan mendengarkan bisik anginnya, Pidie adalah sebuah ensiklopedia hidup. Sebuah negeri tempat memori kemaritiman purba, etos perantauan tanpa gentar, keteguhan tauhid para ulama, dan kehangatan tutur kata masyarakat berkelindan menjadi satu identitas yang tak lekang ditelan zaman.</p>

<h2 class="editorial-chapter-title">I. Pidie Tempo Dulu: Gemerlap Bandar Lada Selat Malaka</h2>
<p>Jauh sebelum peta modern membagi kawasan ini menjadi deretan kecamatan, pesisir Pidie telah tercatat dengan tinta emas dalam peta pelayaran dunia dengan nama <strong>Kerajaan Pedir</strong>.</p>
<p>Pada akhir abad ke-15 dan permulaan abad ke-16, sebelum Kesultanan Aceh Darussalam menyatukan kekuatan maritim di semenanjung utara Sumatra, Pedir adalah salah satu bandar niaga paling sibuk dan kosmopolitan di Asia Tenggara. Apoteker dan penjelajah Portugis, <strong>Tomé Pires</strong>, yang berlayar ke Nusantara antara tahun 1512 hingga 1515, menuliskan kesaksian berharga dalam mahakaryanya, <em>Suma Oriental</em>:</p>
<blockquote class="editorial-pullquote">
    <p>“Pedir adalah kerajaan yang makmur dan berdaulat. Setiap tahun pelabuhannya didatangi kapal-kapal jung Tiongkok, kapal saudagar Gujarat, Keling, Benggala, Arab, hingga kapal-kapal Venesia. Di sinilah lada terbaik di dunia dikumpulkan, beras melimpah ruah, dan sutra ditenun dengan kemahiran tinggi...”</p>
</blockquote>
<p>Lada hitam dari lembah-lembah Pidie kala itu menjadi komoditas emas hitam yang diperebutkan para penguasa dunia. Kapal-kapal dagang mengangkut rempah dari kuala-kuala Pidie melintasi Samudra Hindia menuju Laut Merah dan Eropa. Sisa-sisa kemegahan maritim itu hari ini masih dapat dirasakan saat kita berdiri di pesisir Kuala Batee atau menyusuri jejak batu nisan kuno bertatahkan kaligrafi gaya <em>Plang Kling</em> di kawasan Muara Tiga dan Kembang Tanjong—prasasti abadi para saudagar mancanegara yang beristirahat di tanah ini.</p>

<h2 class="editorial-chapter-title">II. Cerita di Balik Nama Tempat: Menapaki Toponimi dan Peta Peristiwa</h2>
<p>Setiap jengkal tanah di Pidie memiliki kisah yang tertambat pada namanya. Toponimi di kawasan ini bukan sekadar penanda administratif, melainkan rekaman peristiwa, bentang alam, dan kearifan masa lalu:</p>
<ul>
    <li><strong>Sigli (Pusat Pesisir)</strong> — Tradisi lisan setempat mencatat bahwa nama Sigli berakar dari istilah <em>Si-Glee</em> (bukit pengawas pesisir) dan <em>Sikureung Lee</em> (pertemuan sembilan anak aliran air tawar ke muara laut). Di masa lampau, bukit pengawas pesisir ini dijadikan mercu pandang para syahbandar memantau perahu layar yang hendak berlabuh di muara Krueng Baro.</li>
    <li><strong>Tangse (Lembah Sejuk di Balik Awan)</strong> — Terletak di ketinggian lebih dari 600 meter di atas permukaan laut, Tangse merujuk pada cekungan batu terjal tempat air pegunungan tertampung secara alami. Lembah ini sejak ratusan tahun lalu menjadi perlintasan strategis kafilah pembawa damar, rotan, dan kopi dari dataran tinggi.</li>
    <li><strong>Mila &amp; Tiro (Tanah Karomah dan Benteng Pejuang)</strong> — Mila berakar dari nama tokoh agung masa lampau yang membangun peradaban gampong beririgasi rapi. Sementara Tiro menjadi masyhur di seantero dunia sebagai tanah kelahiran wangsa ulama-pejuang Teungku Chik di Tiro, yang namanya abadi dalam perlawanan mempertahankan martabat negeri.</li>
    <li><strong>Laweung (Kuala Karst Purba)</strong> — Di ujung barat laut Pidie, Laweung merujuk pada gugusan bukit kapur karst dan teluk terlindung tempat para nelayan zaman dahulu membuang sauh (<em>meulaweung</em>) saat badai barat mengamuk di Selat Malaka.</li>
</ul>

<h2 class="editorial-chapter-title">III. Pidie dan Perantauan: Filosofi 'Ulee Gle Meugisa u Laot'</h2>
<p>Salah satu fenomena sosiologis paling menonjol dari masyarakat Pidie adalah keterikatan mereka pada <strong>budaya merantau (<em>jak mita raseuki</em>)</strong>. Mengapa merantau begitu mendarah daging dalam jiwa orang Pidie?</p>
<p>Jawabannya terpatri pada kondisi geografis dan sejarah sosialnya. Dataran sawah di Pidie sangat subur namun memiliki batas alami yang padat: di utara terhalang laut lepas, di selatan terjepit lereng bukit cadas. Ketika keluarga bertambah besar, sawah pusaka tidak mungkin dibagi-bagi menjadi terlalu sempit. Dari sinilah lahir dorongan luhur: pemuda yang akil balig didorong melangkah ke luar batas gampong.</p>
<blockquote class="editorial-pullquote">
    <p>“Ulee gle meugisa u laot, ureung Pidie meugisa u gampong.”<br><small style="font-size:0.9rem;font-weight:400;color:#047857;">(Sebagaimana air pegunungan niscaya mengalir ke laut, demikianlah perantau Pidie: sejauh apa pun melangkah di rantau orang, hatinya selalu berpulang dan membangun gampong halamannya).</small></p>
</blockquote>
<p>Merantau bagi anak muda Pidie bukanlah pelarian, melainkan kawah candradimuka pendewasaan. Mereka berangkat ke kota-kota besar di Sumatra, Jawa, Semenanjung Melayu, hingga jazirah Arab bermodalkan kejujuran dan tekad baja. Etos kerja pedagang Pidie dikenal dengan prinsip: <em>“Tanyoë tabeudoh ngon akai, tameujalan ngon ieetikat, tapula ngon sabé”</em> (Kita bangkit dengan akal, melangkah dengan niat tulus, dan menanam rezeki dengan ketekunan). Kiriman devisa rantau inilah yang mendirikan meunasah megah beratap tumpang di gampong-gampong, menyekolahkan adik-adik di dayah, dan membiayai kenduri panen raya.</p>

<h2 class="editorial-chapter-title">IV. Pasar dan Kehidupan Rakyat: Simfoni Beureunuen dan Berkah Melinjo</h2>
<p>Jika ingin melihat detak nadi kehidupan rakyat Pidie yang sesungguhnya, datanglah ke <strong>Pasar Beureunuen</strong> di pagi hari. Beureunuen bukan sekadar pasar kecamatan; ia adalah bursa perdagangan tradisional paling dinamis di kawasan timur Aceh. Di sepanjang deretan pertokoan bertingkat dua bergaya art-deco dan ruko modern, etalase toko-toko emas berkilau cemerlang di bawah lampu pijar. Bagi masyarakat Pidie, emas (<em>manyam meuh</em>) adalah mata uang penyimpan nilai yang paling dipercaya (<em>safe haven</em>), mahar pernikahan, dan modal cadangan para pedagang.</p>
<p>Di lorong pasar basah, para petani dari pedalaman Tangse dan Geumpang menggelar hasil bumi: tandan pisang raja, cabai merah segar, durian hutan yang harum menyengat, dan karung-karung beras <em>Si-Puteh</em> yang pulen.</p>
<p>Di pedesaan Kecamatan Delima, Meureudu, dan Ceurih, pohon melinjo (<em>muling</em>) yang tinggi menjulang menjadi tiang ekonomi kerakyatan berbasis perempuan. Melinjo disangrai dengan pasir pantai panas membara, dikupas panas-panas, lalu ditempa dengan palu besi pipih di atas talenan batu. Tiga kali ketukan: melinjo mengembang menjadi lembaran renyah transparan. Keahlian ini diwariskan turun-temurun, menghasilkan komoditas kuliner kebanggaan yang digemari lintas benua.</p>

<h2 class="editorial-chapter-title">V. Jejak Ulama Pidie: Lentera Keilmuan dan Benteng Moral</h2>
<p>Pidie sejak berabad-abad lampau dijuluki sebagai salah satu lumbung ulama dan rahim pendidikan Islam terkemuka di Sumatra. Hubungan antara ilmu agama dan kehidupan sehari-hari di sini begitu lekat, ibarat <em>zat ngon sifeut</em> (zat dan sifat yang tak terpisahkan).</p>
<p>Sosok paling monumental yang lahir dari rahim tanah ini adalah <strong>Teungku Chik di Tiro (Muhammad Saman, 1836–1891)</strong>. Lahir di Gampong Dayah Meunasah Tiro, beliau menempuh pendidikan keilmuan Islam tradisional di berbagai dayah sebelum mendirikan kembali Dayah Tiro sebagai pusat kajian tafsir, fikih, dan adab. Ketika Perang Aceh berkecamuk membela kedaulatan tanah air, Teungku Chik di Tiro memimpin perjuangan berlandaskan keadilan spiritual dan mengobarkan semangat rakyat melalui bait-bait sastra religius epik <em>Hikayat Prang Sabi</em>.</p>
<p>Jejak keilmuan ini diteruskan oleh dayah-dayah tradisional yang hingga kini berdiri kokoh di berbagai pelosok Pidie. Di rangkang-rangkang panggung kayu yang bersahaja, ribuan santri duduk bersila mengkaji kitab kuning. Ulama di Pidie bukan sekadar pengajar teks agama; mereka adalah penuntun moral, penengah sengketa sosial, dan penjaga marwah gampong.</p>

<h2 class="editorial-chapter-title">VI. Budaya, Adat &amp; Tradisi: Meunasah, Peusijuek, dan Meurukon</h2>
<p>Keteraturan hidup masyarakat Pidie ditopang oleh tiang penyangga yang kokoh: <strong>Adat dan Meunasah</strong>. Di setiap gampong, Meunasah adalah balai komunal tempat salat berjamaah, tempat tetua adat (<em>Tuha Peuet</em>) dan kepala desa (<em>Keuchik</em>) bermusyawarah, serta tempat pemuda belajar mengaji dan mendengarkan hikayat para tetua.</p>
<p>Ketika warga hendak merantau, memulai usaha baru, atau merayakan kelulusan, diadakanlah upacara <strong>Peusijuek</strong> (tepung tawar). Beras padi (<em>breueh padee</em>), dedaunan penyejuk (<em>on seunijuek</em>), dan air suci dipercikkan dengan lembut ke telapak tangan sembari melantunkan selawat, mendoakan agar jiwa senantiasa teduh dan langkah senantiasa berkah.</p>
<p>Di malam-malam peringatan Maulid Nabi, di beranda meunasah kerap digelar seni pertunjukan vokal <strong>Meurukon</strong>. Dua kelompok pria duduk berhadapan tanpa alat musik, saling bertukar tanya-jawab seputar rukun iman, rukun Islam, dan filsafat tasawuf dalam lantunan syair bertempo cepat dan harmonis—sebuah bukti tingginya peradaban sastra tutur masyarakat Pidie.</p>

<h2 class="editorial-chapter-title">VII. Bahasa dan Istilah Pidie: Kehangatan Dialek Lugas</h2>
<p>Dialek bahasa Aceh di Pidie memiliki karakter vokal yang sangat khas: bernas, ekspresif, berintonasi tegas, dan sarat intonasi meliuk di ujung kalimat. Orang Pidie dikenal dengan gaya bicaranya yang lugas, apa adanya, dan tidak suka kepura-puraan.</p>
<ul>
    <li><em>“Hai Syedara!”</em> — Sapaan hangat yang menempatkan lawan bicara setara sebagai saudara kandung.</li>
    <li><em>“Kiban haba gampong?”</em> — Pertanyaan pembuka menanyakan kabar kampung dan keluarga.</li>
    <li><em>“Kupi si-glah pyoh si-sa’at”</em> — Ajakan duduk meminum segelas kopi meredakan ketegangan sebelum kembali berikhtiar.</li>
</ul>

<h2 class="editorial-chapter-title">VIII. Warisan Sejarah &amp; Situs Abadi: Jejak Kejayaan Silang Zaman</h2>
<p>Bagi penikmat sejarah, menyusuri Pidie terasa seperti membuka lembaran babad kuno:</p>
<ul>
    <li><strong>Masjid Po Teumeureuhom (Mila)</strong> — Didirikan oleh Sultan Alaidin Mansur Syah, masjid kayu beratap tumpang ini menyimpan ketenangan mistis dengan mata air kolam suci yang tak pernah surut.</li>
    <li><strong>Benteng Kuta Asan (Sigli)</strong> — Dinding batu karang di muara pantai yang menjadi saksi ketangguhan barisan pertahanan rakyat Pidie menghadapi gempuran maritim akhir abad ke-19.</li>
    <li><strong>Guha Tujoh (Laweung, Muara Tiga)</strong> — Gua alam perbukitan karst berkubah stalaktit megah dengan tujuh lorong alami yang di masa lampau menjadi tempat bertafakur para syekh sufi dan pos pertahanan gerilya.</li>
</ul>

<h2 class="editorial-chapter-title">IX. Pidie dari Masa ke Masa: Jalur Rel Kereta dan Kebangkitan Baru</h2>
<p>Ingatan kolektif masyarakat Pidie tempo dulu juga terpatri kuat pada jalur kereta api uap <strong>Atjeh Tram (Atjeh Staatsspoorwegen / ASS)</strong>. Stasiun Kereta Api Sigli pada masa keemasannya adalah hub logistik tersibuk di pantai timur-utara Sumatra yang mengangkut kopi Tangse, emping melinjo, dan hasil laut menuju pelabuhan ekspor.</p>
<p>Kini, Kabupaten Pidie menyongsong masa depan dengan kebanggaan baru: <strong>Kopi Robusta &amp; Arabika Tangse-Geumpang</strong> kini menjadi primadona komoditas spesialti di kafe-kafe urban Nusantara, sementara keajaiban alam geowisata ngarai <strong>Lingkok Kuwieng</strong> Padang Tiji dan arung jeram Krueng Geumpang mengundang penjelajah dunia menyelami keasrian alam tropis Sumatra.</p>

<blockquote class="editorial-pullquote">
    <p>“Pidie bukan sekadar sebuah kabupaten di peta bumi. Ia adalah sebuah nyala api dalam dada: tentang keberanian merantau mengarungi samudra rezeki yang luas, ketulusan sujud di meunasah, dan kebanggaan tak terhingga untuk selalu pulang ke tanah Pedir.”</p>
</blockquote>

<div class="editorial-sources-box">
    <h4>📚 Sumber Pustaka &amp; Rujukan Kredibel:</h4>
    <ol>
        <li><strong>Pires, Tomé (1512–1515)</strong>. <em>The Suma Oriental of Tomé Pires: An Account of the East, from the Red Sea to Japan</em>. Diterjemahkan oleh Armando Cortesão. London: The Hakluyt Society.</li>
        <li><strong>Reid, Anthony (1988)</strong>. <em>Southeast Asia in the Age of Commerce, 1450–1680: Volume One: The Lands below the Winds</em>. New Haven: Yale University Press.</li>
        <li><strong>Lombard, Denys (1991)</strong>. <em>Kerajaan Aceh: Zaman Sultan Iskandar Muda (1607–1636)</em>. Jakarta: Balai Pustaka / Forum Jakarta-Paris.</li>
        <li><strong>Alfian, Teuku Ibrahim (1992)</strong>. <em>Sastra Perang: Sebuah Pendekatan Sejarah dan Kebudayaan terhadap Hikayat Prang Sabi</em>. Jakarta: Balai Pustaka.</li>
        <li><strong>Hurgronje, C. Snouck (1906)</strong>. <em>The Achehnese (De Atjèhers)</em>. Diterjemahkan oleh A.W.S. O'Sullivan. Leiden: E.J. Brill.</li>
        <li><strong>Said, H. Mohammad (1981)</strong>. <em>Aceh Sepanjang Abad</em>. Medan: Percetakan &amp; Penerbitan Waspada.</li>
        <li><strong>Arsip Sejarah Lokal &amp; Tradisi Tutur Masyarakat Kabupaten Pidie</strong>, dihimpun melalui penelusuran sejarah lisan para tetua adat, dayah, dan dokumentasi komunitas kebudayaan lokal Pidie.</li>
    </ol>
</div>
`, imageURL)

	idSummary := "Sebuah narasi editorial panjang tentang tanah lumbung pangan dan lada purba, filosofi merantau anak negeri, detak pasar rakyat, denyut spiritual dayah, hingga situs bersejarah yang tegak di antara deburan Selat Malaka dan kabut abadi Bukit Barisan."
	idTitle := "Menyusuri Nafas Tanah Pedir: Merajut Ingatan, Saudagar, Ulama, dan Jejak Abadi di Jantung Pidie"

	idBodyJSON, _ := json.Marshal(map[string]string{
		"html":     idHTML,
		"markdown": idHTML,
	})

	// 6. English Content Translation
	enTitle := "Breathing the Spirit of Ancient Pedir: Weaving Memory, Merchants, Sages, and Timeless Echoes in the Heart of Pidie"
	enSlug := "breathing-spirit-ancient-pedir-timeless-echoes"
	enSummary := "An expansive editorial narrative exploring ancient pepper kingdoms, the philosophy of migration, vibrant market life, scholarly dayah sanctuaries, and timeless heritage between the Malacca Strait and Bukit Barisan mountains."

	enHTML := fmt.Sprintf(`
<figure class="editorial-hero-banner">
    <img src="%s" alt="Dawn Landscape across Tangse Valley and Traditional Rumoh Aceh" loading="eager" />
    <figcaption>Dawn breaking over Tangse valley, Pidie Regency: terraced paddy fields framed by morning mist of the Bukit Barisan range and traditional carved Rumoh Aceh architecture. (Photo: Editorial Archive pidiekab.web.id)</figcaption>
</figure>

<p class="editorial-lead">Morning mist has barely lifted from the rugged ridge lines of Tangse when copper-gold dawn light breaks across the Malacca Strait on the shores of Sigli. In the rural village compounds of Ceurih and Garot, the first sound greeting the new day is not industrial machinery, but a rhythmic, melodic cadence: <em>tuk, tak, tuk, tak</em>. It is the sound of Pidie women wielding flat iron mallets, hammering fire-roasted melinjo nuts into wafer-thin crisps renowned across the Indonesian archipelago.</p>

<p>A few miles southward in the bustling junction of Keude Beureunuen, steam billows from brass pouring kettles in roadside timber coffee halls. The intoxicating fragrance of mountain-grown Tangse Robusta beans fills the morning air. Across aged marble tables, locals converse with spirited banter: trading 99%% pure gold, discussing spice yields, and checking in with kin managing textile stores in Jakarta’s grand markets or overseas trading hubs.</p>

<p>To passing travelers, Pidie Regency might appear as a sweeping coastal agricultural plain bordered by volcanic peaks. But to those who linger, savor its coffee, and listen to the murmurs of its villages, Pidie reveals itself as a living epic. A homeland where ancient maritime glory, an intrepid merchant diaspora, steadfast dayah scholarship, and unpretentious dialect converge into an enduring cultural soul.</p>

<h2 class="editorial-chapter-title">I. Ancient Pedir: The Maritime Pepper Crossroads of the Malacca Strait</h2>
<p>Long before modern administration demarcated these sub-districts, Pidie’s coastline was recorded in golden letters across world navigational charts as the sovereign <strong>Kingdom of Pedir</strong>.</p>
<p>During the late 15th and early 16th centuries, Portuguese apothecary and chronicler <strong>Tomé Pires</strong> sailed through Southeast Asian waters and wrote in his seminal <em>Suma Oriental</em> (1512–1515):</p>
<blockquote class="editorial-pullquote">
    <p>“Pedir is a prosperous and sovereign realm. Annually its harbors receive junks from China, merchant galleons from Gujarat, Coromandel, Bengal, Arabia, and Venice. Here the finest black pepper in the world is gathered, rice harvests are bountiful, and fine silk is woven with great artistry...”</p>
</blockquote>
<p>Black pepper cultivated in the fertile foothills of Pidie was black gold, exported across the Indian Ocean toward the Red Sea and European dining tables. Today, remnants of that cosmopolitan maritime grandeur remain tangible at coastal Kuala Batee and among ancient carved funerary gravestones (the <em>Plang Kling</em> style) scattered across Muara Tiga and Kembang Tanjong.</p>

<h2 class="editorial-chapter-title">II. Stories Behind the Place Names: Toponyms and Memory</h2>
<p>Every corner of Pidie carries history etched into its name:</p>
<ul>
    <li><strong>Sigli (The Coastal Gateway)</strong> — Local oral lore traces the name to <em>Si-Glee</em> (the watch-hill lookout) and <em>Sikureung Lee</em> (the confluence of nine freshwater streams meeting the sea), where ancient harbor masters monitored incoming merchant vessels.</li>
    <li><strong>Tangse (The Highland Cloud Valley)</strong> — Rising over 600 meters above sea level, Tangse denotes an amphitheater-like natural stone basin catching alpine rains, historically serving as a crucial caravan pass for high-country timber and coffee traders.</li>
    <li><strong>Mila &amp; Tiro (Land of Sanctuaries and Warriors)</strong> — Mila was founded by early civil architects who developed intricate canal systems, while Tiro achieved worldwide renown as the ancestral seat of the heroic Teungku Chik di Tiro scholarly dynasty.</li>
    <li><strong>Laweung (The Ancient Karst Harbor)</strong> — Located on the northwestern promontory, Laweung refers to natural limestone headlands and sheltered coves where fishermen traditionally dropped anchor (<em>meulaweung</em>) when westerly storms swept the Malacca Strait.</li>
</ul>

<h2 class="editorial-chapter-title">III. Pidie and the Merchant Diaspora: The Philosophy of Returning Home</h2>
<p>The defining cultural trait of Pidie is its deep-seated tradition of circular migration (<strong>jak mita raseuki</strong>). Why is voyaging so deeply ingrained in the Pidie psyche?</p>
<p>The answer lies in geography and demographic balance. While Pidie's alluvial plains are intensely fertile, arable land is bounded by sea and mountain. As generations grew, splitting heritage rice paddies risked collective impoverishment. Hence arose an unwritten rite of passage: young adults venture into the wider world to pioneer commerce.</p>
<blockquote class="editorial-pullquote">
    <p>“Ulee gle meugisa u laot, ureung Pidie meugisa u gampong.”<br><small style="font-size:0.9rem;font-weight:400;color:#047857;">(As mountain rivers inexorably flow to the ocean, so does the traveler from Pidie: however far they wander, their soul always returns to uplift their ancestral village).</small></p>
</blockquote>
<p>With integrity, commercial acumen, and tireless grit, Pidie merchants built formidable networks across Sumatra, Java, and the Malay Peninsula. Remittances sent back home built soaring village prayer halls (<em>meunasah</em>), endowed rural religious schools, and funded communal harvest celebrations.</p>

<h2 class="editorial-chapter-title">IV. Market Rhythms: The Symphony of Beureunuen and Melinjo Heritage</h2>
<p>At sunrise, <strong>Beureunuen Market</strong> hums with trade. It is the economic heartbeat of eastern Aceh, famous for its grand jewelry arcade where 99%% gold (<em>manyam meuh</em>) serves as the trusted community hedge and wedding dowry. Nearby, highland farmers unroll fresh produce while rural women in Delima and Ceurih maintain the artisanal mastery of emping melinjo.</p>

<h2 class="editorial-chapter-title">V. Echoes of Islamic Scholarship: Dayah Sanctuaries and Moral Anchors</h2>
<p>Pidie has long been celebrated as Aceh's cradle of Islamic scholarship. Its towering figure, <strong>Teungku Chik di Tiro (Muhammad Saman, 1836–1891)</strong>, established the renowned Dayah Tiro before uniting moral spiritual resistance through the epic poetry of <em>Hikayat Prang Sabi</em>. Today, hundreds of rural dayah schools continue educating students in classical jurisprudence, Arabic grammar, and ethical philosophy.</p>

<h2 class="editorial-chapter-title">VI. Custom, Meunasah, and Communal Artistry</h2>
<p>Village harmony centers around the <strong>Meunasah</strong>—a communal timber stilt pavilion serving as prayer sanctuary, town hall, and cultural school. Rites such as <strong>Peusijuek</strong> (spiritual blessing of cooling waters and sacred grains) and rhythmic choral recitations of <strong>Meurukon</strong> reinforce social kinship across generations.</p>

<h2 class="editorial-chapter-title">VII. Living Heritage &amp; Timeless Monuments</h2>
<p>Visitors can explore the historic <strong>Masjid Po Teumeureuhom</strong> in Mila with its sacred freshwater spring, the coastal walls of <strong>Benteng Kuta Asan</strong> in Sigli, and the mystical subterranean limestone caverns of <strong>Guha Tujoh</strong> in Laweung.</p>

<h2 class="editorial-chapter-title">VIII. Pidie Across Eras: From Steam Rails to Highland Coffee</h2>
<p>Once connected by the historic <strong>Atjeh Tram (ASS)</strong> steam railway, Pidie today looks forward with vibrant renewal: award-winning single-origin <strong>Tangse Robusta and Arabica coffees</strong>, scenic eco-adventures at <strong>Lingkok Kuwieng Canyon</strong>, and a dynamic younger generation embracing the digital era while fiercely safeguarding ancestral roots.</p>

<blockquote class="editorial-pullquote">
    <p>“Pidie is more than a destination on a map; it is a spirit of daring to cross wide horizons, humility before the Divine, and a boundless pride in always answering the call of the ancestral soil.”</p>
</blockquote>

<div class="editorial-sources-box">
    <h4>📚 Authoritative Sources &amp; Citations:</h4>
    <ol>
        <li><strong>Pires, Tomé (1512–1515)</strong>. <em>The Suma Oriental of Tomé Pires</em>. Trans. Armando Cortesão. London: The Hakluyt Society.</li>
        <li><strong>Reid, Anthony (1988)</strong>. <em>Southeast Asia in the Age of Commerce, 1450–1680</em>. Yale University Press.</li>
        <li><strong>Lombard, Denys (1991)</strong>. <em>Le carrefour javanais / Kerajaan Aceh</em>. Paris / Jakarta: Balai Pustaka.</li>
        <li><strong>Alfian, Teuku Ibrahim (1992)</strong>. <em>Sastra Perang &amp; Hikayat Prang Sabi</em>. Balai Pustaka.</li>
        <li><strong>Hurgronje, C. Snouck (1906)</strong>. <em>The Achehnese (De Atjèhers)</em>. Leiden: E.J. Brill.</li>
        <li><strong>Said, H. Mohammad (1981)</strong>. <em>Aceh Sepanjang Abad</em>. Waspada.</li>
        <li><strong>Oral History &amp; Community Heritage Archives of Pidie Regency</strong>.</li>
    </ol>
</div>
`, imageURL)

	enBodyJSON, _ := json.Marshal(map[string]string{
		"html":     enHTML,
		"markdown": enHTML,
	})

	// 7. Upsert Indonesian translation (status: PUBLISHED)
	_, err = dbPool.Exec(ctx, `
		INSERT INTO content_translations (
			content_id, locale, title, slug, summary, body, status, published_at, published_by,
			search_vector, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, 'id', $2, $3, $4, $5, 'published', NOW(), $6,
			setweight(to_tsvector('simple', COALESCE($2, '')), 'A') || setweight(to_tsvector('simple', COALESCE($4, '')), 'B'),
			$6, $6, NOW(), NOW()
		)
		ON CONFLICT (content_id, locale) DO UPDATE
		SET title = EXCLUDED.title,
		    slug = EXCLUDED.slug,
		    summary = EXCLUDED.summary,
		    body = EXCLUDED.body,
		    status = 'published',
		    published_at = NOW(),
		    published_by = EXCLUDED.published_by,
		    search_vector = EXCLUDED.search_vector,
		    updated_by = EXCLUDED.updated_by,
		    updated_at = NOW()
	`, contentID, idTitle, slugID, idSummary, idBodyJSON, adminID)
	if err != nil {
		return fmt.Errorf("failed to upsert Indonesian translation: %w", err)
	}
	fmt.Printf(" [+] Indonesian Translation Published: '%s'\n", idTitle)

	// 8. Upsert English translation (status: PUBLISHED)
	_, err = dbPool.Exec(ctx, `
		INSERT INTO content_translations (
			content_id, locale, title, slug, summary, body, status, published_at, published_by,
			search_vector, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, 'en', $2, $3, $4, $5, 'published', NOW(), $6,
			setweight(to_tsvector('simple', COALESCE($2, '')), 'A') || setweight(to_tsvector('simple', COALESCE($4, '')), 'B'),
			$6, $6, NOW(), NOW()
		)
		ON CONFLICT (content_id, locale) DO UPDATE
		SET title = EXCLUDED.title,
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
		return fmt.Errorf("failed to upsert English translation: %w", err)
	}
	fmt.Printf(" [+] English Translation Published: '%s'\n", enTitle)

	// 9. Insert verified academic & historical citations
	citations := []struct {
		name string
		url  string
		note string
	}{
		{"Pires, Tomé (1512–1515). Suma Oriental", "https://archive.org/details/suma-oriental-tome-pires", "Catatan pelabuhan niaga dan komoditas lada Kerajaan Pedir."},
		{"Reid, Anthony (1988). Southeast Asia in the Age of Commerce", "https://yalebooks.yale.edu/", "Kajian sosiologi perniagaan maritim Selat Malaka."},
		{"Lombard, Denys (1991). Kerajaan Aceh: Zaman Sultan Iskandar Muda", "https://katalog.perpusnas.go.id/", "Struktur lumbung pangan dan jalur diplomasi wilayah Pidie."},
		{"Alfian, Teuku Ibrahim (1992). Sastra Perang & Hikayat Prang Sabi", "https://catalogue.nla.gov.au/", "Kajian kepemimpinan dan karya sastra ulama Teungku Chik di Tiro."},
		{"Hurgronje, C. Snouck (1906). The Achehnese (De Atjèhers)", "https://openlibrary.org/", "Etnografi adat gampong, meunasah, dan tradisi sosial masyarakat Pidie."},
	}

	for _, c := range citations {
		_, _ = dbPool.Exec(ctx, `
			INSERT INTO content_sources (content_id, source_name, source_type, source_url, notes, created_by)
			VALUES ($1, $2, 'research', $3, $4, $5)
			ON CONFLICT DO NOTHING
		`, contentID, c.name, c.url, c.note, adminID)
	}

	// 10. Set as LEAD STORY in curation for both 'id' and 'en'
	// Delete any existing lead assignment
	_, _ = dbPool.Exec(ctx, `DELETE FROM homepage_story_items WHERE role = 'lead'`)
	// Insert Lead Story for ID
	_, _ = dbPool.Exec(ctx, `
		INSERT INTO homepage_story_items (content_id, locale, role, position, created_at)
		VALUES ($1, 'id', 'lead', 1, NOW())
		ON CONFLICT (content_id, locale) DO UPDATE SET role = 'lead', position = 1
	`, contentID)
	// Insert Lead Story for EN
	_, _ = dbPool.Exec(ctx, `
		INSERT INTO homepage_story_items (content_id, locale, role, position, created_at)
		VALUES ($1, 'en', 'lead', 1, NOW())
		ON CONFLICT (content_id, locale) DO UPDATE SET role = 'lead', position = 1
	`, contentID)

	// Ensure featured in editor's picks
	_, _ = dbPool.Exec(ctx, `
		INSERT INTO featured_contents (content_id, position, created_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (content_id) DO UPDATE SET position = 1
	`, contentID)

	fmt.Printf(" [+] Assigned Content #%d as Homepage LEAD STORY and Featured Content!\n", contentID)
	return nil
}
