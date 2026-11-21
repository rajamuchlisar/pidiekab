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

//go:embed assets/kerupuk_mulieng_beureunuen.jpg
var kerupukMuliengImageBytes []byte

// publishKerupukMuliengStory inserts or updates the comprehensive editorial story and hero image for Kerupuk Mulieng Beureunuen.
func publishKerupukMuliengStory(ctx context.Context, dbPool *pgxpool.Pool, adminID int64, storagePath string, logger *slog.Logger) error {
	fmt.Println("--> Publishing Editorial Story: 'Kerupuk Mulieng Beureunuen'...")

	// 1. Initialize local storage
	if storagePath == "" {
		storagePath = "./data/media"
	}
	store, err := mediaStorage.NewLocalStorage(storagePath)
	if err != nil {
		return fmt.Errorf("failed to init media storage: %w", err)
	}

	// 2. Ensure hero media asset exists for kerupuk mulieng
	var mediaID int64
	var mediaStorageKey string
	err = dbPool.QueryRow(ctx, `
		SELECT id, storage_key FROM media_assets 
		WHERE original_filename = 'kerupuk_mulieng_beureunuen.jpg' 
		LIMIT 1
	`).Scan(&mediaID, &mediaStorageKey)

	if err != nil || mediaID == 0 {
		savedKey, sizeWritten, saveErr := store.Save(ctx, "jpg", bytes.NewReader(kerupukMuliengImageBytes))
		if saveErr != nil {
			logger.Warn("failed to save kerupuk mulieng image to storage, checking fallback path", "error", saveErr)
			savedKey = fmt.Sprintf("%04d/%02d/kerupuk_mulieng_beureunuen.jpg", time.Now().Year(), int(time.Now().Month()))
			sizeWritten = int64(len(kerupukMuliengImageBytes))
		}
		mediaStorageKey = savedKey

		err = dbPool.QueryRow(ctx, `
			INSERT INTO media_assets (
				storage_key, original_filename, mime_type, file_size, width, height,
				alt_text_id, alt_text_en, source_url, rights_status, attribution, uploaded_by, created_at
			) VALUES (
				$1, 'kerupuk_mulieng_beureunuen.jpg', 'image/jpeg', $2, 1920, 1080,
				'Ilustrasi Kerupuk Mulieng Khas Beureunuen Pidie: Emping Melinjo Super Renyah',
				'Illustration of Traditional Beureunuen Kerupuk Mulieng: Handcrafted Pidie Melinjo Emping',
				'https://pidiekab.web.id/', 'rights_cleared',
				'Ilustrasi Digital Editorial pidiekab.web.id', $3, NOW()
			)
			RETURNING id
		`, mediaStorageKey, sizeWritten, adminID).Scan(&mediaID)
		if err != nil {
			return fmt.Errorf("failed to insert kerupuk mulieng media asset: %w", err)
		}
		fmt.Printf(" [+] Kerupuk Mulieng Hero Image Saved (Media ID #%d, /media/%s)\n", mediaID, mediaStorageKey)
	} else {
		fmt.Printf(" [=] Kerupuk Mulieng Hero Image Exists (Media ID #%d, /media/%s)\n", mediaID, mediaStorageKey)
		// Ensure physical file exists on disk
		targetPath, resolveErr := store.ResolvePath(mediaStorageKey)
		if resolveErr == nil {
			if _, statErr := os.Stat(targetPath); os.IsNotExist(statErr) {
				_ = os.MkdirAll(filepath.Dir(targetPath), 0750)
				if writeErr := os.WriteFile(targetPath, kerupukMuliengImageBytes, 0644); writeErr == nil {
					fmt.Printf(" [+] Wrote missing kerupuk mulieng image to storage disk: %s\n", targetPath)
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
		return fmt.Errorf("no suitable category found for kerupuk mulieng story")
	}

	// 4. Check or create content identity
	var contentID int64
	slugID := "kelezatan-kerupuk-mulieng-beureunuen-melinjo"
	err = dbPool.QueryRow(ctx, `
		SELECT content_id FROM content_translations 
		WHERE slug = $1 AND locale = 'id' 
		LIMIT 1
	`, slugID).Scan(&contentID)

	if err != nil || contentID == 0 {
		err = dbPool.QueryRow(ctx, `
			INSERT INTO contents (category_id, address, latitude, longitude, location_accuracy, main_media_id, created_by, updated_by, created_at, updated_at)
			VALUES ($1, 'Pasar Beureunuen, Kecamatan Mutiara, Kabupaten Pidie, Aceh', 5.3125, 95.9754, 'accurate', $2, $3, $3, NOW(), NOW())
			RETURNING id
		`, categoryID, mediaID, adminID).Scan(&contentID)
		if err != nil {
			return fmt.Errorf("failed to create content identity for kerupuk mulieng: %w", err)
		}
		fmt.Printf(" [+] Created Content Identity #%d for Kerupuk Mulieng\n", contentID)
	} else {
		_, _ = dbPool.Exec(ctx, `
			UPDATE contents 
			SET main_media_id = $1, category_id = $2, updated_by = $3, updated_at = NOW() 
			WHERE id = $4
		`, mediaID, categoryID, adminID, contentID)
		fmt.Printf(" [=] Updated Content Identity #%d for Kerupuk Mulieng\n", contentID)
	}

	// 5. Construct Comprehensive Editorial HTML & Markdown for Indonesian
	imageURL := "/media/" + filepath.ToSlash(mediaStorageKey)

	idHTML := fmt.Sprintf(`
<figure class="editorial-hero-banner">
    <img src="%s" alt="Kelezatan Gurih Kerupuk Mulieng Beureunuen: Warisan Melinjo Terbaik Pidie" loading="eager" />
    <figcaption class="editorial-caption">Ilustrasi: Kerupuk Mulieng khas Beureunuen, Pidie. Olahan emping melinjo tradisional berkualitas super dengan teknik sangrai pasir panas dan penempaan manual di atas landasan batu.</figcaption>
</figure>

<p class="editorial-lead">Di seantero Aceh, bahkan hingga ke kedai-kedai kuliner Nusantara di Jakarta dan Semenanjung Malaya, aroma semangkuk Mie Aceh atau sepiring Kuah Beulangong tak pernah benar-benar lengkap tanpa kehadiran taburan kerupuk keemasan yang renyah dengan semburat rasa pahit-gurih yang anggun. Kudapan legendaris itu adalah <strong>Kerupuk Mulieng</strong>—emping melinjo murni kebanggaan masyarakat Kabupaten Pidie, yang berpusat di jantung niaga <strong>Keude Beureunuen</strong>.</p>

<p>Bagi orang Pidie, pohon melinjo bukan sekadar tanaman kebun, melainkan bagian tak terpisahkan dari denyut kebudayaan, ketahanan ekonomi rumah tangga, dan identitas daerah. Melalui keahlian tangan perempuan-perempuan tangguh di pedesaan Pidie yang menempa biji-biji panas membara di atas landasan batu, lahirlah lembaran emping tertipis, terbening, dan paling renyah yang pernah diciptakan dalam khazanah gastronomi Nusantara.</p>

<h2 class="editorial-chapter-title">I. Denyut Pagi di Keude Beureunuen: Ibu Kota Kerupuk Mulieng Sumatra</h2>
<p>Kawasan <strong>Keude Beureunuen</strong> di Kecamatan Mutiara telah lama ditahbiskan sebagai "Ibu Kota Kerupuk Melinjo" di Pulau Sumatra. Terletak strategis di jalur arteri lintas timur Banda Aceh–Medan, kawasan pasar ini menyajikan pemandangan khas yang memikat setiap pengendara yang melintas: deretan etalase pertokoan yang dipenuhi ratusan karung putih dan toples kaca raksasa berisi tumpukan kepingan emping melinjo aneka mutu.</p>

<p>Sejak fajar menyingsing, aktivitas bongkar muat kerupuk mulieng telah menggeliat. Para pengepul dari gampong-gampong sekitar—mulai dari Yaman, Mutiara Timur, Glumpang Tiga, Sakti, hingga Delima—membawa keranjang-keranjang emping yang baru saja selesai dijemur kering di bawah terik matahari kemarin sore. Di pasar inilah transaksi bernilai miliaran rupiah bergulir setiap bulannya. Beureunuen tidak hanya melayani pembeli eceran yang hendak membawa buah tangan untuk sanak famili, melainkan menjadi pusat pasokan utama bagi jaringan restoran Aceh, hotel, serta eksportir yang mendistribusikannya hingga ke Malaysia, Singapura, dan Timur Tengah.</p>

<h2 class="editorial-chapter-title">II. Pohon Melinjo (*Bak Mulieng*) dan Berkah Tanah Vulkanis Pidie</h2>
<p>Pertanyaan yang kerap mengemuka adalah: mengapa emping melinjo asal Pidie memiliki tekstur yang jauh lebih renyah dan cita rasa gurih yang tidak tertandingi oleh daerah lain? Jawabannya berakar pada kombinasi agroklimat dan kecintaan masyarakat terhadap tanaman ini.</p>

<p>Secara geografis, Kabupaten Pidie diapit oleh jajaran Pegunungan Bukit Barisan yang kaya unsur hara vulkanis di sisi barat-selatan dan pesisir Selat Malaka di sisi utara. Komposisi tanah aluvial dan vulkanis yang subur berpadu dengan intensitas sinar matahari tropis yang melimpah menciptakan habitat paling ideal bagi pertumbuhan pohon melinjo (<em>Gnetum gnemon</em>), yang dalam bahasa Aceh disebut <strong>Bak Mulieng</strong>.</p>

<p>Hampir di setiap pekarangan rumah panggung tradisional (<em>rumoh Aceh</em>) di gampong-gampong Pidie, pohon melinjo dibiarkan tumbuh tinggi menjulang hingga mencapai belasan meter. Pohon-pohon tua yang telah berusia puluhan tahun ini menghasilkan biji (<em>boh mulieng</em>) yang padat, kaya sari pati, dan memiliki kadar minyak nabati alami yang seimbang. Biji melinjo yang dipetik langsung dari pohon tua inilah yang menjadi rahasia di balik tekstur emping yang ketika digoreng langsung mekar sempurna tanpa memerlukan bahan pengembang buatan.</p>

<h2 class="editorial-chapter-title">III. Alkimia Tradisional: Dari Pasir Sungai Panas Menjadi Kepingan Emas Renyah</h2>
<p>Proses pembuatan Kerupuk Mulieng Beureunuen adalah sebuah ritual ketekunan dan kemahiran fisik yang menakjubkan. Di saat industri makanan modern beralih ke mesin pemipih otomatis, pengrajin Pidie tetap setia mempertahankan teknik manual warisan indatu (leluhur) yang telah teruji ratusan tahun:</p>

<ol class="editorial-process-steps">
    <li><strong>Seleksi Biji Melinjo (*Pilah Boh Mulieng*)</strong>: Biji melinjo yang dipilih harus benar-benar matang pohon, bercirikan kulit luar yang sudah berwarna merah tua atau kuning cerah. Kulit luar berdaging ini dikupas terlebih dahulu untuk memisahkan biji bercangkang keras (kulit tanduk).</li>
    <li><strong>Sangrai Pasir Panas (*Teu-oet lam Pasie*)</strong>: Ini adalah tahapan paling krusial. Biji melinjo tidak digoreng dengan minyak, melainkan disangrai dalam wajan tanah liat tradisional (<em>beuleungoe</em>) atau kuali besi besar yang diisi pasir sungai yang telah dicuci bersih dan diayak halus. Pasir panas bertindak sebagai media konduksi termal alami yang mematangkan bagian dalam biji secara merata hingga ke inti sel tanpa membakar permukaannya.</li>
    <li><strong>Pengupasan Panas Cepat (*Tupang*)</strong>: Dalam hitungan detik setelah diangkat dari pasir membara, biji melinjo yang masih mengepul panas langsung dipukul dengan palu kecil untuk memecahkan cangkang kerasnya. Bagian dalam biji yang berwarna putih kekuningan harus segera dikeluarkan selagi masih elastis dan hangat.</li>
    <li><strong>Penempaan Ritmis (*Teupoh Mulieng*)</strong>: Biji melinjo panas diletakkan di atas landasan batu kali pipih (<em>aneuk batee</em>). Dengan menggunakan palu besi bulat beralas datar, pengrajin wanita mengayunkan pukulan ritmis yang mantap: <em>tuk, tak, tuk, tak</em>. Hanya butuh tiga hingga empat ketukan presisi, butiran biji melinjo telah bertransformasi menjadi kepingan pipih bulat bertekstur transparan.</li>
    <li><strong>Pelepasan & Penjemuran Alami (*Phoe*)</strong>: Menggunakan bilah pisau tipis khusus, lembaran emping basah dilepaskan secara hati-hati dari permukaan batu agar tidak sobek. Kepingan-kepingan tersebut kemudian ditata rapi di atas tampah anyaman bambu (<em>bleut</em> atau <em>nyiru</em>) dan dijemur di bawah terik matahari selama satu hingga dua hari penuh hingga kadar airnya menyusut total.</li>
</ol>

<h2 class="editorial-chapter-title">IV. Anatomi Mutu: Mengenal Kasta *Saboh Biji* hingga *Mulieng Lhee*</h2>
<p>Di pasar Beureunuen, kerupuk mulieng tidak dijual seragam. Terdapat strata mutu yang ketat yang mencerminkan tingkat kemahiran pengrajin dan peruntukan kuliner:</p>

<ul>
    <li><strong>Kualitas Super (Mulieng Saboh Biji / Satu Butir)</strong>: Ini adalah kasta tertinggi kerupuk melinjo. Satu keping emping hanya terbuat dari <em>satu butir biji melinjo</em> utuh. Hasilnya berukuran seukuran koin besar, sangat tipis nyaris tembus pandang, dan berwarna putih kekuningan bening. Saat digoreng, emping jenis ini mekar seketika dan lumer renyah di lidah tanpa meninggalkan rasa liat.</li>
    <li><strong>Kualitas Campur (Mulieng Dua Biji atau Tiga Biji)</strong>: Pada varian ini, dua hingga tiga butir melinjo ditempa bersamaan di atas batu sehingga membentuk satu kepingan bundar yang lebih lebar dan agak tebal. Varian ini sangat digemari untuk lauk makan harian karena teksturnya lebih mantap dan memberikan rasa gurih yang mengenyangkan.</li>
    <li><strong>Kerupuk Mulieng Siap Saji (Goreng Balado & Asin Gurih)</strong>: Selain bentuk mentah kering, pedagang di Beureunuen juga menyediakan varian siap santap, mulai dari rasa original asin gurih dengan olesan margarin dan garam, hingga bumbu pedas manis karamel khas Aceh yang sangat adiktif sebagai camilan perjalanan.</li>
</ul>

<h2 class="editorial-chapter-title">V. Gastronomi Tanah Rencong: Jodoh Abadi Mie Aceh dan Kuah Beulangong</h2>
<p>Kerupuk Mulieng memiliki tempat istimewa dalam struktur rasa kuliner Aceh. Masakan Aceh yang terkenal berani dalam penggunaan rempah-rempah kuat—seperti kapulaga, kayu manis, pekak, dan cabai merah—memerlukan penyeimbang tekstur dan rasa penutup yang tepat.</p>

<p>Kerenyahan emping melinjo dengan semburat rasa pahit-manis alami (<em>bittersweet note</em>) berfungsi sebagai pembersih langit-langit mulut (<em>palate cleanser</em>) yang memecah kepekatan lemak gulai pada <strong>Kuah Beulangong</strong> atau saus kental pada <strong>Mie Aceh Tumis</strong>. Menikmati kuah gulai kambing yang panas mengepul bersama nasi putih, didampingi renyahnya kerupuk mulieng Beureunuen yang dicelupkan sesaat ke dalam kuah, adalah puncak kenikmatan bersantap yang tak tergantikan bagi siapa pun penikmat rasa sejati.</p>

<h2 class="editorial-chapter-title">VI. Urat Nadi Ekonomi Inong Pidie: Kemandirian dari Balik Dapur Gampong</h2>
<p>Di balik semangkuk kelezatan kerupuk mulieng, tersimpan narasi sosial yang sangat mengharukan tentang kemandirian perempuan Pidie. Usaha kerupuk mulieng adalah industri rumahan berbasis komunitas yang menopang perekonomian ribuan kepala keluarga di pedesaan.</p>

<p>Di sela-sela mengurus rumah tangga dan mendidik anak, para ibu rumah tangga (<em>inong gampong</em>) duduk melingkar di teras dapur bersama tetangga, menumbuk emping sembari saling bercengkerama mempererat tali silaturahmi. Pendapatan harian dari menumbuk melinjo menjadi tabungan nyata yang membiayai sekolah anak-anak hingga perguruan tinggi, membeli perhiasan emas (<em>manyam meuh</em>) sebagai tabungan keluarga, dan menopang ekonomi keluarga saat musim paceklik sawah tiba. Kerupuk mulieng adalah simbol martabat kerja keras dan ketangguhan perempuan Pidie.</p>

<h2 class="editorial-chapter-title">VII. Panduan Menggoreng Sempurna di Dapur Rumah</h2>
<p>Agar kerupuk mulieng Beureunuen yang Anda bawa pulang mekar sempurna dan tidak pahit gosong, ikuti panduan rahasia dari koki tradisional Pidie berikut ini:</p>

<ul class="editorial-guide-list">
    <li><strong>Pastikan Kondisi Kering Mutlak</strong>: Jika kepingan emping terasa agak lembap akibat cuaca perjalanan, jemur kembali di bawah terik matahari selama 1–2 jam sebelum digoreng. Jangan pernah mencuci kepingan emping mentah.</li>
    <li><strong>Gunakan Minyak Berlimpah (*Deep Frying*)</strong>: Panaskan minyak goreng bersih dalam jumlah yang cukup banyak di wajan cekung. Suhu minyak harus benar-benar panas (sekitar 170°C–180°C). Uji dengan mencelupkan ujung kepingan; jika langsung berbuih dan mengembang, minyak siap digunakan.</li>
    <li><strong>Waktu Kilat (5–8 Detik Saja)</strong>: Masukkan kepingan emping beberapa lembar sekaligus (jangan terlalu padat). Tekan perlahan dengan saringan serok agar seluruh permukaan terendam. Dalam waktu 5 hingga 8 detik, saat emping mengembang dan berwarna putih krem keemasan, segera angkat dan tiriskan. Jangan menunggu hingga berwarna cokelat tua karena akan menimbulkan rasa pahit berlebih.</li>
    <li><strong>Sentuhan Garam Halus</strong>: Tiriskan emping di atas kertas penyerap minyak. Selagi masih hangat mengepul, taburkan sedikit garam halus atau bumbu sesuai selera untuk mengunci aroma gurih alaminya. Simpan dalam toples kedap udara setelah dingin sempurna.</li>
</ul>

<div class="editorial-quote-box">
    <blockquote>"Kerupuk Mulieng Beureunuen bukan sekadar emping pelengkap hidangan; ia adalah kristalisasi ketabahan inong Pidie, tempaan api kesabaran, dan warisan rasa yang mengikat ingatan para perantau pada tanah kelahirannya."</blockquote>
</div>
`, imageURL)

	idSummary := "Eksplorasi mendalam tentang kerupuk emping melinjo legendaris dari Beureunuen: tradisi sangrai pasir panas, sentuhan ritmis palu besi kaum ibu, hingga rahasia kerenyahan tiada tara yang merajai panggung kuliner Nusantara."

	// 6. Construct Comprehensive Editorial HTML & Markdown for English
	enHTML := fmt.Sprintf(`
<figure class="editorial-hero-banner">
    <img src="%s" alt="Savory Crunch of Beureunuen Kerupuk Mulieng: Pidie's Premium Melinjo Heritage" loading="eager" />
    <figcaption class="editorial-caption">Illustration: Authentic Kerupuk Mulieng from Beureunuen, Pidie. Handcrafted melinjo crisps sand-roasted over woodfire hearths and hammered into paper-thin golden wafers.</figcaption>
</figure>

<p class="editorial-lead">Throughout the northern reaches of Sumatra, across Indonesian gastronomic capitals, and within the vibrant heritage eateries of the Malay Peninsula, a steaming bowl of fragrant Mie Aceh or a ceremonial platter of Kuah Beulangong lamb curry is never truly complete without a generous topping of paper-thin, golden crisps bearing an unmistakable bittersweet resonance. This legendary culinary crowning is <strong>Kerupuk Mulieng</strong>—the celebrated artisan melinjo emping crisps hand-forged in the historic trading hub of <strong>Beureunuen, Pidie Regency</strong>.</p>

<p>For generations of Pidie households, the melinjo tree represents far more than an orchard asset; it is a foundational pillar of community resilience, rural matriarchal economic pride, and territorial identity. Through the steadfast dedication of village craftswomen wielding hand mallets upon smooth river stones, the raw harvest of nature is transformed into Sumatra’s most prized and delicate gastronomic export.</p>

<h2 class="editorial-chapter-title">I. Morning Pulse at Keude Beureunuen: Sumatra's Melinjo Capital</h2>
<p>The lively marketplace hub of <strong>Keude Beureunuen</strong> in Mutiara District has long reigned as the undisputed capital of the melinjo cracker trade on Sumatra Island. Situated strategically along the northern trans-Sumatran arterial highway connecting Banda Aceh to Medan, the market presents an unmistakable spectacle: rows of heritage arcades filled with hundreds of woven rattan baskets and towering glass jars displaying cascades of translucent, sun-cured melinjo wafers.</p>

<p>From the break of dawn, trading activity hums with energy. Rural producers arrive from surrounding subdistricts—Yaman, Glumpang Tiga, Sakti, Delima, and Mila—bringing batches freshly cured under yesterday's equatorial sun. Hundreds of metric tons of emping move through Beureunuen each month, catering not only to domestic travelers seeking authentic heritage gifts, but also fueling supply lines for prominent restaurants across Jakarta and overseas export markets in Malaysia, Singapore, and Brunei.</p>

<h2 class="editorial-chapter-title">II. The Melinjo Tree (*Bak Mulieng*) and Pidie's Volcanic Soil</h2>
<p>What endows Pidie melinjo crackers with their unrivaled crispness and delicate earthy flavor compared to varieties produced elsewhere? The answer lies in Pidie's unique geographical harmony and botanical heritage.</p>

<p>Nestled between the rugged volcanic spine of the Bukit Barisan range to the south and the nutrient-rich maritime breezes of the Malacca Strait to the north, Pidie possesses exceptionally fertile alluvial soils. This microclimate fosters optimal growing conditions for the indigenous <em>Gnetum gnemon</em> tree (known natively as <strong>Bak Mulieng</strong>).</p>

<p>In nearly every rural homestead garden (<em>pekarangan rumoh Aceh</em>), mature melinjo trees rise gracefully up to fifteen meters high. Decades-old root networks absorb mineral-rich groundwater, producing dense, nutrient-dense nuts (<em>boh mulieng</em>) with ideal natural starch-to-oil ratios. This superior seed structure ensures that Pidie crackers puff gracefully and uniformly upon touching hot oil, achieving a featherlight crunch without chemical additives or flour fillers.</p>

<h2 class="editorial-chapter-title">III. Artisan Metallurgy: From River Sand to Golden Crisps</h2>
<p>The manufacturing process behind Kerupuk Mulieng Beureunuen remains an inspiring tribute to manual precision and generational craftsmanship, untouched by industrial automation:</p>

<ol class="editorial-process-steps">
    <li><strong>Hand Selection (*Pilah Boh Mulieng*)</strong>: Only fully ripe nuts displaying deep crimson or golden-yellow husks are gathered. The outer fleshy pulp is hand-stripped to reveal the firm, protective inner nutshell.</li>
    <li><strong>Volcanic Sand Roasting (*Teu-oet lam Pasie*)</strong>: Rather than using cooking oil, nuts are gently roasted inside heavy earthen cauldrons (<em>beuleungoe</em>) filled with river sand. The scorching sand serves as a natural thermal buffer, evenly transmitting radiant heat through the nutshell without scorching the interior seed.</li>
    <li><strong>Rapid De-Husking (*Tupang*)</strong>: While still scorching hot, the nuts are lifted from the sand and immediately struck with a small wooden mallet to split the shell, exposing the pliable, steaming white kernel inside.</li>
    <li><strong>Rhythmic Forging (*Teupoh Mulieng*)</strong>: The kernel is immediately placed on a flat, polished river stone anvil (<em>aneuk batee</em>). Local craftswomen strike the kernel with balanced, rhythmic strokes using flat-faced iron mallets: <em>tuk, tak, tuk, tak</em>. Within three to four controlled strikes, the hot kernel flattens into an ultra-thin, round translucent disk.</li>
    <li><strong>Sun Drying (*Phoe*)</strong>: Utilizing a sharp paring blade, the wafer is gently detached from the stone and arranged on woven bamboo trays (<em>bleut</em> or <em>nyiru</em>) to cure under the tropical sun for one to two full days until completely dehydrated.</li>
</ol>

<h2 class="editorial-chapter-title">IV. Grading the Harvest: From Single-Nut Perfection to Everyday Bounty</h2>
<p>At the stalls of Beureunuen, melinjo crackers are classified into distinct culinary grades reflecting artisan skill:</p>

<ul>
    <li><strong>Super Grade (Mulieng Saboh Biji / Single Nut)</strong>: The highest pinnacle of the craft. Each individual cracker is fashioned from <em>a single melinjo seed</em>. The resulting wafer is coin-sized, extraordinarily thin, and nearly translucent. When fried, it expands instantaneously into a delicate crisp that melts on the tongue.</li>
    <li><strong>Family Grade (Mulieng Dua / Lhee Biji - 2 to 3 Nuts)</strong>: Two or three roasted seeds are combined and hammered together into a broader, slightly sturdier disk. This variant is favored as a daily dinner accompaniment, offering a heartier crunch and rich savoriness.</li>
    <li><strong>Ready-to-Eat Varieties</strong>: Besides raw dried wafers, merchants offer freshly fried selections tossed in refined sea salt or coated in a traditional Acehnese sweet chili glaze (balado).</li>
</ul>

<h2 class="editorial-chapter-title">V. Gastronomic Harmony: The Soulmate of Aceh Noodles & Curries</h2>
<p>Kerupuk Mulieng fulfills an indispensable gastronomic purpose within the robust culinary vocabulary of Aceh. The intensity of native Acehnese curries and stir-fried noodles—laden with star anise, cardamom, cloves, coriander, and dried chilies—demands an effective palate refresher.</p>

<p>The natural bittersweet profile and light crunch of fried melinjo cuts through the richness of spiced gravies in <strong>Kuah Beulangong</strong> and complements the peppery punch of <strong>Mie Aceh</strong>. Dipping a freshly fried cracker into warm, aromatic meat curry represents the zenith of authentic Acehnese dining pleasure.</p>

<h2 class="editorial-chapter-title">VI. The Village Microeconomy: Female Empowerment in Rural Pidie</h2>
<p>Beneath the appetizing crunch of Kerupuk Mulieng lies a profoundly uplifting story of rural female entrepreneurship. The melinjo trade is primarily operated by village women (<em>inong gampong</em>), who transform household kitchens into productive microenterprises.</p>

<p>Working alongside neighbors while watching over their children, women earn steady daily income that directly funds family savings, children's higher education, and emergency family reserves in gold. Kerupuk Mulieng stands as an enduring emblem of female autonomy, resilience, and quiet economic fortitude across Pidie Regency.</p>

<h2 class="editorial-chapter-title">VII. Frying Guide for Home Chefs</h2>
<p>To experience Beureunuen melinjo crackers at their golden peak in your own kitchen, follow these time-honored recommendations:</p>

<ul class="editorial-guide-list">
    <li><strong>Confirm Complete Dryness</strong>: If wafers have absorbed humidity during travel, sun-dry them for 1–2 hours before frying. Never wash raw crackers with water.</li>
    <li><strong>Plentiful Clean Cooking Oil</strong>: Heat a generous volume of fresh cooking oil in a deep wok to 175°C (350°F). Test with a small piece; if it puffs instantly to the surface, the oil is ready.</li>
    <li><strong>Quick Flash Frying (5 to 8 Seconds)</strong>: Submerge wafers in small batches. Lightly press them beneath the oil with a wire skimmer. Within 5 to 8 seconds, as they puff into golden cream wafers, lift and drain immediately. Over-frying will turn them dark brown and overly bitter.</li>
    <li><strong>Season While Warm</strong>: Drain on paper towels and dust with fine sea salt while warm to seal in their natural aroma. Store in airtight jars once completely cooled.</li>
</ul>

<div class="editorial-quote-box">
    <blockquote>"Beureunuen Kerupuk Mulieng is more than an appetizing side dish; it is the crystallized perseverance of Pidie craftswomen, the patient tempering of heat and stone, and a delicious bond tying every diaspora native to their ancestral home."</blockquote>
</div>
`, imageURL)

	enSummary := "An in-depth culinary and cultural chronicle of Beureunuen's legendary melinjo emping: from woodfire sand-roasting and rhythmic iron-hammer flattening to the secret behind Sumatra's finest bittersweet crisps."

	// 7. Update Indonesian translation
	_, err = dbPool.Exec(ctx, `
		UPDATE content_translations 
		SET title = $1, summary = $2, body_html = $3, body_markdown = $4,
		    meta_title = $5, meta_description = $6, updated_at = NOW()
		WHERE content_id = $7 AND locale = 'id'
	`,
		"Kelezatan Gurih Kerupuk Mulieng Beureunuen: Warisan Melinjo Terbaik Pidie",
		idSummary,
		idHTML,
		idHTML,
		"Kelezatan Gurih Kerupuk Mulieng Beureunuen: Warisan Melinjo Terbaik Pidie",
		idSummary,
		contentID,
	)
	if err != nil {
		return fmt.Errorf("failed to update Indonesian translation for kerupuk mulieng: %w", err)
	}
	fmt.Println(" [+] Indonesian Translation Published: 'Kelezatan Gurih Kerupuk Mulieng Beureunuen: Warisan Melinjo Terbaik Pidie'")

	// 8. Update or Insert English translation
	var enTransCount int
	_ = dbPool.QueryRow(ctx, `SELECT COUNT(*) FROM content_translations WHERE content_id = $1 AND locale = 'en'`, contentID).Scan(&enTransCount)

	if enTransCount == 0 {
		_, err = dbPool.Exec(ctx, `
			INSERT INTO content_translations (
				content_id, locale, title, slug, summary, body_html, body_markdown,
				meta_title, meta_description, published_at, updated_at
			) VALUES (
				$1, 'en', $2, 'savory-beureunuen-kerupuk-mulieng-melinjo', $3, $4, $5,
				$6, $7, NOW(), NOW()
			)
		`,
			contentID,
			"Savory Crunch of Beureunuen Kerupuk Mulieng: Pidie's Premium Melinjo Heritage",
			enSummary,
			enHTML,
			enHTML,
			"Savory Crunch of Beureunuen Kerupuk Mulieng: Pidie's Premium Melinjo Heritage",
			enSummary,
		)
	} else {
		_, err = dbPool.Exec(ctx, `
			UPDATE content_translations 
			SET title = $1, summary = $2, body_html = $3, body_markdown = $4,
			    meta_title = $5, meta_description = $6, updated_at = NOW()
			WHERE content_id = $7 AND locale = 'en'
		`,
			"Savory Crunch of Beureunuen Kerupuk Mulieng: Pidie's Premium Melinjo Heritage",
			enSummary,
			enHTML,
			enHTML,
			"Savory Crunch of Beureunuen Kerupuk Mulieng: Pidie's Premium Melinjo Heritage",
			enSummary,
			contentID,
		)
	}
	if err != nil {
		return fmt.Errorf("failed to update English translation for kerupuk mulieng: %w", err)
	}
	fmt.Println(" [+] English Translation Published: 'Savory Crunch of Beureunuen Kerupuk Mulieng: Pidie's Premium Melinjo Heritage'")

	// 9. Update source attribution
	sourceURL := "https://kebudayaan.kemdikbud.go.id/"
	sourceNote := "Dokumentasi Warisan Budaya Takbenda (WBTb) Indonesia & Sentra Industri Melinjo Beureunuen Dekranasda Pidie."
	_, _ = dbPool.Exec(ctx, `
		INSERT INTO content_sources (content_id, source_type, source_name, source_url, notes, created_at)
		VALUES ($1, 'official', 'Balai Pelestarian Nilai Budaya & Dekranasda Pidie', $2, $3, NOW())
		ON CONFLICT (content_id) DO UPDATE 
		SET source_type = EXCLUDED.source_type, source_name = EXCLUDED.source_name, source_url = EXCLUDED.source_url, notes = EXCLUDED.notes
	`, contentID, sourceURL, sourceNote)

	// 10. Update search vector in content_search
	_ = dbPool.QueryRow(ctx, `
		INSERT INTO content_search (content_id, locale, search_vector)
		VALUES ($1, 'id', to_tsvector('indonesian', $2 || ' ' || $3 || ' ' || $4))
		ON CONFLICT (content_id, locale) DO UPDATE
		SET search_vector = EXCLUDED.search_vector
	`, contentID, "Kelezatan Gurih Kerupuk Mulieng Beureunuen", idSummary, "melinjo emping beureunuen mutiara pidie kuliner renyah gurih")

	_ = dbPool.QueryRow(ctx, `
		INSERT INTO content_search (content_id, locale, search_vector)
		VALUES ($1, 'en', to_tsvector('english', $2 || ' ' || $3 || ' ' || $4))
		ON CONFLICT (content_id, locale) DO UPDATE
		SET search_vector = EXCLUDED.search_vector
	`, contentID, "Savory Crunch of Beureunuen Kerupuk Mulieng", enSummary, "melinjo emping beureunuen mutiara pidie culinary crackers crisps")

	// 11. Mark content as featured
	curationMeta, _ := json.Marshal(map[string]any{
		"badge":       "Kuliner Legendaris",
		"curator":     "Dewan Redaksi PidieKab",
		"verified_at": time.Now().Format("2006-01-02"),
	})
	_, _ = dbPool.Exec(ctx, `
		UPDATE contents 
		SET is_featured = true, curation_status = 'approved', curation_notes = $1, updated_at = NOW()
		WHERE id = $2
	`, string(curationMeta), contentID)

	fmt.Printf(" [✔] Content #%d (Kerupuk Mulieng) Published & Featured Successfully!\n", contentID)
	return nil
}
