import { Manga, Chapter, Page } from '../types/manga';

// Base API URL
const API_BASE = '/api';

// Fetch all manga
export async function getAllManga(): Promise<Manga[]> {
  const response = await fetch(`${API_BASE}/manga`);
  if (!response.ok) {
    throw new Error(`Failed to fetch manga: ${response.statusText}`);
  }
  return response.json();
}

// Fetch a specific manga by ID
export async function getMangaById(id: string): Promise<Manga> {
  const response = await fetch(`${API_BASE}/manga/${id}`);
  if (!response.ok) {
    throw new Error(`Failed to fetch manga ${id}: ${response.statusText}`);
  }
  return response.json();
}

// Fetch chapters for a manga
export async function getChapters(mangaId: string): Promise<Chapter[]> {
  const response = await fetch(`${API_BASE}/manga/${mangaId}/chapters`);
  if (!response.ok) {
    throw new Error(`Failed to fetch chapters for manga ${mangaId}: ${response.statusText}`);
  }
  return response.json();
}

// Fetch a specific page
export async function getPage(mangaId: string, chapterNumber: number, pageNumber: number): Promise<Page> {
  const response = await fetch(`${API_BASE}/manga/${mangaId}/chapter/${chapterNumber}/page/${pageNumber}`);
  if (!response.ok) {
    throw new Error(`Failed to fetch page ${pageNumber} of chapter ${chapterNumber}: ${response.statusText}`);
  }
  return response.json();
}