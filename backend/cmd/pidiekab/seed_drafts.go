package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"pidiekab/internal/platform/config"
	"pidiekab/internal/platform/database"
)

type draftContentDef struct {
	catKey       string
	districtName string
	lat          float64
	lng          float64
	accuracy     string
	address      string
	idTitle      string
	idSlug       string
	idSummary    string
	idBody       string
	enTitle      string
	enSlug       string
	enSummary    string
	enBody       string
	sourceType   string
	sourceName   string
	sourceURL    string
	sourceNote   string
}

func handleSeedDraftsCLI(cfg config.Config, logger *slog.Logger) {
	fmt.Println("--> Seeding authentic draft contents for Kabupaten Pidie...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbPool, err := database.NewPool(ctx, database.Config{
		URL: cfg.DatabaseURL,
	}, logger)
	if err != nil {
		logger.Error("failed to connect to database for seeding drafts", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// 1. Get default admin ID
	var adminID int64
	err = dbPool.QueryRow(ctx, "SELECT id FROM admins WHERE email = 'admin@pidiekab.web.id' LIMIT 1").Scan(&adminID)
	if err != nil {
		// Fallback to any active admin
		_ = dbPool.QueryRow(ctx, "SELECT id FROM admins WHERE is_active = true ORDER BY id ASC LIMIT 1").Scan(&adminID)
	}
	if adminID == 0 {
		fmt.Println("Error: No admin account found to author drafts.")
		os.Exit(1)
	}

	// 2. Fetch category map
	catRows, err := dbPool.Query(ctx, "SELECT category_id, slug FROM category_translations WHERE locale = 'id'")
	if err != nil {
		logger.Error("failed to query categories", "error", err)
		os.Exit(1)
	}
	defer catRows.Close()

	catMap := make(map[string]int64)
	for catRows.Next() {
		var id int64
		var slug string
		if err := catRows.Scan(&id, &slug); err == nil {
			catMap[slug] = id
			if slug == "budaya-sejarah" {
				catMap["sejarah-budaya"] = id
			}
			if slug == "potensi-daerah" {
				catMap["potensi-daerah-ekonomi"] = id
			}
		}
	}

	// 3. Fetch district map
	distRows, err := dbPool.Query(ctx, "SELECT id, name FROM districts")
	if err != nil {
		logger.Error("failed to query districts", "error", err)
		os.Exit(1)
	}
	defer distRows.Close()

	districtMap := make(map[string]int64)
	for distRows.Next() {
		var id int64
		var name string
		if err := distRows.Scan(&id, &name); err == nil {
			districtMap[strings.ToLower(strings.TrimSpace(name))] = id
		}
	}

	// 4. Draft contents inventory
	drafts := []draftContentDef{
		{
			catKey:       "wisata-alam",
			districtName: "Padang Tiji",
			lat:          5.3782,
			lng:          95.8451,
			accuracy:     "approximate",
			address:      "Pedalaman Hutan Pulo Hagu, Kaki Gunung Seulawah Dara, Kecamatan Padang Tiji, Kabupaten Pidie, Aceh",
			idTitle:      "Keajaiban Ngarai Alami Lingkok Kuwieng: Grand Canyon Tersembunyi di Lembah Padang Tiji",
			idSlug:       "keajaiban-ngarai-alami-lingkok-kuwieng-padang-tiji",
			idSummary:    "Ngarai purba berdinding batu berukir alami hasil erosi air sungai berabad-abad, dikenal warga lokal sebagai Uruek Meuh di pedalaman Padang Tiji.",
			idBody:       "Lingkok Kuwieng, atau yang sering disebut sebagai Grand Canyon-nya Aceh dan Uruek Meuh (Lubang Emas), merupakan keajaiban geomorfologi tersembunyi di kaki Gunung Seulawah Dara, Kecamatan Padang Tiji. Formasi bebatuan di sepanjang dinding ngarai terbentang bertingkat-tingkat menyerupai pahatan istana kuno, terbentuk murni dari kikisan aliran sungai pegunungan selama ratusan tahun. Air sungai yang mengalir di dasarnya sangat jernih berwarna kehijauan dengan kolam-kolam alami yang menyegarkan. Menuju lokasi ini membutuhkan perjalanan lintas kebun dan hutan tropis yang menantang, menjadikannya surga bagi para pecinta alam petualang, fotografer lanskap, dan penikmat ketenangan belantara Aceh.",
			enTitle:      "The Geological Marvel of Lingkok Kuwieng: Hidden Grand Canyon in Padang Tiji Valley",
			enSlug:       "geological-marvel-lingkok-kuwieng-padang-tiji-valley",
			enSummary:    "An ancient river gorge featuring naturally sculpted sandstone cliffs and emerald pools, revered by locals as Uruek Meuh.",
			enBody:       "Lingkok Kuwieng, celebrated as Aceh's Grand Canyon and traditionally known as Uruek Meuh (The Golden Hole), is a pristine geological wonder nestled at the foothills of Mount Seulawah Dara in Padang Tiji District. The gorge walls exhibit rhythmic layered limestone and sandstone formations sculpted by mountain torrents over centuries, resembling classical architectural terraces. Crystal-clear emerald waters flow along the canyon floor, creating natural dipping pools and gentle cascades. Reaching this hidden sanctuary involves a rugged off-road trek across remote forest canopies, rewarding outdoor enthusiasts with one of Sumatra's most awe-inspiring untamed landscapes.",
			sourceType:   "official",
			sourceName:   "Dinas Kebudayaan dan Pariwisata Aceh",
			sourceURL:    "https://acehtourism.travel/",
			sourceNote:   "Dokumentasi Destinasi Geowisata Unggulan Pedalaman Pidie.",
		},
		{
			catKey:       "wisata-alam",
			districtName: "Tangse",
			lat:          5.0934,
			lng:          95.9612,
			accuracy:     "accurate",
			address:      "Jalan Lintas Beureunuen–Geumpang, Gampong Meuriya, Kecamatan Tangse, Kabupaten Pidie, Aceh",
			idTitle:      "Pesona Gemericik Air Terjun Tangse di Jalur Hijau Pegunungan Bukit Barisan",
			idSlug:       "pesona-gemericik-air-terjun-tangse-jalur-pegunungan",
			idSummary:    "Air terjun alami berhawa sejuk yang berada tepat di sisi jalur lintas dataran tinggi Pidie, menjadi persinggahan favorit pelepas lelah.",
			idBody:       "Air Terjun Tangse di Gampong Meuriya menyajikan keindahan alam yang unik karena posisinya yang berbatasan langsung dengan badan jalan lintas Beureunuen–Geumpang. Mengalir deras dari lereng pegunungan Bukit Barisan, airnya bersuhu sangat dingin dan murni tanpa polusi. Pengendara yang melintasi dataran tinggi Tangse kerap menyempatkan diri singgah untuk membasuh muka, berfoto, dan menikmati kesejukan kabut air. Di seberang air terjun, berjejer kedai kopi lokal yang menyajikan kopi Robusta hangat sembari memandang lembah persawahan terasering yang membentang asri di bawah rindangnya perbukitan.",
			enTitle:      "The Splendor of Tangse Waterfall along the Highland Mountain Route",
			enSlug:       "splendor-tangse-waterfall-highland-mountain-route",
			enSummary:    "A refreshing roadside natural cascade along the mountain pass of Tangse, famous for crisp alpine air and local coffee stalls.",
			enBody:       "Tangse Waterfall, situated in Meuriya Village, offers a picturesque encounter where pristine mountain waters cascade right alongside the scenic Beureunuen–Geumpang highway. Flowing from the untouched slopes of the Bukit Barisan range, its crystalline alpine stream provides an invigorating stopover for regional travelers. Facing the waterfall, rustic roadside wooden pavilions serve freshly brewed single-origin Tangse Robusta coffee alongside panoramic views of emerald terraced paddy fields framed by forested highland ridges.",
			sourceType:   "community",
			sourceName:   "Komunitas Penjelajah Alam Pidie",
			sourceURL:    "https://disbudpar.acehprov.go.id/",
			sourceNote:   "Kompilasi panduan titik rekreasi alam koridor wisata Tangse-Geumpang.",
		},
		{
			catKey:       "wisata-alam",
			districtName: "Muara Tiga",
			lat:          5.4831,
			lng:          95.8715,
			accuracy:     "accurate",
			address:      "Pesisir Gampong Laweung, Kecamatan Muara Tiga, Kabupaten Pidie, Aceh",
			idTitle:      "Pantai Ujong Pi Laweung: Lanskap Karst dan Deburan Ombak Selat Malaka",
			idSlug:       "pantai-ujong-pi-laweung-lanskap-karst-ombak",
			idSummary:    "Pantai pesisir berkarang eksotis di ujung barat Pidie dengan deretan perahu nelayan tradisional dan pemandangan laut lepas.",
			idBody:       "Pantai Ujong Pi di kawasan Laweung, Muara Tiga, menawarkan daya tarik pesisir yang berbeda dari pantai pasir pada umumnya. Berada di dekat gugusan perbukitan karst, pantai ini dihiasi tebing-tebing karang kokoh yang memecah ombak Selat Malaka. Pagi hari di Ujong Pi diwarnai oleh aktivitas para nelayan tradisional mendaratkan hasil tangkapan laut segar, mulai dari ikan tongkol, kakap, hingga cumi. Sore hari menyuguhkan siluet matahari terbenam yang memukau di cakrawala laut utara Aceh, dipadukan dengan hembusan angin pantai yang menenangkan.",
			enTitle:      "Ujong Pi Beach Laweung: Rugged Karst Coastline on the Malacca Strait",
			enSlug:       "ujong-pi-beach-laweung-rugged-karst-coastline",
			enSummary:    "A dramatic coastal destination featuring sea-facing limestone cliffs, traditional fishing harbors, and open horizon vistas.",
			enBody:       "Ujong Pi Beach in Laweung, Muara Tiga District, presents a dramatic coastal contrast to standard sandy beaches. Framed by surrounding karst cliffs, the shoreline meets the open swells of the Malacca Strait with majestic rocky breakwaters. Visitors can witness traditional artisanal fishermen bringing ashore daily catches of tuna and mackerel in vibrant motorized skiffs, while evening horizons offer sweeping sunset panoramas over northern Sumatra's coastal frontier.",
			sourceType:   "official",
			sourceName:   "Dinas Pariwisata dan Pemuda Olahraga Kabupaten Pidie",
			sourceURL:    "https://pidiekab.web.id/",
			sourceNote:   "Riset pemetaan potensi bahari pesisir utara Kabupaten Pidie.",
		},
		{
			catKey:       "wisata-alam",
			districtName: "Geumpang",
			lat:          4.8872,
			lng:          96.1124,
			accuracy:     "approximate",
			address:      "Aliran Hulu Krueng Geumpang, Kecamatan Geumpang, Kabupaten Pidie, Aceh",
			idTitle:      "Petualangan Arung Jeram Sungai Geumpang: Menembus Belantara Tropis Pidie",
			idSlug:       "petualangan-arung-jeram-sungai-geumpang-belantara-tropis",
			idSummary:    "Destinasi wisata petualangan arung jeram kelas dunia dengan grade jeram menantang di tengah hutan hujan tropis Sumatra.",
			idBody:       "Krueng Geumpang yang berhulu di hutan belantara pegunungan Leuser utara menyimpan potensi arung jeram (white water rafting) terbaik di Aceh. Sungai berarus deras dengan batu-batu granit raksasa ini menawarkan jeram bervariasi dari grade III hingga IV, menjadikannya wahana pemacu adrenalin bagi para pegiat olahraga arus deras. Sepanjang rute pengarungan, peserta disuguhi kanopi hutan hujan primer yang lebat, suara kicauan burung enggang, dan tebing-tebing hijau nan alami yang belum tersentuh hiruk-pikuk perkotaan.",
			enTitle:      "Geumpang River Whitewater Rafting Adventure: Traversing Tropical Wilderness",
			enSlug:       "geumpang-river-whitewater-rafting-tropical-wilderness",
			enSummary:    "A premier ecotourism river rafting corridor with challenging Grade III-IV rapids flowing through pristine rainforest valleys.",
			enBody:       "The upper reaches of Krueng Geumpang, originating from the highland rainforests of central Sumatra, present one of Aceh's finest whitewater rafting corridors. Bounded by giant granite boulders and exhilarating hydraulic chutes ranging from Grade III to Grade IV, this river expedition offers adrenaline rushes amid untouched tropical landscapes. Rafters paddle beneath canopy corridors inhabited by hornbills and diverse endemic wildlife, discovering the untamed ecological heartland of Pidie.",
			sourceType:   "community",
			sourceName:   "Federasi Arung Jeram Indonesia (FAJI) Pengda Aceh",
			sourceURL:    "https://faji.org/",
			sourceNote:   "Survei kelayakan jalur arung jeram ekowisata perairan pedalaman Pidie.",
		},
		{
			catKey:       "kuliner-khas",
			districtName: "Kembang Tanjong",
			lat:          5.3124,
			lng:          96.0125,
			accuracy:     "accurate",
			address:      "Kawasan Sentra Penganan Tradisional, Kecamatan Kembang Tanjong, Kabupaten Pidie, Aceh",
			idTitle:      "Kue Adee Kembang Tanjong: Gurih Legit Bertabur Bawang Goreng Khas Tanah Rencong",
			idSlug:       "kue-adee-kembang-tanjong-gurih-legit-bawang-goreng",
			idSummary:    "Kue tradisional legendaris berbahan singkong atau tepung terigu dengan perpaduan santan kental, gula, dan taburan bawang goreng aromatik.",
			idBody:       "Kue Adee adalah mahakarya kuliner tradisional dari Kecamatan Kembang Tanjong yang telah tersohor ke seantero Aceh dan luar daerah. Kue panggang ini memiliki dua varian utama: Adee Ubi (berbahan singkong parut halus) dan Adee Terigu. Keunikan kue adee terletak pada teksturnya yang kenyal, lembut di dalam, dan sedikit garing di bagian permukaannya. Cita rasa manis gurih santan kelapa diperkaya oleh aroma khas taburan bawang merah goreng renyah dan sedikit biji jintan di bagian atasnya, menjadikannya kudapan wajib saat berbuka puasa, acara pesta adat, dan buah tangan favorit pelancong.",
			enTitle:      "Adee Cake of Kembang Tanjong: Sweet Spiced Heritage Confection of Pidie",
			enSlug:       "adee-cake-kembang-tanjong-sweet-spiced-confection",
			enSummary:    "A celebrated baked confection made from cassava or flour, infused with coconut milk and topped with aromatic crispy shallots.",
			enBody:       "Kue Adee represents one of Pidie's most famous culinary exports, rooted in the baking traditions of Kembang Tanjong District. Available in both cassava (Adee Ubi) and flour variants, this baked cake features a golden caramelized crust with a moist, velvety interior. The rich coconut milk and palm sweetness is uniquely balanced by a savory garnish of crisp golden shallots and subtle anise seeds, creating an unforgettable sweet-savory profile cherished during religious festivals and wedding banquets across Aceh.",
			sourceType:   "media",
			sourceName:   "Kompas TV Program Kuliner Tradisional Nusantara",
			sourceURL:    "https://kompas.tv/",
			sourceNote:   "Liputan warisan kuliner otentik khas pesisir Kembang Tanjong Pidie.",
		},
		{
			catKey:       "kuliner-khas",
			districtName: "Peukan Baro",
			lat:          5.3341,
			lng:          95.9421,
			accuracy:     "accurate",
			address:      "Kemukiman Lampoh Saka, Kecamatan Peukan Baro, Kabupaten Pidie, Aceh",
			idTitle:      "Filosofi Tradisi Kenduri Apam Pidie: Warisan Budaya Tak Benda Bernilai Luhur",
			idSlug:       "filosofi-tradisi-kenduri-apam-pidie-warisan-budaya",
			idSummary:    "Kue apam tradisional bakar wajan tanah liat yang disajikan dengan kuah tuhe gurih manis dalam ritual sosial kemasyarakatan.",
			idBody:       "Tradisi 'Toet Apam' (membakar kue apam) dan Kenduri Apam merupakan warisan budaya tak benda resmi Indonesia yang berakar kuat dalam kebudayaan masyarakat Pidie sejak era Kesultanan. Apam dibuat dari adonan tepung beras dan santan yang dibakar di atas wajan tanah liat tradisional (beuleungoe) dengan bahan bakar daun kelapa kering atau arang. Apam yang harum ini kemudian disajikan bersama 'Kuah Tuhe'—kuah santan kental manis yang dimasak bersama gula aren, nangka masak, atau pisang raja. Tradisi ini bukan sekadar memasak, melainkan media silaturahmi akbar dan sedekah antarwarga di bulan Rajab.",
			enTitle:      "The Cultural Philosophy of Kenduri Apam Pidie: Intangible Cultural Heritage",
			enSlug:       "cultural-philosophy-kenduri-apam-pidie-heritage",
			enSummary:    "Ancient earthen-pan rice flour pancakes paired with sweetened jackfruit-coconut broth, certified as Indonesian Intangible Cultural Heritage.",
			enBody:       "The culinary tradition of 'Toet Apam' (baking apam pancakes) and the communal Kenduri Apam feast constitute a distinguished Intangible Cultural Heritage of Indonesia centered in Pidie. Prepared using fermented rice batter baked in handmade clay pans over dried palm fronds, these delicate round crumpets are accompanied by 'Kuah Tuhe'—a slow-simmered spiced coconut nectar enriched with ripe jackfruit slices and aromatic pandan. Celebrated annually during the Islamic month of Rajab, the ceremony embodies communal solidarity, mutual sharing, and intergenerational culinary transmission.",
			sourceType:   "official",
			sourceName:   "Kementerian Kebudayaan RI - Warisan Budaya Takbenda",
			sourceURL:    "https://warisanbudaya.kemdikbud.go.id/",
			sourceNote:   "Penetapan Karya Budaya Tradisi Kenduri Apam Pidie sebagai WBTB Nasional.",
		},
		{
			catKey:       "kuliner-khas",
			districtName: "Grong-Grong",
			lat:          5.3852,
			lng:          95.9124,
			accuracy:     "accurate",
			address:      "Jalan Raya Banda Aceh–Medan, Pasar Grong-Grong, Kecamatan Grong-Grong, Kabupaten Pidie, Aceh",
			idTitle:      "Mie Caluek Grong-Grong: Ikon Kuliner Pedas Gurih Bumbu Kacang Jalur Lintas",
			idSlug:       "mie-caluek-grong-grong-ikon-kuliner-pedas-gurih",
			idSummary:    "Mi lidi tradisional khas Pidie yang disajikan hangat dengan siraman saus bumbu kacang pedas berempah dan renyahnya kerupuk merah.",
			idBody:       "Bagi siapa saja yang melintasi jalur utama Banda Aceh–Medan di kawasan Grong-Grong, aroma bumbu Mie Caluek di deretan kedai pinggir jalan selalu menggugah selera. Mie Caluek dibuat dari mi lidi tebal yang direbus kenyal, lalu disiram saus kuah kacang kental bercita rasa pedas, gurih, dan sedikit asam segar. Sajian ini dilengkapi dengan tumisan bumbu cabai merah pedas, sayuran kangkung rebus atau kol, serta taburan kerupuk merah gurih. Dinikmati hangat dengan segelas teh tarik dingin atau es timun serut, Mie Caluek menjadi hidangan rakyat yang dicintai semua kalangan.",
			enTitle:      "Mie Caluek of Grong-Grong: The Iconic Savory Spiced Noodle of Trans-Sumatra Route",
			enSlug:       "mie-caluek-grong-grong-savory-spiced-noodles",
			enSummary:    "A beloved roadside delicacy featuring thick noodles doused in aromatic spiced peanut gravy, chili relish, and red crackers.",
			enBody:       "Along the bustling Trans-Sumatra Highway crossing Grong-Grong District, roadside stalls lure travelers with the fragrant aroma of Mie Caluek. This distinctive Pidie street food utilizes springy wheat noodles generously smothered in an intensely aromatic spiced peanut sauce, slow-simmered chili gravy, blanched local greens, and crisp tapioca crackers. Quick, flavorful, and deeply satisfying, Mie Caluek stands as an authentic testament to Pidie's roadside culinary artistry.",
			sourceType:   "media",
			sourceName:   "Portal Berita Budaya & Kuliner Aceh",
			sourceURL:    "https://portalsatu.com/",
			sourceNote:   "Liputan ragam kuliner kaki lima legendaris jalur lintas Banda Aceh-Medan.",
		},
		{
			catKey:       "kuliner-khas",
			districtName: "Muara Tiga",
			lat:          5.4712,
			lng:          95.8841,
			accuracy:     "accurate",
			address:      "Pelabuhan Ikan Laweung, Kecamatan Muara Tiga, Kabupaten Pidie, Aceh",
			idTitle:      "Mie Sure Laweung: Gurihnya Olahan Ikan Tongkol Segar Pesisir Selat Malaka",
			idSlug:       "mie-sure-laweung-gurihnya-ikan-tongkol-pesisir",
			idSummary:    "Varian Mie Aceh istimewa dengan potongan ikan tongkol (sure) segar hasil tangkapan nelayan lokal di pelabuhan Laweung.",
			idBody:       "Jika Mie Aceh pada umumnya dipadukan dengan daging sapi, kepiting, atau udang, maka di pesisir Laweung Anda akan menemukan primadona kuliner lokal bernama Mie Sure. 'Sure' adalah sebutan masyarakat Aceh untuk ikan tongkol. Ikan tongkol segar yang baru diangkat dari perahu nelayan dipotong tebal dan ditumis bersama racikan bumbu rempah kari Aceh yang kaya akan kapulaga, jintan, ketumbar, dan cabai merah. Daging ikan yang padat, gurih, dan manis alami meresap sempurna ke dalam kuah kental beraroma rempah, memberikan sensasi makan mi tepi laut yang tiada duanya.",
			enTitle:      "Mie Sure of Laweung: Fresh Tuna-Infused Acehnese Noodles by the Sea",
			enSlug:       "mie-sure-laweung-fresh-tuna-acehnese-noodles",
			enSummary:    "A specialty seafood noodle preparation highlighting freshly landed skipjack tuna simmered in complex aromatic coastal curries.",
			enBody:       "While Acehnese noodles frequently feature beef, mud crabs, or prawns, the coastal settlement of Laweung in Muara Tiga offers a localized specialty: Mie Sure. In the Acehnese language, 'sure' refers to freshly caught skipjack tuna. Sourced directly from local fishing catamarans, thick succulent tuna chunks are sautéed in small woks with fragrant Aceh curry spices, cardamom, fresh shallots, and fiery chilies. The firm, savory sweetness of the oceanic fish blends seamlessly into the concentrated gravy, making it an essential pilgrimage for seafood gourmands.",
			sourceType:   "media",
			sourceName:   "Harian Serambi Indonesia - Rubrik Kuliner",
			sourceURL:    "https://aceh.tribunnews.com/",
			sourceNote:   "Eksplorasi kuliner bahari pesisir utara Tanah Rencong.",
		},
		{
			catKey:       "kuliner-khas",
			districtName: "Mutiara",
			lat:          5.3215,
			lng:          95.9234,
			accuracy:     "accurate",
			address:      "Kemukiman Busu, Kecamatan Mutiara, Kabupaten Pidie, Aceh",
			idTitle:      "Lincah Busu: Sensasi Segar Rujak Buah Tradisional Kemukiman Busu Pidie",
			idSlug:       "lincah-busu-sensasi-segar-rujak-buah-tradisional",
			idSummary:    "Rujak buah khas Pidie dengan paduan ubi jalar, mangga muda, kedondong, dan kuah sambal rujak gula aren beraroma terasi bakar.",
			idBody:       "Lincah Busu adalah rujak tradisional kebanggaan warga Kemukiman Busu, Kecamatan Mutiara. Yang membedakan Lincah Busu dengan rujak daerah lain adalah komposisi buah-buahan lokal seperti rumbia muda (boh meuria), pisang batu, bengkuang, dan ubi mentah yang diulek kasar bersama gula aren pekat, cabai rawit, garam, dan sedikit terasi panggang khas pesisir. Perpaduan rasa manis, pedas, asam, dan sedikit kelat (sepat) dari buah-buahan tradisional menghadirkan kesegaran alami yang sangat digemari saat cuaca terik siang hari.",
			enTitle:      "Lincah Busu: The Authentic Traditional Fruit Salad of Busu Settlement",
			enSlug:       "lincah-busu-authentic-fruit-salad-busu-settlement",
			enSummary:    "A rustic fruit salad combining tropical roots and sour fruits with artisanal palm sugar, roasted shrimp paste, and fresh bird's eye chilies.",
			enBody:       "Lincah Busu represents a heritage fruit salad crafted exclusively in the historic Busu settlement of Mutiara District. Distinguished by its inclusion of heirloom ingredients such as young sago fruit (boh meuria), wild seed bananas, sweet potatoes, and green mangoes, the salad is tossed in a mortar with stone-ground palm sugar, sea salt, roasted shrimp paste, and fiery chilies. The interplay of sweet, sharp tartness, crunch, and subtle tannin notes creates a revitalizing tropical indulgence celebrating rural botanical abundance.",
			sourceType:   "community",
			sourceName:   "Pemerhati Kuliner Nusantara Cabang Aceh",
			sourceURL:    "https://portalsatu.com/",
			sourceNote:   "Kajian kearifan lokal kuliner buah tradisional masyarakat Kemukiman Busu.",
		},
		{
			catKey:       "kuliner-khas",
			districtName: "Tangse",
			lat:          5.0812,
			lng:          95.9741,
			accuracy:     "accurate",
			address:      "Perkebunan Kopi Dataran Tinggi Tangse, Kecamatan Tangse, Kabupaten Pidie, Aceh",
			idTitle:      "Kopi Robusta Tangse: Sensasi Aroma Rempah Alami dari Ketinggian Pegunungan Pidie",
			idSlug:       "kopi-robusta-tangse-sensasi-aroma-rempah-pegunungan",
			idSummary:    "Biji kopi Robusta pegunungan dengan bodi tebal dan aroma cokelat karamel rempah yang tumbuh subur di ketinggian 600-900 mdpl.",
			idBody:       "Selain Gayo yang tersohor dengan Arabikanya, Aceh memiliki sentra penghasil kopi Robusta unggulan di lembah pegunungan Tangse, Kabupaten Pidie. Ditanam secara organik oleh petani lokal di tanah vulkanik yang subur, Kopi Robusta Tangse memiliki karakteristik bodi yang sangat tebal (full body), tingkat keasaman rendah, serta aftertaste rasa cokelat pahit manis dengan sentuhan aroma rempah alami. Disangrai dengan metode tradisional kuali tanah liat atau mesin modern, kopi ini menjadi minuman wajib yang dinikmati masyarakat di ratusan warung kopi (warkop) dari pesisir hingga pegunungan Pidie.",
			enTitle:      "Tangse Robusta Coffee: Natural Herbal Notes from Highland Mountains of Pidie",
			enSlug:       "tangse-robusta-coffee-natural-herbal-notes-highlands",
			enSummary:    "Highland Robusta coffee beans cultivated on volcanic mountain soils, noted for bold body and deep chocolate-spiced undertones.",
			enBody:       "Complementing Gayo's famous Arabica varieties, Pidie's highland valley of Tangse yields some of Sumatra's most distinctive Robusta beans. Cultivated by smallholder farmers at elevations between 600 and 900 meters above sea level on rich volcanic slopes, Tangse Robusta boasts a full, syrupy body, low acidity, and lingering bittersweet cocoa notes interwoven with herbal overtones. Roasted using both artisanal earthen-kiln techniques and modern drum roasters, it powers the legendary coffee culture that thrives throughout Pidie.",
			sourceType:   "official",
			sourceName:   "Dinas Perkebunan dan Peternakan Provinsi Aceh",
			sourceURL:    "https://disbunnak.acehprov.go.id/",
			sourceNote:   "Statistik Komoditas Perkebunan Unggulan Daerah Kopi Rakyat Tangse.",
		},
		{
			catKey:       "sejarah-budaya",
			districtName: "Kota Sigli",
			lat:          5.3812,
			lng:          95.9621,
			accuracy:     "accurate",
			address:      "Kawasan Lampoih Lada, Kota Sigli, Kabupaten Pidie, Aceh",
			idTitle:      "Benteng Kuta Asan Sigli: Jejak Epik Benteng Pertahanan Melawan Invasi Kolonial",
			idSlug:       "benteng-kuta-asan-sigli-jejak-benteng-pertahanan",
			idSummary:    "Situs pertahanan maritim bersejarah yang menjadi saksi bisu kegigihan pejuang Pidie dalam menahan gempuran armada penjajah.",
			idBody:       "Benteng Kuta Asan di Kota Sigli merupakan salah satu situs pertahanan militer terpenting dalam sejarah perlawanan rakyat Aceh di kawasan pesisir Peudir. Dibangun dari batu bata kokoh dan karang laut, benteng ini dilengkapi menara pengintai dan dudukan meriam pertahanan pantai yang oleh masyarakat setempat dikenal dengan istilah 'Kupaleh'. Dalam catatan sejarah Perang Aceh abad ke-19, Kuta Asan menjadi kubu pertahanan sengit yang berkali-kali menahan serangan darat dan laut ekspedisi militer Belanda. Hari ini, kompleks bersejarah ini menjadi pengingat abadi akan keberanian dan patriotisme para leluhur Pidie.",
			enTitle:      "Kuta Asan Fort Sigli: Epic Bastion of Maritime Resistance Against Colonial Forces",
			enSlug:       "kuta-asan-fort-sigli-bastion-maritime-resistance",
			enSummary:    "A historic maritime coastal citadel that witnessed fierce defensive battles fought by Pidie warriors against 19th-century colonial incursions.",
			enBody:       "Kuta Asan Fort in Sigli stands as a foundational monument to Acehnese military resilience along the historic Peudir seaboard. Constructed from durable red clay bricks and marine coral masonry, the fortress incorporated gun emplacements and surveillance watchtowers known in local lore as 'Kupaleh'. During the 19th-century Aceh War, Kuta Asan formed a formidable frontline stronghold that withstood repeated naval bombardments and infantry assaults. Today, its heritage grounds serve as an enduring symbol of unyielding civic courage and historical sovereignty.",
			sourceType:   "community",
			sourceName:   "Masyarakat Peduli Sejarah Aceh (MAPESA)",
			sourceURL:    "https://mapesaaceh.com/",
			sourceNote:   "Dokumentasi riset arkeologi dan toponimi benteng pertahanan pesisir Pidie.",
		},
		{
			catKey:       "sejarah-budaya",
			districtName: "Tiro / Truseb",
			lat:          5.2341,
			lng:          95.9412,
			accuracy:     "accurate",
			address:      "Gampong Meunje, Kemukiman Truseb, Kecamatan Tiro, Kabupaten Pidie, Aceh",
			idTitle:      "Ziarah Sejarah Makam Teungku Chik Di Tiro: Mengenang Kobaran Semangat Perang Sabil",
			idSlug:       "ziarah-sejarah-makam-teungku-chik-di-tiro-pahlawan",
			idSummary:    "Situs ziarah cagar budaya makam Pahlawan Nasional Muhammad Saman (Teungku Chik Di Tiro), ulama kharismatik panglima Perang Sabil.",
			idBody:       "Di lembah perbukitan Kemukiman Truseb, Kecamatan Tiro, bersemayam jasad Pahlawan Nasional Indonesia, Teungku Chik Di Tiro (Muhammad Saman). Beliau adalah ulama besar, intelektual, dan panglima perang yang mengobarkan kembali perang semesta melawan penjajah Belanda melalui seruan 'Hikayat Prang Sabi'. Di bawah kepemimpinannya, benteng-benteng pertahanan kolonial berhasil direbut kembali oleh rakyat Aceh. Kompleks makam yang asri dan tenang ini selalu ramai dikunjungi para peziarah, pelajar, dan peneliti sejarah untuk mengenang jasa dan meneladani keteguhan prinsip sang pahlawan.",
			enTitle:      "Historical Pilgrimage to the Tomb of Teungku Chik Di Tiro: National Hero of Pidie",
			enSlug:       "pilgrimage-tomb-teungku-chik-di-tiro-national-hero",
			enSummary:    "The sacred resting sanctuary of National Hero Teungku Chik Di Tiro, the charismatic scholar who led the decisive resistance in Aceh.",
			enBody:       "Situated in the quiet hills of Truseb, Tiro District, rests National Hero Teungku Chik Di Tiro (Muhammad Saman). A revered Islamic jurist and military strategist, he galvanized regional resistance against colonial expansion in the late 19th century through the iconic patriotic poetry of Hikayat Prang Sabi. Under his decisive command, Acehnese forces liberated strategic fortresses and unified regional clans. His memorial sanctuary welcomes pilgrims, scholars, and historians seeking to honor a lifetime dedicated to justice, faith, and national dignity.",
			sourceType:   "official",
			sourceName:   "Kementerian Sosial RI - Direktori Pahlawan Nasional",
			sourceURL:    "https://kemensos.go.id/",
			sourceNote:   "Profil Pahlawan Nasional Teungku Chik Di Tiro (SK Presiden RI No. 087/TK/1973).",
		},
		{
			catKey:       "sejarah-budaya",
			districtName: "Pidie",
			lat:          5.3612,
			lng:          95.9341,
			accuracy:     "approximate",
			address:      "Kecamatan Pidie, Kabupaten Pidie, Aceh",
			idTitle:      "Seni Tari Seudati: Harmoni Hentakan Kaki dan Tepukan Dada Khas Kesultanan Pidie",
			idSlug:       "seni-tari-seudati-harmoni-hentakan-kaki-tepukan-dada",
			idSummary:    "Tarian tradisional heroik tanpa alat musik pengiring yang mengandalkan tepukan dada, petikan jari, dan lantunan syair bertema dakwah serta patriotisme.",
			idBody:       "Tari Seudati adalah salah satu mahakarya seni pertunjukan paling prestisius yang lahir dan mengakar kuat dalam kebudayaan masyarakat Kabupaten Pidie. Berbeda dengan tarian lain, Seudati tidak menggunakan alat musik tabuh sama sekali; seluruh ritme dan harmoni tercipta dari tubuh para penari melalui tepukan dada (keutip dada), tepukan pinggul, petikan jemari, dan hentakan kaki serempak yang bergemuruh di atas lantai panggung kayu. Dipimpin oleh seorang Syeikh dan didampingi lantunan vokal syair oleh Aneuk Syahi, Tari Seudati memancarkan semangat ksatria, kekompakan, dan nilai dakwah yang memukau penonton di festival-festival budaya nasional maupun internasional.",
			enTitle:      "The Art of Seudati Dance: Rhythmic Chest Percussion and Heroic Camaraderie",
			enSlug:       "art-seudati-dance-rhythmic-chest-percussion-pidie",
			enSummary:    "A dynamic, instrument-free traditional dance that generates powerful rhythmic harmonies through synchronized footwork and chest-slaps.",
			enBody:       "Tari Seudati ranks among the most captivating performance traditions originating from the cultural heartlands of Pidie. Uniquely acoustic, Seudati employs no instrumental backing; every beat, dynamic pulse, and crescendo is produced entirely by the dancers through synchronized chest-slapping, hip-strikes, finger-snapping, and thunderous footwork against elevated timber stages. Directed by a master choreographer (Syeikh) and accompanied by the melodious chanting of lyrical vocalists (Aneuk Syahi), Seudati embodies heroic camaraderie, spiritual devotion, and artistic virtuosity on world stages.",
			sourceType:   "official",
			sourceName:   "Balai Pelestarian Kebudayaan Wilayah I Aceh",
			sourceURL:    "https://kebudayaan.kemdikbud.go.id/bpk1/",
			sourceNote:   "Inventarisasi Seni Pertunjukan Tradisional Warisan Budaya Tak Benda Aceh.",
		},
		{
			catKey:       "potensi-daerah",
			districtName: "Delima",
			lat:          5.3524,
			lng:          95.8912,
			accuracy:     "accurate",
			address:      "Gampong Aree, Kecamatan Delima, Kabupaten Pidie, Aceh",
			idTitle:      "Seni Kriya Sulam Kasab Delima: Keanggunan Benang Emas Pelaminan Adat Pidie",
			idSlug:       "seni-kriya-sulam-kasab-delima-benang-emas-pelaminan",
			idSummary:    "Kerajinan sulam benang emas dan perak pada kain beludru untuk ornamen pelaminan, sarung bantal adat, dan busana pengantin tradisional.",
			idBody:       "Kecamatan Delima, khususnya kawasan Gampong Aree, dikenal sejak lama sebagai sentra utama kerajinan sulaman Kasab (tekat benang emas) di Kabupaten Pidie. Dikerjakan dengan ketelitian tangan para pengrajin wanita secara turun-temurun, kasab memadukan kain beludru berwarna merah marun, hijau tua, atau kuning keemasan dengan sulaman timbul benang emas dan payet berkilau. Motif-motif yang diangkat sarat dengan filosofi kearifan lokal seperti motif pucuk rebung, bungong kalimah, dan awan meucanek. Karya kriya ini menjadi elemen utama dalam upacara perkawinan adat Aceh dan dekorasi rumah kebesaran masyarakat terhormat.",
			enTitle:      "The Heritage of Kasab Gold Embroidery in Delima: Regal Acehnese Bridal Art",
			enSlug:       "heritage-kasab-gold-embroidery-delima-regal-art",
			enSummary:    "Exquisite handcrafted raised embroidery using metallic gold and silver threads on velvet for ceremonial wedding backdrops and royal regalia.",
			enBody:       "Delima District, particularly around Gampong Aree, stands as the revered epicenter of Kasab gold-thread needlework in Pidie Regency. Handcrafted by master artisan women across generations, Kasab craftsmanship applies intricate raised metallic gold and silver filaments upon rich velvet backdrops of crimson, royal green, and indigo. The intricate relief motifs—such as bamboo shoots (pucuk rebung), floral medallions, and stylistic clouds—convey deep philosophical blessings of fertility and nobility, forming the centerpiece of traditional Acehnese bridal chambers and state ceremonies.",
			sourceType:   "community",
			sourceName:   "Dewan Kerajinan Nasional Daerah (Dekranasda) Kabupaten Pidie",
			sourceURL:    "https://dekranasda.acehprov.go.id/",
			sourceNote:   "Katalog Kerajinan Unggulan Tekstil dan Tenun Tradisional Kabupaten Pidie.",
		},
		{
			catKey:       "potensi-daerah",
			districtName: "Mutiara Timur",
			lat:          5.3045,
			lng:          95.9582,
			accuracy:     "accurate",
			address:      "Kawasan Beureunuen, Kecamatan Mutiara Timur, Kabupaten Pidie, Aceh",
			idTitle:      "Sentra Kerupuk Mulieng Beureunuen: Geliat Ekonomi Kreatif Melinjo Legendaris Pidie",
			idSlug:       "sentra-kerupuk-mulieng-beureunuen-ekonomi-kreatif-melinjo",
			idSummary:    "Pusat perdagangan dan produksi emping melinjo berkualitas ekspor yang menjadi urat nadi perekonomian masyarakat Beureunuen.",
			idBody:       "Beureunuen di Kecamatan Mutiara Timur adalah pusat perniagaan dan perdagangan emping melinjo (kerupuk mulieng) terbesar di Provinsi Aceh. Setiap sudut pasar dan pertokoan Beureunuen dipenuhi oleh deretan kerupuk mulieng berbagai ukuran, mulai dari kualitas super tipis satu biji hingga ukuran koin renyah. Biji melinjo yang dipanen dari perkebunan rakyat diolah secara tradisional dengan disangrai pasir panas sebelum ditempa di atas landasan batu pipih. Usaha rumahan yang digerakkan ribuan ibu rumah tangga ini tidak hanya memasok kebutuhan pasar lokal dan nasional, tetapi juga telah diekspor ke berbagai negara Asia Tenggara.",
			enTitle:      "Beureunuen Melinjo Center: The Thriving Creative Economy of Iconic Pidie Crackers",
			enSlug:       "beureunuen-melinjo-center-creative-economy-crackers",
			enSummary:    "The commercial hub for premium hand-flattened melinjo emping crackers, driving household micro-economies across eastern Pidie.",
			enBody:       "The bustling market city of Beureunuen in East Mutiara District serves as Sumatra's foremost trading capital for artisanal melinjo emping (Kerupuk Mulieng). Along its vibrant commerce avenues, thousands of tons of sun-dried melinjo crackers are distributed throughout the archipelago and regional export markets. Sourced from prolific regional orchards, the seeds are sand-roasted over woodfire hearths and individually flattened on stone anvils by skilled home-based craftswomen, sustaining a resilient microeconomic ecosystem that defines Pidie's entrepreneurial identity.",
			sourceType:   "official",
			sourceName:   "Dinas Perdagangan, Koperasi dan UKM Kabupaten Pidie",
			sourceURL:    "https://pidiekab.web.id/",
			sourceNote:   "Laporan Tahunan Sentra Industri Kecil dan Menengah (IKM) Olahan Pangan Pidie.",
		},
	}

	createdCount := 0
	skippedCount := 0

	for _, it := range drafts {
		catID, ok := catMap[it.catKey]
		if !ok || catID == 0 {
			for _, cid := range catMap {
				catID = cid
				break
			}
		}

		// Check if content translation already exists
		var existingID int64
		err := dbPool.QueryRow(ctx, "SELECT content_id FROM content_translations WHERE locale = 'id' AND slug = $1", it.idSlug).Scan(&existingID)
		if err == nil && existingID > 0 {
			skippedCount++
			continue
		}

		var contentID int64
		err = dbPool.QueryRow(ctx, `
			INSERT INTO contents (category_id, address, latitude, longitude, location_accuracy, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $6)
			RETURNING id
		`, catID, it.address, it.lat, it.lng, it.accuracy, adminID).Scan(&contentID)
		if err != nil {
			logger.Error("failed to insert draft content", "title", it.idTitle, "error", err)
			continue
		}

		idBodyJSON, _ := json.Marshal(map[string]string{
			"markdown": it.idBody,
			"html":     "<p>" + it.idBody + "</p>",
		})
		enBodyJSON, _ := json.Marshal(map[string]string{
			"markdown": it.enBody,
			"html":     "<p>" + it.enBody + "</p>",
		})

		// Indonesian translation (status: DRAFT)
		_, err = dbPool.Exec(ctx, `
			INSERT INTO content_translations (content_id, locale, title, slug, summary, body, status, search_vector, created_by, updated_by)
			VALUES ($1, 'id', $2, $3, $4, $5, 'draft', setweight(to_tsvector('simple', COALESCE($2, '')), 'A') || setweight(to_tsvector('simple', COALESCE($4, '')), 'B'), $6, $6)
			ON CONFLICT (content_id, locale) DO NOTHING
		`, contentID, it.idTitle, it.idSlug, it.idSummary, idBodyJSON, adminID)
		if err != nil {
			logger.Error("failed to insert ID draft translation", "contentID", contentID, "error", err)
		}

		// English translation (status: DRAFT)
		_, err = dbPool.Exec(ctx, `
			INSERT INTO content_translations (content_id, locale, title, slug, summary, body, status, search_vector, created_by, updated_by)
			VALUES ($1, 'en', $2, $3, $4, $5, 'draft', setweight(to_tsvector('simple', COALESCE($2, '')), 'A') || setweight(to_tsvector('simple', COALESCE($4, '')), 'B'), $6, $6)
			ON CONFLICT (content_id, locale) DO NOTHING
		`, contentID, it.enTitle, it.enSlug, it.enSummary, enBodyJSON, adminID)
		if err != nil {
			logger.Error("failed to insert EN draft translation", "contentID", contentID, "error", err)
		}

		// Insert verified content source citation
		_, err = dbPool.Exec(ctx, `
			INSERT INTO content_sources (content_id, source_name, source_type, source_url, notes, created_by)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, contentID, it.sourceName, it.sourceType, it.sourceURL, it.sourceNote, adminID)
		if err != nil {
			logger.Error("failed to insert content source", "contentID", contentID, "error", err)
		}

		createdCount++
		fmt.Printf(" [+] Draft Created (#%d): %s [%s]\n", contentID, it.idTitle, it.districtName)
	}

	fmt.Printf("==================================================\n")
	fmt.Printf("Draft Seeding Summary:\n")
	fmt.Printf(" - Newly Created Drafts : %d\n", createdCount)
	fmt.Printf(" - Existing / Skipped   : %d\n", skippedCount)
	fmt.Printf("All draft contents are available in /cms/contents\n")
	fmt.Printf("==================================================\n")
}
