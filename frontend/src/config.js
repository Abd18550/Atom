// API Configuration
// In development, Vite serves on port 5173 and proxies /api to localhost:8080
// In production, set VITE_API_URL to the backend URL
const API_BASE_URL = import.meta.env.VITE_API_URL || 'https://atom-kk7h.onrender.com'

export { API_BASE_URL }
