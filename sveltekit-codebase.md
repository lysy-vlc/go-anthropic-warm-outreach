# .gitignore

```
node_modules

# Output
.output
.vercel
.netlify
.wrangler
/.svelte-kit
/build

# OS
.DS_Store
Thumbs.db

# Env
.env
.env.*
!.env.example
!.env.test
.env.local

# Vite
vite.config.js.timestamp-*
vite.config.ts.timestamp-*

# db
local.db
```

# .npmrc

```
engine-strict=true

```

# .prettierignore

```
# Package Managers
package-lock.json
pnpm-lock.yaml
yarn.lock

```

# .prettierrc

```
{
	"useTabs": true,
	"singleQuote": true,
	"trailingComma": "none",
	"printWidth": 100,
	"plugins": ["prettier-plugin-svelte", "prettier-plugin-tailwindcss"],
	"overrides": [
		{
			"files": "*.svelte",
			"options": {
				"parser": "svelte"
			}
		}
	]
}

```

# eslint.config.js

```js
import prettier from "eslint-config-prettier";
import js from "@eslint/js";
import { includeIgnoreFile } from "@eslint/compat";
import svelte from "eslint-plugin-svelte";
import globals from "globals";
import { fileURLToPath } from "node:url";
import ts from "typescript-eslint";
const gitignorePath = fileURLToPath(new URL("./.gitignore", import.meta.url));

export default ts.config(
  includeIgnoreFile(gitignorePath),
  js.configs.recommended,
  ...ts.configs.recommended,
  ...svelte.configs["flat/recommended"],
  prettier,
  ...svelte.configs["flat/prettier"],
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
  },
  {
    files: ["**/*.svelte"],

    languageOptions: {
      parserOptions: {
        parser: ts.parser,
      },
    },
  }
);
```

# postcss.config.js

```js
export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
};
```

# README.md

```md
# sv

Everything you need to build a Svelte project, powered by [`sv`](https://github.com/sveltejs/cli).

## Creating a project

If you're seeing this, you've probably already done this step. Congrats!

\`\`\`bash

# create a new project in the current directory

npx sv create

# create a new project in my-app

npx sv create my-app
\`\`\`

## Developing

Once you've created a project and installed dependencies with `npm install` (or `pnpm install` or `yarn`), start a development server:

\`\`\`bash
npm run dev

# or start the server and open the app in a new browser tab

npm run dev -- --open
\`\`\`

## Building

To create a production version of your app:

\`\`\`bash
npm run build
\`\`\`

You can preview the production build with `npm run preview`.

> To deploy your app, you may need to install an [adapter](https://svelte.dev/docs/kit/adapters) for your target environment.
```

# src/app.css

```css
@import "tailwindcss/base";
@import "tailwindcss/components";
@import "tailwindcss/utilities";
```

# src/app.d.ts

```ts
// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
  namespace App {
    interface Locals {}
    interface PageData {}
    interface Platform {}
    interface Error {}
  }
}

export {};
```

# src/app.html

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <link rel="icon" href="%sveltekit.assets%/favicon.png" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    %sveltekit.head%
  </head>
  <body data-sveltekit-preload-data="hover">
    <div style="display: contents">%sveltekit.body%</div>
  </body>
</html>
```

# src/env.d.ts

```ts
/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_ANTHROPIC_API_KEY: string;
  readonly VITE_AIRTABLE_ACCESS_TOKEN: string;
  readonly VITE_AIRTABLE_BASE_ID: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

declare namespace App {
  interface Locals {}
  interface PageData {}
  interface Platform {}
  interface Error {}
}
```

# src/lib/db.ts

```ts
import Database from "better-sqlite3";
import { join } from "path";

// Użyjmy join do stworzenia ścieżki względnej do bazy danych
const dbPath = join(process.cwd(), "local.db");
const db = new Database(dbPath);

// Dodajmy interfejs dla wartości z bazy
interface ConfigRow {
  value: string;
}

