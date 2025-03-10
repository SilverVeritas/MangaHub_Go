import './styles/main.css';
import { initRouter } from './routes';
import { renderHomePage } from './components/home-page';

// Initialize the application
function init() {
  // Initial render
  renderHomePage();
  
  // Set up router
  initRouter();
  
  console.log('MangaHub frontend initialized');
}

// Start the app when DOM is ready
document.addEventListener('DOMContentLoaded', init);