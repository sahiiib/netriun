function applyTheme(themeName) {
    const theme = document.getElementById("theme-link");
    const version = theme?.dataset.version;
    const suffix = version ? `?v=${encodeURIComponent(version)}` : "";
    const href = themeName === "light" ? `/static/css/light.css${suffix}` : `/static/css/dark.css${suffix}`;

    theme.setAttribute("href", href);
    localStorage.setItem("netriun-theme", themeName);
}

function toggleTheme() {
    const theme = document.getElementById("theme-link");
    const nextTheme = theme.getAttribute("href").includes("dark") ? "light" : "dark";
    applyTheme(nextTheme);
}

document.addEventListener("DOMContentLoaded", () => {
    applyTheme(localStorage.getItem("netriun-theme") || "dark");
});
