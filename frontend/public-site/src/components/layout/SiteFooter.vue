<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

const route = useRoute()
const locale = computed<'id' | 'en'>(() => (route.path.startsWith('/en') ? 'en' : 'id'))
const isId = computed(() => locale.value === 'id')
</script>

<template>
  <footer class="site-footer">
    <div class="footer-inner">
      <div class="footer-grid">
        <!-- Col 1: About Platform -->
        <div class="footer-col col-main">
          <div class="footer-brand">
            <span class="brand-badge">PK</span>
            <span class="brand-name">PidieKab</span>
          </div>
          <p class="brand-desc">
            {{ isId
              ? 'Platform promosi daerah independen untuk memperkenalkan potensi alam, kekayaan budaya, kuliner legendaris, dan UMKM Kabupaten Pidie, Aceh.'
              : 'An independent regional discovery platform showcasing the natural landscapes, cultural heritage, culinary treasures, and local crafts of Pidie Regency, Aceh.'
            }}
          </p>
          <div class="disclaimer-box">
            <strong>{{ isId ? 'Pernyataan Independen:' : 'Independent Notice:' }}</strong>
            <p>
              {{ isId
                ? 'PidieKab adalah inisiatif platform digital independen masyarakat, bukan situs resmi pemerintah ataupun representasi resmi Pemerintah Kabupaten Pidie.'
                : 'PidieKab is an independent community digital initiative and is not an official government website or representative of the Pidie Regency Government.'
              }}
            </p>
          </div>
        </div>

        <!-- Col 2: Explore -->
        <div class="footer-col">
          <h4 class="col-title">{{ isId ? 'Eksplorasi Pidie' : 'Explore Pidie' }}</h4>
          <ul class="footer-links">
            <li>
              <RouterLink :to="isId ? '/id/wisata-destinasi/' : '/en/tourism-destinations/'">
                {{ isId ? 'Wisata & Alam' : 'Tourism & Nature' }}
              </RouterLink>
            </li>
            <li>
              <RouterLink :to="isId ? '/id/kuliner-umkm/' : '/en/culinary-local-business/'">
                {{ isId ? 'Kuliner & UMKM' : 'Culinary & Local Crafts' }}
              </RouterLink>
            </li>
            <li>
              <RouterLink :to="isId ? '/id/budaya-sejarah/' : '/en/culture-history/'">
                {{ isId ? 'Seni Budaya & Sejarah' : 'Heritage & History' }}
              </RouterLink>
            </li>
          </ul>
        </div>

        <!-- Col 3: Information -->
        <div class="footer-col">
          <h4 class="col-title">{{ isId ? 'Informasi' : 'Information' }}</h4>
          <ul class="footer-links">
            <li>
              <RouterLink :to="isId ? '/id/tentang-kami/' : '/en/about-us/'">
                {{ isId ? 'Tentang PidieKab' : 'About PidieKab' }}
              </RouterLink>
            </li>
            <li>
              <RouterLink :to="isId ? '/id/kebijakan-privasi/' : '/en/privacy/'">
                {{ isId ? 'Kebijakan Privasi' : 'Privacy Policy' }}
              </RouterLink>
            </li>
            <li>
              <RouterLink :to="isId ? '/id/syarat-ketentuan/' : '/en/terms/'">
                {{ isId ? 'Syarat & Ketentuan' : 'Terms & Conditions' }}
              </RouterLink>
            </li>
          </ul>
        </div>

        <!-- Col 4: Contact -->
        <div class="footer-col">
          <h4 class="col-title">{{ isId ? 'Kontak & Wilayah' : 'Contact & Region' }}</h4>
          <p class="contact-info">
            📍 Sigli, Kabupaten Pidie, Aceh, Indonesia<br />
            ✉️ kontak@pidiekab.id
          </p>
          <div class="copyright-sub">
            &copy; {{ new Date().getFullYear() }} PidieKab. Rights-cleared regional content.
          </div>
        </div>
      </div>
    </div>
  </footer>
</template>

<style scoped>
.site-footer {
  background: #ffffff;
  color: #475569;
  padding: 3.5rem 1.5rem 2.5rem 1.5rem;
  border-top: 1px solid #e2e8f0;
  margin-top: 4rem;
}

.footer-inner {
  max-width: 1200px;
  margin: 0 auto;
}

.footer-grid {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1.25fr;
  gap: 2.5rem;
}

@media (max-width: 900px) {
  .footer-grid {
    grid-template-columns: 1fr 1fr;
  }
  .col-main {
    grid-column: span 2;
  }
}

@media (max-width: 580px) {
  .footer-grid {
    grid-template-columns: 1fr;
  }
  .col-main {
    grid-column: span 1;
  }
}

.footer-brand {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  margin-bottom: 0.75rem;
}

.brand-badge {
  background: #047857;
  color: #ffffff;
  font-weight: 800;
  width: 32px;
  height: 32px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.95rem;
}

.brand-name {
  font-size: 1.25rem;
  font-weight: 700;
  color: #0f172a;
}

.brand-desc {
  font-size: 0.85rem;
  line-height: 1.6;
  margin: 0 0 1rem 0;
  color: #64748b;
}

.disclaimer-box {
  background: #ecfdf5;
  border-left: 3px solid #047857;
  padding: 0.65rem 0.85rem;
  border-radius: 0 4px 4px 0;
}

.disclaimer-box strong {
  display: block;
  font-size: 0.75rem;
  color: #047857;
  margin-bottom: 0.2rem;
}

.disclaimer-box p {
  margin: 0;
  font-size: 0.725rem;
  line-height: 1.4;
  color: #475569;
}

.col-title {
  color: #0f172a;
  font-size: 0.95rem;
  font-weight: 600;
  margin: 0 0 1rem 0;
}

.footer-links {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.footer-links a {
  color: #475569;
  text-decoration: none;
  font-size: 0.875rem;
  transition: color 0.15s;
}

.footer-links a:hover {
  color: #047857;
}

.contact-info {
  font-size: 0.85rem;
  line-height: 1.6;
  margin: 0 0 1rem 0;
}

.copyright-sub {
  font-size: 0.75rem;
  color: #64748b;
}
</style>
