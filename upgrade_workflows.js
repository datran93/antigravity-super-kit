const fs = require("fs");

function updateFile(filePath, replacements) {
  if (!fs.existsSync(filePath)) return;
  let content = fs.readFileSync(filePath, "utf8");
  for (const [search, replace] of replacements) {
    content = content.replace(search, replace);
  }
  fs.writeFileSync(filePath, content);
  console.log(`Updated: ${filePath}`);
}

const plannerReplacements = [
  [
    /- Read `ANCHORS\.md` to refresh immutable guardrails\./g,
    `- **Context Pruning**: Use \`manage_anchors\` (action: "list") or \`recall_knowledge\` to dynamically fetch only the domain-specific constraints relevant to the task (e.g., Auth, DB, UI) instead of loading the entire \`ANCHORS.md\` file.`,
  ],
];

const coderReplacements = [
  [
    /## Phase 2: Pattern Discovery 🔍\n\n> You are extending an existing codebase\. Your code MUST look like it belongs\.\n\n1\. \*\*Find reference\*\*: `search_code` for ≥ 1 existing file of the same type \(handler, service, repository\)\./g,
    `## Phase 2: AST Pattern & Dependency Discovery 🔍\n\n> You are extending an existing codebase. Your code MUST look like it belongs and you MUST know its blast radius.\n\n1. **360° Context**: Use \`context\` or \`search_symbol\` (via \`codebase-explorer\`) to get an AST-based view of the classes/functions you are modifying. This reveals definitions and related chunks without keyword-guessing.\n2. **Blast Radius**: Use \`find_usages\` to map out dependencies before modifying shared code.`,
  ],
  [/2\. \*\*Study\*\*: Error handling/g, `3. **Study**: Error handling`],
  [/3\. \*\*Document\*\*: State which pattern/g, `4. **Document**: State which pattern`],
  [/4\. \*\*Deviations\*\*: NEVER deviate/g, `5. **Deviations**: NEVER deviate`],
];

const buildReplacements = [
  [
    /## Phase 2: Pattern Discovery 🔍\n\n> You are extending an existing codebase\. Your code MUST look like it belongs\.\n\n1\. `search_code` for ≥ 1 existing file of the same type\.\n2\. Study: error handling, response format, naming conventions\.\n3\. State which pattern you're following before writing code\./g,
    `## Phase 2: AST Pattern & Dependency Discovery 🔍\n\n> You are extending an existing codebase. Your code MUST look like it belongs and avoid unintended side-effects.\n\n1. Use \`context\` or \`search_symbol\` (via \`codebase-explorer\`) for a 360° AST view of target files and dependencies.\n2. Use \`find_usages\` to understand the blast radius if modifying shared utilities.\n3. Study: error handling, response format, naming conventions.\n4. State which pattern you're following before writing code.`,
  ],
];

const fastFixReplacements = [
  [
    /2\. Identify target file\(s\) — read them before modifying\./g,
    `2. Identify target file(s). Use \`context\` or \`find_usages\` (via \`codebase-explorer\`) if modifying a shared function to quickly gauge its blast radius.\n3. Read the files before modifying.`,
  ],
  [/3\. Quick pattern scan/g, `4. Quick pattern scan`],
];

const targets = [
  { file: "planner-architect.md", replacements: plannerReplacements },
  { file: "coder-implementation.md", replacements: coderReplacements },
  { file: "build.md", replacements: buildReplacements },
  { file: "fast-fix.md", replacements: fastFixReplacements },
];

for (const target of targets) {
  updateFile(`**/commands/${target.file}`, target.replacements);
  updateFile(`**/workflows/${target.file}`, target.replacements);
}

console.log("Upgrade complete.");
