export interface Translation {
  locale: 'id' | 'en'
  title: string
  slug: string
  description: string
  body: string
  status: 'published' | 'draft' | 'archived'
}

export interface CategoryTranslation {
  locale: 'id' | 'en'
  name: string
  slug: string
  description: string
}

export interface Category {
  id: string
  translations: Record<'id' | 'en', CategoryTranslation>
}

export interface ContentItem {
  id: string
  categoryId: string
  translations: Record<'id' | 'en', Translation>
}

export const CATEGORIES: Category[] = [
  {
    id: 'cat-wisata-alam',
    translations: {
      id: {
        locale: 'id',
        name: 'Wisata Alam',
        slug: 'wisata-alam',
        description: 'Pesona alam pegunungan, lembah hijau, sungai, dan pesisir pantai Kabupaten Pidie.',
      },
      en: {
        locale: 'en',
        name: 'Nature & Landscapes',
        slug: 'nature-landscapes',
        description: 'Lush valleys, mountain ranges, rivers, and coastal beaches of Pidie Regency.',
      },
    },
  },
  {
    id: 'cat-sejarah-budaya',
    translations: {
      id: {
        locale: 'id',
        name: 'Sejarah & Warisan Budaya',
        slug: 'sejarah-budaya',
        description: 'Situs bersejarah, cagar budaya, makam ulama dan raja, serta jejak Kesultanan Aceh di Pidie.',
      },
      en: {
        locale: 'en',
        name: 'History & Cultural Heritage',
        slug: 'history-heritage',
        description: 'Historical landmarks, heritage mosques, royal tombs, and cultural legacy in Pidie.',
      },
    },
  },
  {
    id: 'cat-kuliner-khas',
    translations: {
      id: {
        locale: 'id',
        name: 'Kuliner Khas',
        slug: 'kuliner-khas',
        description: 'Penganan tradisional, kue legit warisan leluhur, kopi Robusta khas, dan cita rasa autentik Pidie.',
      },
      en: {
        locale: 'en',
        name: 'Traditional Culinary',
        slug: 'traditional-culinary',
        description: 'Authentic confections, traditional dishes, Pidie Robusta coffee, and culinary traditions.',
      },
    },
  },
  {
    id: 'cat-potensi-daerah',
    translations: {
      id: {
        locale: 'id',
        name: 'Potensi Daerah & Ekonomi',
        slug: 'potensi-daerah',
        description: 'Komoditas unggulan seperti melinjo, kopi, pertanian padi, perikanan pesisir, dan UMKM.',
      },
      en: {
        locale: 'en',
        name: 'Regional Economy & Products',
        slug: 'regional-products',
        description: 'Flagship commodities including melinjo crackers, agriculture, fisheries, and crafts.',
      },
    },
  },
  {
    id: 'cat-layanan-publik',
    translations: {
      id: {
        locale: 'id',
        name: 'Layanan & Info Publik',
        slug: 'layanan-publik',
        description: 'Panduan informasi publik, sarana fasilitas umum, dan kontak darurat di wilayah Kabupaten Pidie.',
      },
      en: {
        locale: 'en',
        name: 'Public Information & Services',
        slug: 'public-services',
        description: 'Public visitor guidance, emergency contacts, and community amenities in Pidie.',
      },
    },
  },
]