interface AirtableFieldRow {
  id: number;
  name: string;
  airtable_name: string;
  type: string;
  required: number;
}

// Initialize database with required tables
db.exec(`
  CREATE TABLE IF NOT EXISTS config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
  );

  CREATE TABLE IF NOT EXISTS airtable_fields (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    airtable_name TEXT NOT NULL,
    type TEXT NOT NULL,
    required BOOLEAN NOT NULL DEFAULT 0
  );
`);

export interface AirtableField {
  id?: number;
  name: string;
  airtable_name: string;
  type: "text" | "email" | "url" | "textarea";
  required: boolean;
}

export interface Config {
  anthropic_api_key: string;
  airtable_access_token: string;
  airtable_base_id: string;
  airtable_table_name: string;
  default_language: string;
  airtable_fields: AirtableField[];
}

export function saveConfig(config: Config) {
  try {
    const stmt = db.prepare(
      "INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)"
    );

    // Begin transaction
    const transaction = db.transaction(() => {
      // Save API keys and basic config
      stmt.run("anthropic_api_key", config.anthropic_api_key);
      stmt.run("airtable_access_token", config.airtable_access_token);
      stmt.run("airtable_base_id", config.airtable_base_id);
      stmt.run("airtable_table_name", config.airtable_table_name);
      stmt.run("default_language", config.default_language);

      // Clear existing fields
      db.prepare("DELETE FROM airtable_fields").run();

      // Insert new fields
      const insertField = db.prepare(`
        INSERT INTO airtable_fields (name, airtable_name, type, required)
        VALUES (?, ?, ?, ?)
      `);

      for (const field of config.airtable_fields) {
        insertField.run(
          field.name,
          field.airtable_name,
          field.type,
          field.required ? 1 : 0
        );
      }
    });

    transaction();
    return true;
  } catch (error) {
    console.error("Error saving config to database:", error);
    throw new Error("Failed to save configuration to database");
  }
}

export function loadConfig(): Config {
  try {
    const getConfigValue = db.prepare<[string], ConfigRow>(
      "SELECT value FROM config WHERE key = ?"
    );
    const getFields = db.prepare<[], AirtableFieldRow[]>(
      "SELECT * FROM airtable_fields"
    );

    const config: Config = {
      anthropic_api_key: getConfigValue.get("anthropic_api_key")?.value || "",
      airtable_access_token:
        getConfigValue.get("airtable_access_token")?.value || "",
      airtable_base_id: getConfigValue.get("airtable_base_id")?.value || "",
      airtable_table_name:
        getConfigValue.get("airtable_table_name")?.value || "",
      default_language: getConfigValue.get("default_language")?.value || "",
      airtable_fields: getFields.all().map((field) => ({
        id: field.id,
        name: field.name,
        airtable_name: field.airtable_name,
        type: field.type as AirtableField["type"],
        required: Boolean(field.required),
      })),
    };

    return config;
  } catch (error) {
    console.error("Error loading config from database:", error);
    // Zwróć domyślną konfigurację jeśli nie ma zapisanej
    return {
      anthropic_api_key: "",
      airtable_access_token: "",
      airtable_base_id: "",
      airtable_table_name: "",
      default_language: "",
      airtable_fields: [],
    };
  }
}

export function getRequiredConfig() {
  const config = loadConfig();
  if (
    !config.anthropic_api_key ||
    !config.airtable_access_token ||
    !config.airtable_base_id ||
    !config.airtable_table_name
  ) {
    throw new Error(
      "Missing required configuration. Please visit the configuration page."
    );
  }
  return config;
}
```

# src/lib/index.ts

```ts
// place files you want to import through the `$lib` alias in this folder.
```

# src/routes/+layout.svelte

```svelte
<script lang="ts">
	import '../app.css';
	let { children } = $props();
</script>

<nav class="bg-gray-800 text-white mb-4">
  <div class="container mx-auto px-4 py-2 flex justify-between items-center">
    <a href="/" class="text-lg font-semibold">AI Outreach Generator</a>
    <a href="/config" class="text-sm hover:text-gray-300">Configuration</a>
  </div>
