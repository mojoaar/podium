// Copy code button functionality
document.addEventListener("DOMContentLoaded", function () {
  // Find all pre elements with code inside
  const preElements = document.querySelectorAll("pre");

  preElements.forEach((pre) => {
    // Skip if already has a copy button
    if (pre.querySelector(".copy-code-btn")) {
      return;
    }

    const codeBlock = pre.querySelector("code");
    if (!codeBlock) {
      return;
    }

    // Create copy button
    const copyButton = document.createElement("button");
    copyButton.className = "copy-code-btn";
    copyButton.innerHTML = `
      <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
      </svg>
      <span>Copy</span>
    `;
    copyButton.setAttribute("aria-label", "Copy code to clipboard");

    // Add click handler
    copyButton.addEventListener("click", async () => {
      const code = codeBlock.textContent;

      try {
        await navigator.clipboard.writeText(code);

        // Visual feedback
        copyButton.innerHTML = `
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="20 6 9 17 4 12"></polyline>
          </svg>
          <span>Copied!</span>
        `;
        copyButton.classList.add("copied");

        // Reset after 2 seconds
        setTimeout(() => {
          copyButton.innerHTML = `
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
              <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
            </svg>
            <span>Copy</span>
          `;
          copyButton.classList.remove("copied");
        }, 2000);
      } catch (err) {
        console.error("Failed to copy code:", err);
        copyButton.textContent = "Failed!";
        setTimeout(() => {
          copyButton.innerHTML = `
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
              <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
            </svg>
            <span>Copy</span>
          `;
        }, 2000);
      }
    });

    // Add button directly to pre element
    pre.appendChild(copyButton);

    // Make sure pre has position relative
    pre.style.position = "relative";
  });
});
