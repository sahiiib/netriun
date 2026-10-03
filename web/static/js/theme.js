(() => {
    const key = "netriun-theme-preference";
    const allowed = ["system", "light", "dark"];
    const system = window.matchMedia("(prefers-color-scheme: dark)");
    const stylesheet = document.getElementById("theme-link");
    let preference = "system";
    try {
        const saved = localStorage.getItem(key);
        if (allowed.includes(saved)) preference = saved;
    } catch (_) { /* Theme selection also works without browser storage. */ }

    function render() {
        const theme = preference === "system" ? (system.matches ? "dark" : "light") : preference;
        stylesheet.media = theme === "light" ? "all" : "not all";
        document.documentElement.dataset.theme = theme;
        document.documentElement.style.colorScheme = theme;
        const select = document.getElementById("theme-select");
        if (select) select.value = preference;
    }
    render();
    system.addEventListener("change", render);
    window.addEventListener("storage", (event) => {
        if (event.key === key || event.key === null) {
            preference = allowed.includes(event.newValue) ? event.newValue : "system";
            render();
        }
    });
    document.addEventListener("DOMContentLoaded", () => {
        render();
        document.getElementById("theme-select")?.addEventListener("change", (event) => {
            preference = allowed.includes(event.target.value) ? event.target.value : "system";
            try { localStorage.setItem(key, preference); } catch (_) {}
            render();
        });
    });
})();
