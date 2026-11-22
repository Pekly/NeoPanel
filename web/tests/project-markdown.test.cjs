const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const vm = require('node:vm');
const root = path.join(__dirname, '../static');
const assets = ['vendor/marked-18.2.1/marked.umd.js','vendor/dompurify-3.4.18/purify.min.js','js/project-markdown.js'];
const read = p => fs.readFileSync(path.join(root,p),'utf8');

test('Markdown dependencies are local, pinned, licensed and checksum-verified',()=>{
  const index = read('index.html');
  let prior = -1;
  for(const asset of [...assets,'js/plugin-details.js']) {const pos=index.indexOf(`src="/${asset}"`);assert.ok(pos>prior);prior=pos;}
  for(const pkg of ['marked-18.2.1','dompurify-3.4.18']) {
    const base=path.join(root,'vendor',pkg), info=JSON.parse(fs.readFileSync(path.join(base,'SOURCES.json')));
    assert.match(info.integrity,/^sha512-/);
    for(const file of info.files) assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.join(base,file.file))).digest('hex'),file.sha256);
    assert.ok(fs.statSync(path.join(base,'LICENSE')).size>1000);
  }
});

test('renderer fails closed without parser or sanitizer',()=>{
  const c={window:{},URL};vm.runInNewContext(read('js/project-markdown.js'),c);
  for(const partial of [{},{marked:{parse:()=>'<img onerror=alert(1)>'}},{DOMPurify:{sanitize:x=>x,isSupported:true}}]) {
    Object.assign(c.window,{marked:undefined,DOMPurify:undefined},partial);
    const result=c.window.NPProjectMarkdown.render('<script>alert(1)</script> **bold**');
    assert.match(result,/Formatted preview unavailable/);assert.match(result,/&lt;script&gt;/);assert.doesNotMatch(result,/<script|<img/);
  }
});

test('Chrome: standard Markdown, safe author HTML, hostile input and narrow layouts',{skip:!process.env.NP_LAYOUT_PLAYWRIGHT},async()=>{
  const {chromium}=require(process.env.NP_LAYOUT_PLAYWRIGHT);
  const browser=await chromium.launch({headless:true,executablePath:process.env.NP_LAYOUT_CHROME});
  try {
    const page=await browser.newPage({viewport:{width:390,height:900}});
    await page.route('**/*',route=>route.abort()); // No provider/font requests in this deterministic rendering test.
    await page.setContent('<html><head><style>:root{--border:#555;--card-inner:#222}body{margin:16px}.np-project-markdown{max-width:100%}</style></head><body><article class="np-project-markdown"></article></body></html>');
    await page.addStyleTag({content:read('css/plugin-details.css')});
    for(const asset of assets) await page.addScriptTag({content:read(asset)});
    const markdown = `# Simple Voice Chat-style description

[![Badge](https://example.com/badge.svg)](https://example.com/docs)

## Features

A **bold _nested emphasis_** paragraph with ~~removed text~~ and [reference docs][docs].

1. First item
   - Nested child
2. Second item

- [x] Ready
- [ ] Optional

| Platform | Support |
| --- | --- |
| Paper | **Yes** |

> A quoted paragraph

\`\`\`yaml
voice_host: "example.com"
port: 24454
\`\`\`

<div align="center"><a href="https://example.com"><img src="https://example.com/banner.png" width="160" alt="Project banner"></a></div>

<details><summary>Setup instructions</summary><p>Configure your voice port.</p></details>

[docs]: https://example.com/reference "Documentation"
`;
    await page.evaluate(text=>document.querySelector('article').innerHTML=NPProjectMarkdown.render(text),markdown);
    assert.equal(await page.locator('article p').first().evaluate(el=>getComputedStyle(el).whiteSpace),'normal');
    assert.equal(await page.locator('h1').textContent(),'Simple Voice Chat-style description');
    assert.equal(await page.locator('strong em').textContent(),'nested emphasis');
    assert.equal(await page.locator('ol > li').count(),2);assert.equal(await page.locator('ol ul li').count(),1);
    assert.equal(await page.locator('.np-markdown-table table tbody tr').count(),1);
    assert.equal(await page.locator('input:enabled').count(),0);assert.equal(await page.locator('input[type=checkbox]').count(),2);
    assert.match(await page.locator('pre code').textContent(),/port: 24454/);
    assert.equal(await page.locator('a[href="https://example.com/reference"]').getAttribute('rel'),'noopener noreferrer');
    assert.equal(await page.locator('img[alt="Project banner"]').getAttribute('referrerpolicy'),'no-referrer');
    await page.locator('summary').click();assert.equal(await page.locator('details').getAttribute('open'),'');
    assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1));

    const attacks = [
      '<script>window.pwned=1</script><img src="https://example.com/x" onerror="window.pwned=1">',
      '[bad](javascript:alert(1)) <a href="jav&#x61;script:alert(1)">bad</a> <a href="data:text/html,x">bad</a>',
      '<svg><a xlink:href="javascript:alert(1)"><text>x</text></a></svg><math><mtext><img src=x onerror="window.pwned=1"></mtext></math>',
      '<iframe srcdoc="<script>parent.pwned=1</script>"></iframe><object data="https://example.com"></object><embed src="x">',
      '<style>body{display:none}</style><div style="position:fixed;inset:0" class="np-project" id="main" name="NPPlugins">spoof</div>',
      '<form action="https://example.com"><input name="token"><button>Submit</button></form>',
      '<img src="data:image/svg+xml,<svg onload=alert(1)>" srcset="https://example.com/x 1x"><a href="https://user:pass@example.com">bad</a>',
      '<math><mtext><table><mglyph><style><!--</style><img title="--><img src=1 onerror=window.pwned=1>">',
      '<details ontoggle="window.pwned=1" open><summary>x</summary>safe</details>'
    ];
    for(const input of attacks) {
      await page.evaluate(text=>document.querySelector('article').innerHTML=NPProjectMarkdown.render(text),input);
      assert.equal(await page.locator('article script,article style,article svg,article math,article iframe,article object,article embed,article form,article button,article [onerror],article [ontoggle],article [id],article [name],article [style],article [srcset],article input:not([disabled])').count(),0,input);
      const urls=await page.locator('article [href],article [src]').evaluateAll(nodes=>nodes.flatMap(n=>['href','src'].filter(a=>n.hasAttribute(a)).map(a=>n.getAttribute(a))));
      for(const value of urls) assert.match(value,/^https?:\/\//);
      assert.equal(await page.evaluate(()=>window.pwned),undefined);
    }
    await page.evaluate(()=>document.querySelector('article').innerHTML=NPProjectMarkdown.render('x'.repeat(200000)));
    assert.match(await page.locator('article').textContent(),/Content shortened/);
    await browser.close();
  } finally {await browser.close();}
});
