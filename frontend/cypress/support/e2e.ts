// Cypress E2E Support File
// This file is loaded before every E2E test

// Import commands
import "./commands";

// Hide fetch/XHR requests from command log for cleaner output
const app = window.top;
if (
  app &&
  !app.document.head.querySelector("[data-hide-command-log-request]")
) {
  const style = app.document.createElement("style");
  style.setAttribute("data-hide-command-log-request", "");
  style.innerHTML =
    ".command-name-request, .command-name-xhr { display: none }";
  app.document.head.appendChild(style);
}
