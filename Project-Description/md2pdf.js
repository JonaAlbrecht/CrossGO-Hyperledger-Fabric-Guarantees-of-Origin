const fs = require('fs');
const path = require('path');
const { marked } = require('marked');
const puppeteer = require('puppeteer-core');

const CHROME_PATH = 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';

async function main() {
  const inputFile = process.argv[2] || 'architecture-diagrams.md';
  const outputFile = inputFile.replace(/\.md$/, '.pdf');
  const markdown = fs.readFileSync(path.resolve(__dirname, inputFile), 'utf-8');

  // Convert markdown → HTML
  let html = marked.parse(markdown);

  // marked encodes mermaid blocks as <pre><code class="language-mermaid">…</code></pre>
  // Convert them to <div class="mermaid">…</div> so mermaid.js picks them up
  html = html.replace(
    /<pre><code class="language-mermaid">([\s\S]*?)<\/code><\/pre>/g,
    (_match, code) => {
      const decoded = code
        .replace(/&amp;/g, '&')
        .replace(/&lt;/g, '<')
        .replace(/&gt;/g, '>')
        .replace(/&quot;/g, '"')
        .replace(/&#39;/g, "'");
      return `<div class="mermaid">${decoded}</div>`;
    }
  );

  const fullHtml = `<!DOCTYPE html>
<html lang="en"><head>
<meta charset="utf-8">
<title>${path.basename(inputFile, '.md')}</title>
<style>
  @page { size: A4; margin: 20mm 15mm; }
  body {
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
    font-size: 11pt;
    line-height: 1.55;
    color: #1a1a1a;
    max-width: 100%;
  }
  h1 { font-size: 22pt; border-bottom: 2.5px solid #222; padding-bottom: 8px; margin-top: 0; }
  h2 { font-size: 16pt; border-bottom: 1px solid #bbb; padding-bottom: 5px; margin-top: 32px; page-break-after: avoid; }
  h3 { font-size: 13pt; margin-top: 24px; page-break-after: avoid; }
  p, li { orphans: 3; widows: 3; }
  blockquote {
    border-left: 4px solid #999; margin: 12px 0; padding: 8px 16px;
    background: #f7f7f7; color: #444; font-size: 10pt;
  }
  pre {
    background: #f4f4f4; padding: 12px; border-radius: 4px;
    font-size: 9pt; overflow-x: auto; white-space: pre-wrap;
    page-break-inside: avoid;
  }
  code { background: #eee; padding: 1px 4px; border-radius: 3px; font-size: 9.5pt; }
  pre code { background: none; padding: 0; }
  table { border-collapse: collapse; width: 100%; margin: 14px 0; page-break-inside: avoid; }
  th, td { border: 1px solid #ccc; padding: 6px 10px; font-size: 10pt; }
  th { background: #f0f0f0; font-weight: 600; }
  hr { border: none; border-top: 1px solid #ddd; margin: 24px 0; }

  /* Mermaid diagrams */
  .mermaid {
    text-align: center;
    margin: 18px 0;
    page-break-inside: avoid;
  }
  .mermaid svg {
    max-width: 100%;
    height: auto;
  }
</style>
<script src="https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js"></script>
</head>
<body>
${html}
<script>
  mermaid.initialize({
    startOnLoad: true,
    theme: 'default',
    securityLevel: 'loose',
    flowchart: { useMaxWidth: true, htmlLabels: true },
    sequence: { useMaxWidth: true }
  });
</script>
</body></html>`;

  console.log('Launching Chrome...');
  const browser = await puppeteer.launch({
    executablePath: CHROME_PATH,
    headless: 'new',
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });
  const page = await browser.newPage();
  await page.setContent(fullHtml, { waitUntil: 'networkidle0', timeout: 60000 });

  // Wait for mermaid to finish rendering all diagrams
  console.log('Waiting for Mermaid diagrams to render...');
  await page.waitForFunction(
    () => {
      const pending = document.querySelectorAll('.mermaid:not([data-processed="true"])');
      return pending.length === 0;
    },
    { timeout: 30000 }
  );
  // Small extra delay for SVG layout finalization
  await new Promise(r => setTimeout(r, 1500));

  console.log('Generating PDF...');
  await page.pdf({
    path: path.resolve(__dirname, outputFile),
    format: 'A4',
    margin: { top: '20mm', right: '15mm', bottom: '20mm', left: '15mm' },
    printBackground: true,
    displayHeaderFooter: false,
    preferCSSPageSize: false
  });

  await browser.close();
  console.log(`Done → ${outputFile}`);
}

main().catch(err => { console.error(err); process.exit(1); });
