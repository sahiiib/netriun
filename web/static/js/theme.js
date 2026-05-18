function applyTheme(themeName) {
    const theme = document.getElementById("theme-link");
    const href = themeName === "light" ? "/static/css/light.css" : "/static/css/dark.css";

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
