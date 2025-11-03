// Share button functionality
(function () {
  // Get current page URL and title
  const url = encodeURIComponent(window.location.href);
  const title = encodeURIComponent(document.title);

  // Twitter share
  window.shareOnTwitter = function () {
    const twitterUrl = `https://twitter.com/intent/tweet?url=${url}&text=${title}`;
    window.open(twitterUrl, "twitter-share", "width=550,height=420");
  };

  // LinkedIn share
  window.shareOnLinkedIn = function () {
    const linkedInUrl = `https://www.linkedin.com/sharing/share-offsite/?url=${url}`;
    window.open(linkedInUrl, "linkedin-share", "width=550,height=420");
  };

  // Facebook share
  window.shareOnFacebook = function () {
    const facebookUrl = `https://www.facebook.com/sharer/sharer.php?u=${url}`;
    window.open(facebookUrl, "facebook-share", "width=550,height=420");
  };

  // Reddit share
  window.shareOnReddit = function () {
    const redditUrl = `https://reddit.com/submit?url=${url}&title=${title}`;
    window.open(redditUrl, "reddit-share", "width=550,height=420");
  };

  // Email share
  window.shareViaEmail = function () {
    const subject = decodeURIComponent(title);
    const body = `I thought you might find this interesting: ${decodeURIComponent(
      url
    )}`;
    window.location.href = `mailto:?subject=${encodeURIComponent(
      subject
    )}&body=${encodeURIComponent(body)}`;
  };

  // Copy link to clipboard
  window.copyLink = function () {
    const tempInput = document.createElement("input");
    tempInput.value = window.location.href;
    document.body.appendChild(tempInput);
    tempInput.select();
    document.execCommand("copy");
    document.body.removeChild(tempInput);

    // Show feedback
    const copyBtn = document.querySelector(".share-copy");
    if (copyBtn) {
      const originalText = copyBtn.innerHTML;
      copyBtn.innerHTML = "✓ Copied!";
      copyBtn.classList.add("copied");
      setTimeout(() => {
        copyBtn.innerHTML = originalText;
        copyBtn.classList.remove("copied");
      }, 2000);
    }
  };
})();
