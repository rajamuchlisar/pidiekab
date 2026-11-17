import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { execSync } from 'node:child_process'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const rootDir = path.resolve(__dirname, '..')

console.log('[native-prerender] Starting Vite native pre-rendering comparison build...')
// This demonstrates the low-level Vite SSR pre-rendering alternative
// comparing configuration overhead, build time, and output predictability.
console.log('[native-prerender] Native approach requires 2-stage build (client + server bundle).')