</nav>

{@render children()}

```

# src/routes/+page.svelte

```svelte
<script lang="ts">
  import { onMount } from 'svelte';

  interface Contact {
    id: string;
    fullname: string;
    company_name: string;
    business_segment: string;
    website: string;
    phone: string;
    city: string;
    country: string;
    email: string;
    outreach_text?: string;
    error?: string;
  }

  let contacts: Contact[] = [];
  let prompt: string = '';
  let loading = false;
  let fetchingContacts = false;
  let currentContact: Contact | null = null;
  let error: string | null = null;
  let selectedLanguage: string;
  let config: Config;

  const languages = [
    { code: 'en', name: 'English' },
    { code: 'pl', name: 'Polish' },
    { code: 'de', name: 'German' },
    { code: 'es', name: 'Spanish' },
    { code: 'fr', name: 'French' }
  ];

  async function fetchContacts() {
    error = null;
    fetchingContacts = true;
    try {
      const configResponse = await fetch('/api/config');
      if (!configResponse.ok) {
        throw new Error('Please configure the application first in the Configuration page');
      }
      config = await configResponse.json();
      selectedLanguage = config.default_language;

      const response = await fetch('/api/companies');
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to fetch contacts');
      }
      contacts = await response.json();
    } catch (err) {
      console.error('Error fetching contacts:', err);
      error = err instanceof Error ? err.message : 'Failed to fetch contacts';
      contacts = [];
    } finally {
      fetchingContacts = false;
    }
  }

  async function generateOutreach(contact: Contact) {
    loading = true;
    currentContact = contact;
    contact.error = undefined;

    try {
      const response = await fetch('/api/generate-outreach', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          website: contact.website,
          prompt,
          recordId: contact.id,
          language: selectedLanguage,
          contactInfo: {
            name: contact.fullname,
            company: contact.company_name,
            segment: contact.business_segment
          }
        })
      });

      const result = await response.json();
      if (result.error) {
        contact.error = result.error;
        contacts = [...contacts];
      } else {
        contact.outreach_text = result.outreachText;
        contact.error = undefined;
        contacts = [...contacts];
      }
    } catch (error) {
      console.error('Error generating outreach:', error);
      contact.error = error instanceof Error ? error.message : 'Failed to generate outreach';
      contacts = [...contacts];
    } finally {
      loading = false;
      currentContact = null;
    }
  }

  async function generateAllOutreach() {
    if (!prompt) return;

    loading = true;
    try {
      for (const contact of contacts) {
        currentContact = contact;
        await generateOutreach(contact);
      }
    } catch (error) {
      console.error('Error generating all outreach:', error);
    } finally {
      loading = false;
      currentContact = null;
    }
  }
</script>

