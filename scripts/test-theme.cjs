const {readFileSync} = require('node:fs');
const {runInNewContext} = require('node:vm');
const assert = require('node:assert/strict');
const source = readFileSync('web/static/js/theme.js', 'utf8');
function setup(saved, blocked = false) {
    const events = {}, root = {dataset: {}, style: {}}, sheet = {}, select = {};
    const media = {matches: false, addEventListener: (_, callback) => events.system = callback};
    const store = new Map(saved ? [['netriun-theme-preference', saved]] : []);
    runInNewContext(source, {
        window: {matchMedia: () => media, addEventListener: (name, cb) => events[name] = cb},
        document: {documentElement: root, getElementById: id => id === 'theme-link' ? sheet : select,
            addEventListener: (name, cb) => events[name] = cb},
        localStorage: {getItem: key => {if (blocked) throw Error(); return store.get(key);},
            setItem: (key, value) => {if (blocked) throw Error(); store.set(key, value);}}
    });
    select.addEventListener = (_, cb) => events.select = cb;
    events.DOMContentLoaded();
    return {events, root, sheet, select, media, store};
}
const app = setup();
assert.equal(app.select.value, 'system');
assert.equal(app.root.dataset.theme, 'light');
app.media.matches = true; app.events.system();
assert.equal(app.root.dataset.theme, 'dark');
app.events.select({target: {value: 'light'}});
assert.equal(app.sheet.media, 'all');
app.events.system();
assert.equal(app.root.dataset.theme, 'light');
assert.equal(setup(app.store.get('netriun-theme-preference')).select.value, 'light');
app.events.select({target: {value: 'system'}});
assert.equal(app.root.dataset.theme, 'dark');
app.events.storage({key: 'netriun-theme-preference', newValue: 'light'});
assert.equal(app.select.value, 'light');
assert.equal(setup('invalid').select.value, 'system');
const blocked = setup(undefined, true);
blocked.events.select({target: {value: 'dark'}});
assert.equal(blocked.root.dataset.theme, 'dark');
console.log('Theme: system changes, overrides, persistence, cross-tab sync, invalid and blocked storage passed.');
