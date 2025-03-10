// Manga series interface
export interface Manga {
    id: string;
    title: string;
    description: string;
    coverImage: string;
    genres: string[];
    author: string;
    artist?: string;
    status: string;
    publishedYear?: number;
    lastUpdated: string;
    chapterCount: number;
    altTitles?: string[];
  }
  
  // Chapter interface
  export interface Chapter {
    id: string;
    mangaId: string;
    number: number;
    title: string;
    releaseDate: string;
    pageCount: number;
    volume?: number;
    special?: boolean;
  }
  
  // Page interface
  export interface Page {
    number: number;
    imageUrl: string;
    chapterId: string;
    mangaId: string;
    totalPages?: number;
    nextPage?: number;
    prevPage?: number;
    nextChapter?: string;
    prevChapter?: string;
  }