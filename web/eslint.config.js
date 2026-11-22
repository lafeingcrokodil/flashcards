import prettier from 'eslint-config-prettier';
import path from 'node:path';
import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import { defineConfig, includeIgnoreFile } from 'eslint/config';
import globals from 'globals';
import ts from 'typescript-eslint';

const gitignorePath = path.resolve(import.meta.dirname, '.gitignore');

export default defineConfig(
	includeIgnoreFile(gitignorePath),
	js.configs.all,
	// Start with every rule enabled, then disable the ones we decide we don't
	// want, with a comment explaining why.
	ts.configs.all,
	svelte.configs.all,
	prettier,
	svelte.configs.prettier,
	{
		languageOptions: { globals: { ...globals.browser, ...globals.node } },
		rules: {
			// TypeScript already checks for undefined variables, so typescript-eslint
			// recommends turning this rule off. See:
			// https://typescript-eslint.io/troubleshooting/faqs/eslint/#i-get-errors-from-the-no-undef-rule-about-global-variables-not-being-defined-even-though-there-are-no-typescript-errors
			'no-undef': 'off'
		}
	},
	{
		// Type information is needed for the type-aware rules.
		languageOptions: {
			parserOptions: {
				projectService: {
					// These config files aren't part of the TypeScript project.
					allowDefaultProject: ['eslint.config.js', 'prettier.config.js']
				},
				extraFileExtensions: ['.svelte'],
				parser: ts.parser
			}
		}
	},
	{
		rules: {
			// Require TypeScript in <script> blocks (the default is to forbid the
			// lang attribute, i.e. to require JavaScript).
			'svelte/block-lang': ['error', { script: 'ts' }],
			// Allow comments that wrap onto several lines (each continuation line
			// would otherwise have to start with a capital letter).
			'capitalized-comments': ['error', 'always', { ignoreConsecutiveComments: true }],
			// Keys and imports are better ordered by meaning than alphabetically,
			// and sort-imports can't fix the order of import statements itself.
			'sort-keys': 'off',
			'sort-imports': 'off',
			// Combining consecutive declarations into one statement makes them
			// harder to read and to change.
			'one-var': 'off',
			// Ternaries are fine for simple conditional values.
			'no-ternary': 'off',
			// TypeScript already tracks where undefined is allowed, and some APIs
			// (e.g. Svelte's compiler options) use it as a meaningful value.
			'no-undefined': 'off',
			// Many types from libraries can't be made read-only, so this rule
			// would require frequent exceptions.
			'@typescript-eslint/prefer-readonly-parameter-types': 'off',
			// 0 and 1 (e.g. in comparisons with a length, or to step through a
			// list) are clearer as literals than as named constants.
			'@typescript-eslint/no-magic-numbers': ['error', { ignore: [0, 1] }],
			// Picking the selector type by how many elements it matches (e.g. an
			// ID for a single element) makes styles change form whenever the
			// markup does. Classes and element types are both fine.
			'svelte/consistent-selector-style': 'off',
			// The rule's defaults, plus an exception for property names that need
			// quotes (e.g. 'no-undef' as the name of an ESLint rule).
			'@typescript-eslint/naming-convention': [
				'error',
				{
					selector: 'default',
					format: ['camelCase'],
					leadingUnderscore: 'allow',
					trailingUnderscore: 'allow'
				},
				{ selector: 'import', format: ['camelCase', 'PascalCase'] },
				{
					selector: 'variable',
					format: ['camelCase', 'UPPER_CASE'],
					leadingUnderscore: 'allow',
					trailingUnderscore: 'allow'
				},
				{ selector: 'typeLike', format: ['PascalCase'] },
				{ selector: 'objectLiteralProperty', modifiers: ['requiresQuotes'], format: null }
			]
		}
	}
);