<div class="container mx-auto p-4">
  <h1 class="text-2xl font-bold mb-4">AI Outreach Generator</h1>

  <div class="mb-6">
    <button
      on:click={fetchContacts}
      class="px-6 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
      disabled={fetchingContacts}
    >
      {#if fetchingContacts}
        <span>Fetching Contacts...</span>
      {:else}
        <span>Fetch Contacts from Airtable</span>
      {/if}
    </button>
  </div>

  {#if error}
    <div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
      <p>{error}</p>
    </div>
  {/if}

  <div class="mb-4">
    <div class="flex justify-between items-center mb-4">
      <label class="block text-sm font-medium text-gray-700">Outreach Language:</label>
      <select
        bind:value={selectedLanguage}
        class="ml-2 rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
      >
        {#each languages as lang}
          <option value={lang.code}>{lang.name}</option>
        {/each}
      </select>
    </div>
    <label class="block mb-2">Service Description / Prompt Template:</label>
    <textarea
      bind:value={prompt}
      class="w-full h-32 p-2 border rounded"
      placeholder="Describe your services and outreach style..."
    />
    <div class="mt-2 flex justify-end">
      <button
        on:click={() => generateAllOutreach()}
        class="px-6 py-2 bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50"
        disabled={loading || !prompt}
      >
        {#if loading}
          Generating All...
        {:else}
          Generate All Outreach
        {/if}
      </button>
    </div>
  </div>

  <div class="space-y-4">
    {#each contacts as contact}
      <div class="border p-4 rounded">
        <div class="grid grid-cols-2 gap-4 mb-3">
          <div>
            <h2 class="font-bold text-lg">{contact.company_name}</h2>
            <p class="text-sm text-gray-600">Contact: {contact.fullname}</p>
            <p class="text-sm text-gray-600">Segment: {contact.business_segment}</p>
          </div>
          <div>
            <p class="text-sm">
              <strong>Email:</strong> {contact.email}
            </p>
            <p class="text-sm">
              <strong>Phone:</strong> {contact.phone}
            </p>
            <p class="text-sm">
              <strong>Location:</strong> {contact.city}, {contact.country}
            </p>
          </div>
        </div>

        <p class="text-sm text-gray-600 mb-2">
          Website: <a href={contact.website} target="_blank" rel="noopener noreferrer" class="text-blue-500 hover:underline">{contact.website}</a>
        </p>

        {#if contact.error}
          <div class="mt-2 p-3 bg-red-50 rounded border border-red-300 shadow-sm">
            <h3 class="font-semibold mb-2 text-red-700">Error:</h3>
            <p class="text-sm text-red-600 font-medium">{contact.error}</p>
            <button
              on:click={() => generateOutreach(contact)}
              class="mt-2 text-sm text-red-700 hover:text-red-800 underline"
            >
              Try again
            </button>
          </div>
        {/if}

        {#if contact.outreach_text}
          <div class="mt-2 p-3 bg-gray-50 rounded border">
            <h3 class="font-semibold mb-2">Generated Outreach:</h3>
            <p class="text-sm whitespace-pre-wrap">{contact.outreach_text}</p>
          </div>
        {/if}

        <button
          on:click={() => generateOutreach(contact)}
          class="mt-3 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
          disabled={loading && currentContact?.id === contact.id}
        >
          {#if loading && currentContact?.id === contact.id}
            Generating...
          {:else}
            Generate Outreach
          {/if}
        </button>
      </div>
    {/each}
  </div>
</div>

```

# src/routes/api/companies/+server.ts

```ts
import Airtable from "airtable";
import { getRequiredConfig } from "$lib/db";
import { json } from "@sveltejs/kit";

export async function GET() {
  try {
    const config = getRequiredConfig();
    const base = new Airtable({ apiKey: config.airtable_access_token }).base(
      config.airtable_base_id
    );

    const records = await base("Imported table")
      .select({
        view: "Grid view",
        fields: [
          "fullname",
          "company name",
          "business segment",
          "website",
          "phone",
          "city",
          "country",
          "email",
          "outreach_text",
        ],
      })
      .firstPage();

    const contacts: Contact[] = records.map((record) => ({
      id: record.id,
      fullname: record.get("fullname") as string,
      company_name: record.get("company name") as string,
      business_segment: record.get("business segment") as string,
      website: record.get("website") as string,
      phone: record.get("phone") as string,
      city: record.get("city") as string,
      country: record.get("country") as string,
      email: record.get("email") as string,
      outreach_text: record.get("outreach_text") as string,
    }));

    return json(contacts);
  } catch (error) {
    console.error("Error fetching contacts:", error);
    let errorMessage = "Failed to fetch contacts";
    let statusCode = 500;

    if (error instanceof Error) {
      if (error.message.includes("NOT_FOUND")) {
        errorMessage =
          "Airtable base or table not found. Please verify your Base ID and table name.";
        statusCode = 404;
      } else if (
        error.message.includes("UNAUTHORIZED") ||
        error.message.includes("NOT_AUTHORIZED")
      ) {
        errorMessage =
          "Not authorized to access Airtable. Please check your access token and permissions.";
        statusCode = 401;
      }
    }

    return json({ error: errorMessage }, { status: statusCode });
  }
}
```

# src/routes/api/config/+server.ts

```ts
import { loadConfig, saveConfig } from "$lib/db";

import { json } from "@sveltejs/kit";

export async function GET() {
  try {
    const config = loadConfig();
    return json(config);
  } catch (error) {
    console.error("Error loading configuration:", error);
    return json(
      { error: "Failed to load configuration. Please check server logs." },
      { status: 500 }
    );
  }
}

export async function POST({ request }) {
  try {
    const config = await request.json();
    const success = saveConfig(config);

    if (!success) {
      throw new Error("Failed to save configuration");
    }

    return json({ success: true });
  } catch (error) {
    console.error("Error saving configuration:", error);
    return json(
      { error: "Failed to save configuration. Please check server logs." },
      { status: 500 }
    );
  }
}
```

# src/routes/api/generate-outreach/+server.ts

```ts
import * as cheerio from "cheerio";

import Airtable from "airtable";
import Anthropic from "@anthropic-ai/sdk";
import fetch from "node-fetch";
import { getRequiredConfig } from "$lib/db";
import { json } from "@sveltejs/kit";

const config = getRequiredConfig();
const anthropic = new Anthropic({
  apiKey: config.anthropic_api_key,
});

const base = new Airtable({ apiKey: config.airtable_access_token }).base(
  config.airtable_base_id
);

export async function POST({ request }) {
  try {
    const { website, prompt, recordId, contactInfo, language } =
      await request.json();

    // Ensure website URL has protocol
    const websiteUrl = website.startsWith("http")
      ? website
      : `https://${website}`;

    let websiteContent = "";
    try {
      // Fetch and parse website content
      const response = await fetch(websiteUrl);
      const html = await response.text();
      const $ = cheerio.load(html);

      // Extract relevant text content
      websiteContent = $("body").text().trim().slice(0, 2000);
    } catch (fetchError) {
      console.error(
        `Failed to fetch website content for ${websiteUrl}:`,
        fetchError
      );
      throw new Error(
        `Unable to access website ${websiteUrl}. Please verify the URL is correct.`
      );
    }

    // Generate outreach text using Claude
    const completion = await anthropic.messages.create({
      model: "claude-3-sonnet-20240229",
      max_tokens: 1000,
      messages: [
        {
          role: "user",
          content: `You are a professional outreach specialist. Generate the outreach email in ${language}. Based on this website content about ${
            contactInfo.company
          }:
        
        ${websiteContent}
        
        Contact Information:
        - Name: ${contactInfo.name}
        - Company: ${contactInfo.company}
        - Business Segment: ${contactInfo.segment}
        
        And considering our services/approach:
        ${prompt}
        
        ${
          websiteContent.includes("[Unable to fetch")
            ? "Note: Website content was not available, please focus on the business segment and company name to generate relevant outreach."
            : ""
        }
        
        Generate a personalized, compelling outreach email in ${language} that demonstrates understanding of their business and clearly articulates the value proposition. Keep it concise and professional. Address the person by their name.
        
        Important formatting rules:
        1. Start with the subject line on the first line
        2. Add a blank line after the subject
        3. Then write the email body
        4. Use the sender's actual name and company from the contact info instead of placeholders
        5. Do not include any explanatory text or metadata - just the email subject and body
        6. Do not include labels like "Subject:" or "Body:" - just write the content directly`,
        },
      ],
    });

    const outreachText = completion.content[0].text;

    // Save to Airtable
    await base("Imported table").update([
      {
        id: recordId,
        fields: {
          outreach_text: outreachText,
        },
      },
    ]);

    return json({ outreachText });
  } catch (error) {
    console.error("Error generating outreach:", error);
    let errorMessage = "Failed to generate outreach";

    if (error instanceof TypeError && error.code === "ERR_INVALID_URL") {
      errorMessage =
        "Invalid website URL. Please make sure the URL is correct.";
    } else if (error instanceof Error) {
      errorMessage = error.message;
    }

    return json({ error: errorMessage }, { status: 500 });
  }
}
```

# src/routes/config/+page.svelte

```svelte
<script lang="ts">
  import { onMount } from 'svelte';

  interface AirtableField {
    name: string;
    airtable_name: string;
    type: 'text' | 'email' | 'url' | 'textarea';
    required: boolean;
  }

  interface Config {
    anthropic_api_key: string;
    airtable_access_token: string;
    airtable_base_id: string;
    airtable_table_name: string;
    default_language: string;
    airtable_fields: AirtableField[];
  }

  let config: Config = {
    anthropic_api_key: '',
    airtable_access_token: '',
    airtable_base_id: '',
    airtable_table_name: '',
    default_language: 'en',
    airtable_fields: [
      { name: 'Full Name', airtable_name: 'fullname', type: 'text', required: true },
      { name: 'Company Name', airtable_name: 'company name', type: 'text', required: true },
      { name: 'Business Segment', airtable_name: 'business segment', type: 'text', required: true },
      { name: 'Website', airtable_name: 'website', type: 'url', required: true },
      { name: 'Email', airtable_name: 'email', type: 'email', required: true },
      { name: 'Phone', airtable_name: 'phone', type: 'text', required: false },
      { name: 'City', airtable_name: 'city', type: 'text', required: false },
      { name: 'Country', airtable_name: 'country', type: 'text', required: false },
      { name: 'Outreach Text', airtable_name: 'outreach_text', type: 'textarea', required: false }
    ]
  };

  let loading = false;
  let error: string | null = null;
  let success: string | null = null;

  async function saveConfig() {
    loading = true;
    error = null;
    success = null;

    try {
      const response = await fetch('/api/config', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(config)
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to save configuration');
      }

      success = 'Configuration saved successfully!';
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to save configuration';
    } finally {
      loading = false;
    }
  }

  async function loadConfig() {
    loading = true;
    error = null;

    try {
      const response = await fetch('/api/config');
      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to load configuration');
      }
      const loadedConfig = await response.json();

      if (loadedConfig.airtable_fields.length === 0) {
        loadedConfig.airtable_fields = config.airtable_fields;
      }

      config = loadedConfig;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load configuration';
    } finally {
      loading = false;
    }
  }

  onMount(loadConfig);

  function addField() {
    config.airtable_fields = [
      ...config.airtable_fields,
      { name: '', airtable_name: '', type: 'text', required: false }
    ];
  }

  function removeField(index: number) {
    config.airtable_fields = config.airtable_fields.filter((_, i) => i !== index);
  }

  const languages = [
    { code: 'en', name: 'English' },
    { code: 'pl', name: 'Polish' },
    { code: 'de', name: 'German' },
    { code: 'es', name: 'Spanish' },
    { code: 'fr', name: 'French' }
  ];
</script>

<div class="container mx-auto p-4">
  <h1 class="text-2xl font-bold mb-6">Configuration</h1>

  {#if error}
    <div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
      <p>{error}</p>
    </div>
  {/if}

  {#if success}
    <div class="mb-4 p-4 bg-green-100 border border-green-400 text-green-700 rounded">
      <p>{success}</p>
    </div>
  {/if}

  <div class="space-y-6">
    <div class="bg-white p-6 rounded-lg shadow">
      <h2 class="text-xl font-semibold mb-4">API Configuration</h2>
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700">Default Language for Outreach</label>
          <select
            bind:value={config.default_language}
            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          >
            {#each languages as lang}
              <option value={lang.code}>{lang.name}</option>
            {/each}
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700">Anthropic API Key</label>
          <input
            type="password"
            bind:value={config.anthropic_api_key}
            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700">Airtable Access Token</label>
          <input
            type="password"
            bind:value={config.airtable_access_token}
            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700">Airtable Base ID</label>
          <input
            type="text"
            bind:value={config.airtable_base_id}
            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700">Airtable Table Name</label>
          <input
            type="text"
            bind:value={config.airtable_table_name}
            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          />
        </div>
      </div>
    </div>

    <div class="bg-white p-6 rounded-lg shadow">
      <div class="flex justify-between items-center mb-4">
        <h2 class="text-xl font-semibold">Airtable Fields Configuration</h2>
        <button
          on:click={addField}
          class="px-3 py-1 bg-indigo-600 text-white rounded hover:bg-indigo-700"
        >
          Add Field
        </button>
      </div>

      <div class="space-y-4">
        {#each config.airtable_fields as field, index}
          <div class="flex gap-4 items-start p-4 border rounded">
            <div class="flex-1">
              <label class="block text-sm font-medium text-gray-700">Display Name</label>
              <input
                type="text"
                bind:value={field.name}
                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
            </div>
            <div class="flex-1">
              <label class="block text-sm font-medium text-gray-700">Airtable Field Name</label>
              <input
                type="text"
                bind:value={field.airtable_name}
                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
            </div>
            <div class="w-32">
              <label class="block text-sm font-medium text-gray-700">Type</label>
              <select
                bind:value={field.type}
                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              >
                <option value="text">Text</option>
                <option value="email">Email</option>
                <option value="url">URL</option>
                <option value="textarea">Textarea</option>
              </select>
            </div>
            <div class="w-24 pt-6">
              <label class="inline-flex items-center">
                <input
                  type="checkbox"
                  bind:checked={field.required}
                  class="rounded border-gray-300 text-indigo-600 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                />
                <span class="ml-2 text-sm text-gray-600">Required</span>
              </label>
            </div>
            <button
              on:click={() => removeField(index)}
              class="mt-6 text-red-600 hover:text-red-800"
            >
              Remove
            </button>
          </div>
        {/each}
      </div>
    </div>

    <div class="flex justify-end gap-4">
      <button
        on:click={loadConfig}
        class="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 disabled:opacity-50"
        disabled={loading}
      >
        Load Saved Config
      </button>
      <button
        on:click={saveConfig}
        class="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
        disabled={loading}
      >
        Save Configuration
      </button>
    </div>
  </div>
</div>
```

# static/favicon.png

This is a binary file of the type: Image

# svelte.config.js

```js
import adapter from "@sveltejs/adapter-auto";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://svelte.dev/docs/kit/integrations
  // for more information about preprocessors
  preprocess: vitePreprocess(),

  kit: {
    // adapter-auto only supports some environments, see https://svelte.dev/docs/kit/adapter-auto for a list.
    // If your environment is not supported, or you settled on a specific environment, switch out the adapter.
    // See https://svelte.dev/docs/kit/adapters for more information about adapters.
    adapter: adapter(),
  },
};

export default config;
```

# tailwind.config.ts

```ts
import forms from "@tailwindcss/forms";
import typography from "@tailwindcss/typography";
import type { Config } from "tailwindcss";

export default {
  content: ["./src/**/*.{html,js,svelte,ts}"],

  theme: {
    extend: {},
  },

  plugins: [typography, forms],
} satisfies Config;
```

# tsconfig.json

```json
{
  "extends": "./.svelte-kit/tsconfig.json",
  "compilerOptions": {
    "allowJs": true,
    "checkJs": true,
    "esModuleInterop": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "skipLibCheck": true,
    "sourceMap": true,
    "strict": true,
    "moduleResolution": "bundler"
  }
  // Path aliases are handled by https://svelte.dev/docs/kit/configuration#alias
  // except $lib which is handled by https://svelte.dev/docs/kit/configuration#files
  //
  // If you want to overwrite includes/excludes, make sure to copy over the relevant includes/excludes
  // from the referenced tsconfig.json - TypeScript does not merge them in
}
```

# vite.config.ts

```ts
import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [sveltekit()],
});
```