export const CONTENTS: ContentItem[] = [
  {
    id: 'content-guha-tujoh',
    categoryId: 'cat-sejarah-budaya',
    translations: {
      id: {
        locale: 'id',
        title: 'Eksplorasi Guha Tujoh Laweung: Keajaiban Karst Pesisir Selat Malaka',
        slug: 'eksplorasi-guha-tujoh-laweung-keajaiban-karst',
        description: 'Gua alam bersejarah dengan formasi stalaktit spektakuler dan legenda tujuh pintu di pesisir barat Selat Malaka.',
        body: 'Guha Tujoh (Gua Tujuh) terletak di perbukitan kapur Gampong Laweung, Kecamatan Muara Tiga. Gua karst ini memiliki tujuh lorong utama yang menurut riwayat lokal saling terhubung. Di dalam gua terdapat formasi stalaktit dan stalagmit yang menyerupai pelaminan pengantin, kubah masjid, dan formasi bebatuan unik lainnya. Udara di dalam gua terasa sejuk berpadu dengan hembusan angin laut Selat Malaka yang hanya berjarak beberapa ratus meter dari bibir gua. Lokasi ini merupakan destinasi wisata geologi dan spiritual yang sarat nilai sejarah daerah.',
        status: 'published',
      },
      en: {
        locale: 'en',
        title: 'Exploring Guha Tujoh Laweung: The Coastal Karst Wonder on Malacca Strait',
        slug: 'exploring-guha-tujoh-laweung-coastal-karst',
        description: 'A historic natural limestone cave featuring magnificent stalactite formations and the legend of seven hidden chambers.',
        body: 'Guha Tujoh (The Seven Caves) is nestled in the limestone hills of Laweung, Muara Tiga District. This geological wonder features seven primary caverns that folklore links to legendary passages. Inside, visitor galleries showcase ancient stalactites and stalagmites shaped by natural erosion over millennia. The cool underground atmosphere combined with refreshing breezes from the Malacca Strait makes it one of Pidie premier eco-tourism and cultural highlights.',
        status: 'published',
      },
    },
  },
  {
    id: 'content-masjid-po-teumeureuhom',
    categoryId: 'cat-sejarah-budaya',
    translations: {
      id: {
        locale: 'id',
        title: 'Masjid Po Teumeureuhom Mila: Jejak Spiritual dan Diplomasi Abad ke-17',
        slug: 'masjid-po-teumeureuhom-mila-jejak-spiritual',
        description: 'Situs cagar budaya masjid bersejarah peninggalan era Kesultanan Aceh yang sarat nilai arsitektur kayu tradisional.',
        body: 'Masjid Po Teumeureuhom di Kecamatan Mila adalah salah satu monumen spiritual tertua di Kabupaten Pidie. Dibangun dengan konstruksi tiang kayu jati pilihan berukir khas Aceh, masjid ini menjadi saksi kejayaan Po Teumeureuhom dalam menyebarkan syiar Islam dan mempersatukan masyarakat Pidie. Kompleks masjid ini juga dikelilingi pemakaman tokoh bersejarah dan sumur tua yang airnya tidak pernah kering.',
        status: 'published',
      },
      en: {
        locale: 'en',
        title: 'Po Teumeureuhom Mosque of Mila: 17th-Century Spiritual Heritage',
        slug: 'po-teumeureuhom-mosque-mila-spiritual-heritage',
        description: 'A protected heritage sanctuary reflecting classical Acehnese timber craft and historical diplomacy.',
        body: 'Located in Dayah Sinthop, Mila District, the Po Teumeureuhom Mosque stands as a timeless testament to 17th-century Acehnese architecture. Built using durable hand-carved teak pillars, the site played an instrumental role in regional scholarship and social unity during the golden age of the Aceh Sultanate. The sanctuary complex includes ancient historic tombs and traditional water reservoirs preserved with reverence by the local community.',
        status: 'published',
      },
    },
  },
  {
    id: 'content-lembah-tangse',
    categoryId: 'cat-wisata-alam',
    translations: {
      id: {
        locale: 'id',
        title: 'Pesona Lembah Tangse: Sejuknya Udara Pegunungan dan Kopi Robusta Asli',
        slug: 'pesona-lembah-tangse-pegunungan-dan-kopi',
        description: 'Lembah permai di kaki Bukit Barisan dengan aliran sungai jernih, air terjun alami, dan hamparan kebun kopi Robusta.',
        body: 'Tangse adalah surga tersembunyi di pedalaman Pidie. Berada di ketinggian lebih dari 600 meter di atas permukaan laut, Tangse memiliki iklim sejuk sepanjang tahun. Daerah ini dialiri sungai berbatu cadas yang cocok untuk arung jeram dan rekreasi keluarga. Di sepanjang lereng bukit terbentang perkebunan kopi Robusta Tangse yang terkenal dengan aroma rempah alaminya. Wisatawan dapat menikmati sajian kopi tubruk segar langsung di pinggir sungai dengan latar belakang perbukitan hijau yang memukau.',
        status: 'published',
      },
      en: {
        locale: 'en',
        title: 'The Splendor of Tangse Valley: Mountain Streams and Robusta Heritage',
        slug: 'splendor-of-tangse-valley-nature-and-coffee',
        description: 'A picturesque valley framed by the Barisan Mountains, clear flowing streams, and renowned highland coffee estates.',
        body: 'Tangse is a serene mountainous haven in the southern highlands of Pidie. Positioned over 600 meters above sea level, it enjoys cool temperatures and verdant scenery year-round. Its crystal-clear mountain streams offer ideal spots for eco-tourism and outdoor relaxation. Tangse distinctive Robusta coffee, cultivated across misty hillside slopes, is famed for its rich aroma and earthy spice notes.',
        status: 'published',
      },
    },
  },
  {
    id: 'content-kerupuk-mulieng',
    categoryId: 'cat-kuliner-khas',
    translations: {
      id: {
        locale: 'id',
        title: 'Kelezatan Gurih Kerupuk Mulieng Beureunuen: Warisan Melinjo Terbaik Pidie',
        slug: 'kelezatan-kerupuk-mulieng-beureunuen-melinjo',
        description: 'Kerupuk emping melinjo legendaris dari Beureunuen yang diproduksi secara tradisional dengan kerenyahan tiada tara.',
        body: 'Kerupuk Mulieng (emping melinjo) adalah ikon kuliner dan komoditas kebanggaan masyarakat Pidie. Berpusat di kawasan Beureunuen dan sekitarnya, para pengrajin wanita memproses biji melinjo pilihan dengan cara disangrai di atas pasir panas lalu dipipihkan menggunakan palu besi secara presisi. Kerupuk mulieng Pidie terkenal tipis, renyah, dan memiliki cita rasa gurih alami tanpa bahan pengawet. Produk ini telah menjadi oleh-oleh wajib bagi setiap pelancong yang melintasi jalur Banda Aceh - Medan.',
        status: 'published',
      },
      en: {
        locale: 'en',
        title: 'Savory Crunch of Beureunuen Kerupuk Mulieng: Pidies Premium Melinjo',
        slug: 'savory-beureunuen-kerupuk-mulieng-melinjo',
        description: 'The celebrated artisan melinjo crisps of Beureunuen, crafted by hand using centuries-old roasting traditions.',
        body: 'Kerupuk Mulieng (melinjo crisps) is Pidies signature gastronomic treasure. Centered in Beureunuen marketplace, skilled local artisans sand-roast premium melinjo seeds before flattening them with hand hammers into delicate, golden wafers. Noted for its crisp texture and authentic bittersweet flavor, Pidie melinjo crackers are revered across the Indonesian archipelago and international diaspora as an indispensable delicacy.',
        status: 'published',
      },
    },
  },
  {
    id: 'content-kue-adee',
    categoryId: 'cat-kuliner-khas',
    translations: {
      id: {
        locale: 'id',
        title: 'Kue Adee Khas Pidie: Manis Legit Tradisi Kue Basah Aceh',
        slug: 'kue-adee-khas-pidie-manis-legit-tradisi',
        description: 'Kue panggang tradisional berbahan singkong atau terigu dengan aroma daun pandan dan santan kental yang lumer di mulut.',
        body: 'Kue Adee (atau dikenal luas sebagai bingka khas Pidie) adalah kudapan manis legendaris yang memiliki tekstur lembut di dalam dan renyah kecokelatan di bagian luar. Terbuat dari perpaduan singkong parut halus (atau tepung terigu pilihan), santan kelapa kental murni, gula pasir, telur bebek, dan perasan daun pandan suji, kue adee selalu dipanggang dalam loyang bundar dengan aroma wangi yang menggugah selera. Kue ini menjadi hidangan wajib dalam upacara adat, hari raya Idulfitri, dan hantaran pernikahan masyarakat Pidie.',
        status: 'published',
      },
      en: {
        locale: 'en',
        title: 'Traditional Adee Cake: Sweet Fragrance of Pidie Confectionery',
        slug: 'traditional-adee-cake-pidie-confectionery',
        description: 'A luscious baked cassava cake enriched with velvety coconut milk, duck eggs, and aromatic pandan leaves.',
        body: 'Kue Adee is Pidies premier sweet confection. Featuring a caramelized golden crust and a soft, melt-in-the-mouth center, it is crafted from finely grated cassava or premium wheat flour, freshly pressed coconut milk, free-range duck eggs, and native pandan extracts. Baked to perfection in circular pans, Adee cake is traditionally shared during festive gatherings, family reunions, and festive Eid celebrations.',
        status: 'published',
      },
    },
  },
  {
    id: 'content-pantai-pelangi',
    categoryId: 'cat-wisata-alam',
    translations: {
      id: {
        locale: 'id',
        title: 'Pantai Pelangi Kota Sigli: Titik Temu Masyarakat di Pesisir Selat Malaka',
        slug: 'pantai-pelangi-kota-sigli-pesisir-selat-malaka',
        description: 'Kawasan wisata tepi laut di jantung ibukota Sigli dengan panorama matahari terbit, kuliner laut, dan ruang publik terbuka.',
        body: 'Pantai Pelangi yang berada tepat di tepi teluk Kota Sigli merupakan ruang terbuka hijau dan rekreasi paling populer bagi warga Pidie. Dihiasi oleh jalur pejalan kaki yang rapi, pepohonan rindang, dan pemandangan perahu nelayan tradisional yang bersandar, pantai ini menjadi magnet bagi keluarga di sore hari. Pengunjung dapat menikmati sebutir kelapa muda segar, sajian rujak u-groe (rujak batok kelapa muda khas Pidie), atau mi caluk khas Sigli sembari menyaksikan ombak Selat Malaka yang tenang.',
        status: 'published',
      },
      en: {
        locale: 'en',
        title: 'Sigli Rainbow Beach: Vibrant Coastal Gathering on the Malacca Strait',
        slug: 'sigli-rainbow-beach-vibrant-coastal-waterfront',
        description: 'A scenic seaside promenade in the heart of Sigli capital, featuring sunrise vistas, public greenery, and local street delicacies.',
        body: 'Pantai Pelangi (Rainbow Beach), located along the waterfront of downtown Sigli, serves as the bustling civic and leisure center for Pidie residents. Lined with pedestrian walkways, coastal flora, and historic harbor jetties, the promenade is especially vibrant at sunset. Visitors can savor refreshing green coconuts, regional Rujak U-Groe, and iconic Mie Caluk noodles along the serene shores of the Malacca Strait.',
        status: 'published',
      },
    },
  },
]

/**
 * Helper to get all published routes that must be generated into static HTML.
 * Strictly excludes any draft or archived routes.
 */
export function getPublishedRoutes(): string[] {
  const routes: string[] = ['/id/', '/en/']

  for (const cat of CATEGORIES) {
    routes.push(`/id/${cat.translations.id.slug}/`)
    routes.push(`/en/${cat.translations.en.slug}/`)
  }

  for (const item of CONTENTS) {
    const cat = CATEGORIES.find((c) => c.id === item.categoryId)
    if (!cat) continue

    if (item.translations.id.status === 'published') {
      routes.push(`/id/${cat.translations.id.slug}/${item.translations.id.slug}/`)
    }
    if (item.translations.en.status === 'published') {
      routes.push(`/en/${cat.translations.en.slug}/${item.translations.en.slug}/`)
    }
  }

  return routes
}
