const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const root = path.join(__dirname, '../static');
const css = fs.readFileSync(path.join(root, 'css/style.css'), 'utf8');
const app = fs.readFileSync(path.join(root, 'js/app.js'), 'utf8');
const index = fs.readFileSync(path.join(root, 'index.html'), 'utf8');
const start = app.indexOf('function closeAllCustomSelects()');
const end = app.indexOf("document.addEventListener('click',", app.indexOf('function enhanceSelects(', start));
assert.ok(start >= 0 && end > start, 'shared dropdown implementation is present');
const dropdowns = app.slice(start, end);

const fixture = `
<main style="max-width:640px;margin:auto;padding:24px">
  <h1>Dropdown alignment fixture</h1>
  <label class="f" id="auth-field"><span>Authentication</span><select id="auth"><option value="password">Password</option><option value="key">Private key</option></select></label>
  <section class="np-server-settings-group"><label class="f"><span>Runtime</span><select id="runtime"><option>Java 21</option><option>Java 25</option></select></label></section>
  <section class="sfi-modal" style="width:100%;padding:0"><label class="f"><span>Import authentication</span><select id="import-auth"><option>Password</option><option>Private key</option></select></label></section>
  <label class="f"><span>Long option</span><select id="long"><option>Automatic (Linux: bwrap + systemd required; Windows: Job Object, no filesystem jail)</option></select></label>
  <label class="f"><span>Unavailable</span><select id="disabled" disabled><option>Unavailable</option></select></label>
  <select id="standalone"><option>Password</option><option>Private key</option></select>
</main>`;

test('field-caption styling does not leak into custom dropdown spans', () => {
  assert.match(css, /label\.f\s*>\s*span\s*\{/);
  assert.doesNotMatch(css, /label\.f\s+span\s*\{/);
});

test('browser: dropdown values and chevrons stay centered in labeled fields', { skip: !process.env.NP_LAYOUT_PLAYWRIGHT }, async () => {
  const { chromium } = require(process.env.NP_LAYOUT_PLAYWRIGHT);
  const browser = await chromium.launch({ headless: true, ...(process.env.NP_LAYOUT_CHROME ? { executablePath: process.env.NP_LAYOUT_CHROME } : {}) });
  try {
    const page = await browser.newPage();
    const sheets = [...index.matchAll(/href="\/css\/([^"]+)"/g)].map(match => match[1]);
    const styles = sheets.map(file => fs.readFileSync(path.join(root, 'css', file), 'utf8')).join('\n');
    await page.emulateMedia({ reducedMotion: 'reduce' });
    for (const theme of ['dark', 'light']) {
      for (const width of [1440, 390, 320]) {
        await page.setViewportSize({ width, height: 900 });
        await page.setContent(`<html data-theme="${theme}"><head><style>${styles}</style></head><body>${fixture}</body></html>`);
        await page.addScriptTag({ content: dropdowns + '\nenhanceSelects(document);' });
        const measurements = await page.locator('.cselect-trigger').evaluateAll(buttons => buttons.map(button => {
          const value = button.querySelector('.cselect-val');
          const arrow = button.querySelector('.cselect-arrow');
          const svg = arrow.querySelector('svg');
          const center = element => { const box = element.getBoundingClientRect(); return box.top + box.height / 2; };
          const style = getComputedStyle(button), valueStyle = getComputedStyle(value);
          return {
            id: button.parentElement.dataset.for,
            valueOffset: center(value) - center(button),
            arrowOffset: center(svg) - center(button),
            valueMargin: valueStyle.marginBottom,
            arrowMargin: getComputedStyle(arrow).marginBottom,
            arrowDisplay: getComputedStyle(arrow).display,
            font: valueStyle.fontSize, buttonFont: style.fontSize,
            color: valueStyle.color, buttonColor: style.color,
            textAlign: style.textAlign,
            overflow: button.scrollWidth > button.clientWidth + 1,
          };
        }));
        for (const item of measurements) {
          const context = `${theme}/${width}/${item.id}: ${JSON.stringify(item)}`;
          assert.ok(Math.abs(item.valueOffset) <= 0.6, `selected text is off-center: ${context}`);
          assert.ok(Math.abs(item.arrowOffset) <= 0.6, `chevron is off-center: ${context}`);
          assert.equal(item.valueMargin, '0px', context);
          assert.equal(item.arrowMargin, '0px', context);
          assert.equal(item.arrowDisplay, 'flex', context);
          assert.equal(item.font, item.buttonFont, context);
          assert.equal(item.color, item.buttonColor, context);
          assert.equal(item.textAlign, 'left', context);
          assert.equal(item.overflow, false, context);
        }
        const caption = await page.locator('#auth-field > span').evaluate(element => ({ margin: getComputedStyle(element).marginBottom, size: getComputedStyle(element).fontSize }));
        assert.equal(caption.margin, '5px', 'caption spacing remains unchanged');
        assert.equal(caption.size, '12.5px', 'caption typography remains unchanged');
        assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), 'no horizontal overflow');

        const trigger = page.locator('[data-for="auth"] .cselect-trigger');
        await trigger.focus();
        await page.keyboard.press('Enter');
        await page.locator('.cselect-menu.cselect-portaled .cselect-opt').filter({ hasText: 'Private key' }).click();
        assert.equal(await page.locator('#auth').inputValue(), 'key');
        assert.equal(await trigger.locator('.cselect-val').textContent(), 'Private key');
        assert.equal(await trigger.getAttribute('aria-expanded'), 'false');
        assert.equal(await page.locator('[data-for="disabled"] .cselect-trigger').isDisabled(), true);
      }
    }
  } finally {
    await browser.close();
  }
});
