// Theme toggle functionality
const themeToggle = document.getElementById("theme-toggle");
const htmlElement = document.documentElement;

// Check for saved theme preference or default to 'light'
const currentTheme = localStorage.getItem("theme") || "light";
htmlElement.setAttribute("data-theme", currentTheme);
updateThemeButton(currentTheme);

themeToggle.addEventListener("click", function () {
  const currentTheme = htmlElement.getAttribute("data-theme");
  const newTheme = currentTheme === "dark" ? "light" : "dark";

  htmlElement.setAttribute("data-theme", newTheme);
  localStorage.setItem("theme", newTheme);
  updateThemeButton(newTheme);
});

function updateThemeButton(theme) {
  if (theme === "dark") {
    themeToggle.innerHTML = "☀️ Light";
  } else {
    themeToggle.innerHTML = "🌙 Dark";
  }
}
