import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	server: {
		proxy: {
			// In development, forward API requests to the Go server (make run).
			'/api': 'http://localhost:8080'
		}
	},
	plugins: [
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/u).includes('node_modules') ? undefined : true
			},
			// Build a single-page app: every route that isn't a static file is served
			// index.html, and routing happens in the browser.
			adapter: adapter({ fallback: 'index.html' })
		})
	]
});
